package main

import (
    "flag"
    "io/ioutil"
    "log"
    "os"
    "regexp"
    "strings"
    "time"


    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/system"
    "github.com/influxdata/toml"
)

type ValueStruct struct {
    Source  string
    Value   string
    Default string
    Regexp  string
    RE      *regexp.Regexp
}

type VariableStruct struct {
    Name   string
    Source string
    Value  string
    Regexp string
    RE     *regexp.Regexp
}

type FilterStruct struct {
    Source string
    Value  string
    Regexp string
    RE     *regexp.Regexp
}

type ResultStruct struct {
    Destination string
    Value       ValueStruct
    Filter      []FilterStruct
    Variable    []VariableStruct
}

type RuleStruct struct {
    Source      string
    Filter      []FilterStruct
    Result      []ResultStruct
}

type ConfigStruct struct {
    MQTT struct {
        Host     string
        User     string
        Password string
    }

    Rule []RuleStruct
}


const (
    defaultMQTThost = "localhost:1883"
)

var debug          bool
var mqttHost       string
var configFile     string
var variableRE     *regexp.Regexp

func init() {
    flag.BoolVar(&debug,         "debug",   false,             "debug mode")
    flag.StringVar(&mqttHost,    "mqtt",    defaultMQTThost,   "MQTT server, can set without port number")
    flag.StringVar(&configFile,  "config",  "clerk.conf",      "path to config file")

    variableRE = regexp.MustCompile("\\$([a-z0-9]+)")
}

func readConfig(filename string) (ConfigStruct, error) {
    var config ConfigStruct

    f, err := os.Open(filename)
    if err != nil {
        return config, err
    }
    defer f.Close()

    buf, err := ioutil.ReadAll(f)
    if err != nil {
        return config, err
    }

    if err := toml.Unmarshal(buf, &config); err != nil {
        return ConfigStruct{}, err
    }

    return config, err
}

func checkRuleCorrect(rule RuleStruct) bool {
    return rule.Source != "" && len(rule.Result) > 0
}

func getResultData(source, topic, message, destination, value string) string {
    if source == "topic" {
        return topic
    } else if source == "destination" {
        return destination
    } else if source == "value" {
        return value
    }

    return message
}

func getData(source, topic, message string) string {
    return getResultData(source, topic, message, "", "")
}

func applyVariables(value string, vars map[string]string) string {
    for res := variableRE.FindStringSubmatch(value); res != nil; res = variableRE.FindStringSubmatch(value) {
        val := ""
        if vars[res[1]] != "" {
            val = vars[res[1]]
        }
        value = variableRE.ReplaceAllString(value, val)
    }

    return value
}

func getValue(rule ValueStruct, topic, message string, vars map[string]string) string {
    value := getData(rule.Source, topic, message)
    if rule.RE != nil {
        if rule.RE.MatchString(value) {
            value = rule.RE.ReplaceAllString(value, "$1")
        } else {
            value = rule.Default
        }
    } else if rule.Value != "" {
        value = rule.Value
    }

    return applyVariables(value, vars)
}

func getVariables(rules []VariableStruct, topic, message string) map[string]string {
    res := map[string]string{}

    for _, rule := range rules {
        if rule.Name != "" {
            if rule.RE != nil {
                res[rule.Name] = rule.RE.ReplaceAllString(getData(rule.Source, topic, message), "$1")
            } else {
                res[rule.Name] = rule.Value
            }
        }
    }

    return res
}

func checkFilterValue(filter FilterStruct, value string) bool {
    return (filter.RE != nil && filter.RE.MatchString(value)) ||
            (filter.Value != "" && filter.Value == value)
}

func checkResultFilter(filter []FilterStruct, topic, message, destination, value string) bool {
    res := true

    for _, f := range filter {
        if !checkFilterValue(f, getResultData(f.Source, topic, message, destination, value)) {
            res = false
            break
        }
    }

    return res
}

func checkFilter(filter []FilterStruct, topic, message string) bool {
    return checkResultFilter(filter, topic, message, "", "")
}

func compileFilters(filters []FilterStruct) ([]FilterStruct) {
    for inx, filter := range filters {
        if filter.Source == "" {
            filters[inx].Source = "message"
        }
        if filter.Regexp != "" {
            filters[inx].RE = regexp.MustCompile(".*" + filter.Regexp + ".*")
        }
    }

    return filters
}


func showDebugSourceMessage(id uint64, topic, message string) {
    log.Printf("[%5d]        source topic: %s\n", id, topic)
    log.Printf("[%5d]      source message: %s\n", id, message)
}

func main() {
    flag.Parse()

    if debug {
        log.Println("Started")
    }

    config, err := readConfig(configFile)

    if err != nil {
        log.Fatal(err)
    }

    if mqttHost == defaultMQTThost {
        mqttHost = config.MQTT.Host
    }

    mqtt.Connect(mqttHost)

    for _, rule := range config.Rule {
        if checkRuleCorrect(rule) {
            rule.Filter = compileFilters(rule.Filter)

            for inx, result := range rule.Result {
                if (result.Destination != "") {
                    if result.Value.Source == "" {
                        rule.Result[inx].Value.Source = "message"
                    }
                    if result.Value.Regexp != "" {
                        rule.Result[inx].Value.RE = regexp.MustCompile(".*" + result.Value.Regexp + ".*")
                    }

                    rule.Result[inx].Filter = compileFilters(result.Filter)

                    for i, _ := range result.Variable {
                        if result.Variable[i].Regexp != "" {
                            rule.Result[inx].Variable[i].RE = regexp.MustCompile(".*" + result.Variable[i].Regexp + ".*")
                        }
                    }
                }
            }

            go func(rule RuleStruct) {
                mqtt.Subscribe(rule.Source, func(topic, message []byte) {
                    tpc := strings.TrimSpace(string(topic[:]))
                    msg := strings.TrimSpace(string(message[:]))
                    id := system.GetGID()
                    if (checkFilter(rule.Filter, tpc, msg)) {
                        for _, result := range rule.Result {
                            if (result.Destination != "") {
                                variables := getVariables(result.Variable, tpc, msg)
                                value := getValue(result.Value, tpc, msg, variables)
                                dest := applyVariables(result.Destination, variables)
                                filtered := true
                                if (checkResultFilter(rule.Filter, tpc, msg, dest, value)) {
                                    filtered = false
                                    mqtt.Publish(dest, value, false)
                                }
                                if debug && !filtered {
                                    if filtered {
                                        log.Printf("[%5d]     result filtered\n", id)
                                    }
                                    showDebugSourceMessage(id, tpc, msg)
                                    log.Printf("[%5d]           variables: %+v\n", id, variables)
                                    log.Printf("[%5d]   destination topic: %s\n", id, dest)
                                    log.Printf("[%5d] destination message: %s\n", id, value)
                                }
                            }
                        }
                    } else if debug {
                        log.Printf("[%5d]     source filtered\n", id)
                        showDebugSourceMessage(id, tpc, msg)
                    }
                })
            }(rule)
        }
    }

    c := time.Tick(time.Second)
    for _ = range c {}
}
