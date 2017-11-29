package main

import (
    "flag"
    "log"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/ilgizar/smarthome/libs/files"
    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
    "github.com/ilgizar/smarthome/libs/system"
)

type NodeStruct struct {
    title          string
    ip             string
    proto          string
    state          bool
}

type DataStruct struct {
    sync.Mutex
    types          map[string]bool
    nodes          map[string]NodeStruct
}

type ConfigStruct struct {
    Main struct {
        Debug      bool
        Nodes      string
        Usage      string
    }
    MQTT struct {
        Host       string
        User       string
        Password   string
        Topic      string
    }
}

const (
    defaultNodesConfig = "nodes.conf"
    defaultUsageConfig = "usage.conf"
)

var debug          bool
var cmdDebug       bool
var configFile     string
var sharedData     DataStruct
var config         ConfigStruct
var nodeConfig     smarthome.NodesConfigStruct
var usageConfig    smarthome.UsageConfigStruct



func init() {
    flag.BoolVar(&cmdDebug,      "debug",   false,             "debug mode")
    flag.StringVar(&configFile,  "config",  "watchcat.conf",   "path to config file")

    sharedData = DataStruct{
        types:  make(map[string]bool),
        nodes:  make(map[string]NodeStruct),
    }

    for _, t := range smarthome.TypesOfDays {
        sharedData.types[t] = false
    }
}

func initDebug() {
    debug = cmdDebug || config.Main.Debug
}

func mqttConnect() {
    mqtt.Connect(config.MQTT.Host)
}

func checkCondition(cond smarthome.UsageConfigConditionStruct) bool {
    res := true

    if len(cond.DatePeriod) > 0 {
    }

    if len(cond.Weekdays) > 0 {
    }

    if len(cond.Time) > 0 {
    }

    return res
}

func checkUsage() {
    for _, rule := range usageConfig.Rule {
        if len(rule.Nodes) > 0 {
            for _, cond := range rule.Denied {
                if checkCondition(cond) {
                }
            }

            for _, node := range rule.Nodes {
                if _, ok := sharedData.nodes[node]; ok {
                }
            }
        }
    }
}

func logUnattended(node, service, attribute, value string) {
    if debug {
        log.Printf("Unattended node(%s) service(%s) attribute(%s) value(%s)\n", node, service, attribute, value)
    }
}

func nodeSubscribe() {
    mqtt.Subscribe(config.MQTT.Topic, func(topic, message []byte) {
        value := string(message)
        parts := strings.Split(string(topic), "/")
        attribute := parts[len(parts) - 1]
        service := parts[len(parts) - 2]
        node := parts[len(parts) - 3]
        if _, ok := sharedData.nodes[node]; ok {
            switch service {
                case "ping":
                    switch attribute {
                        case "droprate":
                            v, err := strconv.ParseInt(value, 10, 64)
                            if err == nil {
                                m := sharedData.nodes[node]
                                m.state = v != 100
                                sharedData.Lock()
                                sharedData.nodes[node] = m
                                sharedData.Unlock()
                            } else if debug {
                                log.Printf("Failed convert to int value for node %s: %s", node, value)
                            }
                        default:
                            logUnattended(node, service, attribute, value)
                    }
                default:
                    logUnattended(node, service, attribute, value)
            }
        } else if debug {
            log.Printf("Unknown node: %s", node)
        }
    })
}

func initNodes() {
    for _, node := range nodeConfig.Node {
        sharedData.nodes[node.Name] = NodeStruct{
            title: node.Title,
            ip:    node.IP,
            proto: node.Protocol,
            state: false,
        }
    }
}

func readMainConfig() {
    if err := files.ReadTypedConfig(configFile, &config); err != nil {
        log.Fatal(err)
    }

    if config.Main.Nodes == "" {
        config.Main.Nodes = defaultNodesConfig
    }

    if config.Main.Usage == "" {
        config.Main.Usage = defaultUsageConfig
    }
}

func parseDateTime(value string) (time.Time, error) {
    return time.Parse("02.01.2006 15:04:05", value)
}

func timeToTimestamp(value string) int64 {
    now := time.Now()
    t, err := parseDateTime(now.Format("02.01.2006") + " " + value + ":00")
    if err == nil {
        return t.Unix()
    }

    return -1
}

func dateToTimestamp(value string, dawn bool) int {
    time := "00:00:00"
    if !dawn {
        time = "23:59:59"
    }
    t, err := parseDateTime(value + " " + time)
    if err == nil {
        return int(t.Unix())
    }

    return -1
}

