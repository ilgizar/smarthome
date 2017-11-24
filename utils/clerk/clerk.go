package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "regexp"
    "time"

    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/influxdata/toml"
)

type ValueStruct struct {
    Source string
    Value  string
    RE     *regexp.Regexp
}

type VariableStruct struct {
    Name   string
    Source string
    Value  string
    RE     *regexp.Regexp
}

type RuleStruct struct {
    Source      string
    Destination string
    Value       ValueStruct
    Variable    []VariableStruct
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

func init() {
    flag.BoolVar(&debug,         "debug",   false,             "debug mode")
    flag.StringVar(&mqttHost,    "mqtt",    defaultMQTThost,   "MQTT server, can set without port number")
    flag.StringVar(&configFile,  "config",  "clerk.conf",      "path to config file")
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
    return rule.Source != "" &&
        rule.Destination != "" &&
        rule.Value.Value != ""
}

func getData(source, topic, message string) string {
    if source == "topic" {
        return topic
    }

    return message
}

func getValue(rule ValueStruct, topic, message string) string {
    return rule.RE.ReplaceAllString(getData(rule.Source, topic, message), "$1")
}

func getVariables(rules []VariableStruct, topic, message string) map[string]string {
    res := map[string]string{}

    for _, rule := range rules {
        if rule.Name != "" {
            res[rule.Name] = rule.RE.ReplaceAllString(getData(rule.Source, topic, message), "$1")
        }
    }

    return res
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
            if rule.Value.Source == "" {
                rule.Value.Source = "message"
            }
            rule.Value.RE = regexp.MustCompile(".*" + rule.Value.Value + ".*")
            for i, _ := range rule.Variable {
                rule.Variable[i].RE = regexp.MustCompile(".*" + rule.Variable[i].Value + ".*")
            }
            mqtt.Subscribe(rule.Source, func(topic, message []byte) {
                tpc := string(topic[:])
                msg := string(message[:])
                value := getValue(rule.Value, tpc, msg)
                variables := getVariables(rule.Variable, tpc, msg)
                fmt.Printf("%s %s %s %+v\n", tpc, msg, value, variables)
            })
        }
    }

    time.Sleep(60 * time.Minute)
}
