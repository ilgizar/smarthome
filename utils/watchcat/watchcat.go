package main

import (
    "flag"
    "log"
    "regexp"
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
    name           string
    title          string
    ip             string
    proto          string
    state          bool
    changed        bool
    online         int
    offline        int
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
var variableRE     *regexp.Regexp



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

    variableRE = regexp.MustCompile(`\$([a-z0-9-]+)`)
}

func initDebug() {
    debug = cmdDebug || config.Main.Debug
}

func mqttConnect() {
    mqtt.Connect(config.MQTT.Host)
}

func checkCondition(cond smarthome.UsageConfigConditionStruct) bool {
    res := true

    now := time.Now()
    if len(cond.DatePeriod) > 0 {
        ts := int(now.Unix())
        res = false
        for _, v := range cond.DatePeriod {
            if res = ts >= v.Begin && ts <= v.End; res {
                break
            }
        }

        if !res {
            return false
        }
    }

    if len(cond.Weekdays) > 0 {
        wd := now.Weekday().String()
        res = false
        for _, v := range cond.Weekdays {
            if res = v == wd; res {
                break
            }
        }

        if !res {
            return false
        }
    }

    if len(cond.TimePeriod) > 0 {
        t := now.Hour() * 60 + now.Minute()
        res = false
        for _, v := range cond.TimePeriod {
            if res = t >= v.Begin && t <= v.End; res {
                break
            }
        }

        if !res {
            return false
        }
    }

    return res
}

func checkNodeState(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct,
        online bool) bool {
    if !node.changed || node.state != online {
        return false
    }

    currentState := node.offline
    previousState := node.online
    if online {
        currentState, previousState = previousState, currentState
    }

    now := int(time.Now().Unix())

    return currentState + cond.Pause * 60 < now &&
            (cond.After == 0 ||
                previousState + cond.After * 60 < now) &&
            (cond.Before == 0 ||
                currentState + cond.Before * 60 > now)
}

func checkOnlineState(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) bool {
    return checkNodeState(node, cond, true)
}

func checkOfflineState(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) bool {
    return checkNodeState(node, cond, false)
}

func convertActionValue(value string,
        node NodeStruct) string {
    for res := variableRE.FindStringSubmatch(value); res != nil; res = variableRE.FindStringSubmatch(value) {
        val := ""
        switch res[1] {
            case "name":
                val = node.name
            case "title":
                val = node.title
            case "ip":
                val = node.ip
        }
        value = variableRE.ReplaceAllString(value, val)
    }

    return value
}

func actionNode(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) {
    for _, a := range cond.Action {
        if a.Value != "" {
            a.Value = convertActionValue(a.Value, node)
            switch a.Type {
                case "say":
log.Printf("Say on %s: %s\n", a.Destination, a.Value)
                case "telegram":
log.Printf("Telegram to %s: %s\n", a.Destination, a.Value)
                case "command":
                    a.Value = convertActionValue(a.Value, node)
log.Printf("Command to %s: %s\n", a.Destination, a.Value)
                case "mqtt":
log.Printf("MQTT publish to %s: %s\n", a.Destination, a.Value)
            }
        }
    }
}

func offChangedState(node string) {
    m := sharedData.nodes[node]
    m.changed = false
    sharedData.Lock()
    sharedData.nodes[node] = m
    sharedData.Unlock()
}

func checkUsage() {
    for _, rule := range usageConfig.Rule {
        if rule.Enabled {
            for _, cond := range rule.Offline {
                if checkCondition(cond.UsageConfigConditionStruct) {
                    for _, node := range rule.Nodes {
                        if n, ok := sharedData.nodes[node]; ok {
                            if checkOfflineState(n, cond) {
                                offChangedState(node)
                                actionNode(n, cond)
                            }
                        }
                    }
                    break
                }
            }

            for _, cond := range rule.Online {
                if checkCondition(cond.UsageConfigConditionStruct) {
                    for _, node := range rule.Nodes {
                        if n, ok := sharedData.nodes[node]; ok {
                            if checkOnlineState(n, cond) {
                                offChangedState(node)
                                actionNode(n, cond)
                            }
                        }
                    }
                    break
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
                                state := v != 100
                                if state != m.state {
log.Printf("Changed state %s %+v -> %+v\n", node, m.state, state)
                                    m.state = state
                                    m.changed = true
                                    now := int(time.Now().Unix())
                                    if m.state {
                                        m.online = now
                                    } else {
                                        m.offline = now
                                    }
                                    sharedData.Lock()
                                    sharedData.nodes[node] = m
                                    sharedData.Unlock()
                                }
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
            name:  node.Name,
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
        usageConfig.Rule[r].Enabled = rule.Enable && len(rule.Nodes) > 0
        if rule.Enable {
            for c, _ := range rule.Denied {
                prepareCondition(&usageConfig.Rule[r].Denied[c])
            }
            for c, _ := range rule.Allowed {
                prepareCondition(&usageConfig.Rule[r].Allowed[c])
            }
            for c, _ := range rule.Limited {
                prepareCondition(&usageConfig.Rule[r].Limited[c].UsageConfigConditionStruct)
            }
            for c, _ := range rule.Online {
                prepareCondition(&usageConfig.Rule[r].Online[c].UsageConfigConditionStruct)
            }
            for c, _ := range rule.Offline {
                prepareCondition(&usageConfig.Rule[r].Offline[c].UsageConfigConditionStruct)
            }
        }
    }
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