func weekdayToInt(weekday string) int {
    for i, wd := range system.Days {
        if wd == weekday {
            return i
        }
    }

    return -1
}

func timeToMinutes(time string) int {
    parts := strings.Split(time, ":")
    hour, err := strconv.ParseInt(parts[0], 10, 64)
    if err == nil && hour >=0 && hour <= 23 {
        min, err := strconv.ParseInt(parts[1], 10, 64)
        if err == nil && min >= 0 && min <= 59 {
            return int(hour * 60 + min)
        }
    }

    return -1
}

func prepareCondition(cond *smarthome.UsageConfigConditionStruct) {
    if len(cond.Date) > 0 {
        var begin int
        var end   int
        periods := []smarthome.UsageConfigPeriodStruct{}
        for _, d := range cond.Date {
            dp := strings.Split(d, "-")
            begin = dateToTimestamp(dp[0], true)
            if len(dp) == 2 {
                end = dateToTimestamp(dp[1], false)
            } else {
                end = dateToTimestamp(dp[0], false)
            }
            if begin != -1 && end != -1 {
                if begin > end {
                    begin, end = end, begin
                }
                period := smarthome.UsageConfigPeriodStruct{
                    Begin: begin,
                    End: end,
                }
                periods = append(periods, period)
            }
        }
        cond.DatePeriod = periods
    }

    if len(cond.Weekday) > 0 {
        var begin int
        var end   int
        periods := []string{}
        for _, w := range cond.Weekday {
            wp := strings.Split(w, "-")
            begin = weekdayToInt(wp[0])
            if len(wp) == 2 {
                end = weekdayToInt(wp[1])
            } else {
                end = weekdayToInt(wp[0])
            }
            if begin != -1 && end != -1 {
                if begin > end {
                    for wd := begin; wd <= 6; wd++ {
                        periods = append(periods, time.Weekday(wd).String())
                    }
                    for wd := 0; wd <= end; wd++ {
                        periods = append(periods, time.Weekday(wd).String())
                    }
                } else {
                    for wd := begin; wd <= end; wd++ {
                        periods = append(periods, time.Weekday(wd).String())
                    }
                }
            }
        }
        cond.Weekdays = periods
    }

    if len(cond.Time) > 0 {
        var begin int
        var end   int
        periods := []smarthome.UsageConfigPeriodStruct{}
        for _, t := range cond.Time {
            tp := strings.Split(t, "-")
            begin = timeToMinutes(tp[0])
            if len(tp) == 2 {
                end = timeToMinutes(tp[1])
            } else {
                end = timeToMinutes(tp[0])
            }
            if begin != -1 && end != -1 {
                if begin > end {
                    begin, end = end, begin
                }
                period := smarthome.UsageConfigPeriodStruct{
                    Begin: begin,
                    End: end,
                }
                periods = append(periods, period)
            }
        }
        cond.TimePeriod = periods
    }
}

func readUsageConfig() {
    if err := files.ReadTypedConfig(config.Main.Usage, &usageConfig); err != nil {
        log.Fatal(err)
    }

    for r, rule := range usageConfig.Rule {
        if len(rule.Nodes) > 0 {
            for c, _ := range rule.Denied {
                prepareCondition(&usageConfig.Rule[r].Denied[c])
            }
            for c, _ := range rule.Allowed {
                prepareCondition(&usageConfig.Rule[r].Allowed[c])
            }
/*
            for c, _ := range rule.Limited {
                prepareCondition(&smarthome.UsageConfigConditionStruct(usageConfig.Rule[r].Limited[c]))
            }
            for c, _ := range rule.Offline {
                prepareCondition(&smarthome.UsageConfigConditionStruct(usageConfig.Rule[r].Online[c]))
            }
            for c, _ := range rule.Online {
                prepareCondition(&smarthome.UsageConfigConditionStruct(usageConfig.Rule[r].Offline[c]))
            }
*/
        }
    }

    log.Printf("%+v\n", usageConfig)
}

func main() {
    flag.Parse()

    readMainConfig()

    if err := files.ReadTypedConfig(config.Main.Nodes, &nodeConfig); err != nil {
        log.Fatal(err)
    }

    readUsageConfig()

    initDebug()

    if (debug) {
        log.Println("Started")
    }

    initNodes()

    mqttConnect()

    nodeSubscribe()

    c := time.Tick(time.Second)
    for _ = range c {
        checkUsage()
    }
}
