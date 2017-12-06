package main

import (
    "log"
    "strings"
    "time"

    "github.com/ilgizar/smarthome/libs/files"
    "github.com/ilgizar/smarthome/libs/influx"
    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

type ConfigStruct struct {
    Main struct {
        Debug      bool
        Nodes      string
        Usage      string
        Topic      string
        Delay      int
    }
    MQTT struct {
        Host       string
        User       string
        Password   string
        Topic      string
    }
    Influx struct {
        Host       string
        User       string
        Password   string
        DB         string
    }
}

const (
    defaultNodesConfig = "nodes.conf"
    defaultUsageConfig = "usage.conf"
    defaultTopic       = "smarthome"

    defaultInflux      = "https://localhost:8086"
    defaultDB          = "telegraf"
)

var config         ConfigStruct
var nodeConfig     smarthome.NodesConfigStruct
var usageConfig    smarthome.UsageConfigStruct


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

    if len(cond.Action) > 0 {
        for i, _ := range cond.Action {
            cond.Action[i].Enabled = !cond.Action[i].Disable &&
                cond.Action[i].Type != "" &&
                cond.Action[i].Value != "" &&
                cond.Action[i].Destination != ""
        }
    }
}

func initUsageConfig() {
    for r, rule := range usageConfig.Rule {
        usageConfig.Rule[r].Enabled = !rule.Disable && len(rule.Nodes) > 0
        if usageConfig.Rule[r].Enabled {
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
                prepareCondition(&usageConfig.Rule[r].Online[c])
            }
            for c, _ := range rule.Offline {
                prepareCondition(&usageConfig.Rule[r].Offline[c])
            }
        }
    }
}
func readUsageConfig() {
    if err := files.ReadTypedConfig(config.Main.Usage, &usageConfig); err != nil {
        log.Fatal(err)
    }

    initUsageConfig()
}

func readNodesConfig() {
    if err := files.ReadTypedConfig(config.Main.Nodes, &nodeConfig); err != nil {
        log.Fatal(err)
    }
}

func initMainConfig() {
    if config.Main.Nodes == "" {
        config.Main.Nodes = defaultNodesConfig
    }

    if config.Main.Usage == "" {
        config.Main.Usage = defaultUsageConfig
    }

    if config.Main.Topic == "" {
        config.Main.Topic = defaultTopic
    }

    if config.Influx.Host == "" {
        config.Influx.Host = defaultInflux
    }

    if config.Influx.DB == "" {
        config.Influx.DB = defaultDB
    }
}

func readMainConfig() {
    if err := files.ReadTypedConfig(configFile, &config); err != nil {
        log.Fatal(err)
    }

    initMainConfig()
}

func readConfigs() {
    readMainConfig()
    readNodesConfig()
    readUsageConfig()
}

func reloadConfigFile(configFile string, cfg interface{}) bool {
    if err := files.ReadTypedConfig(configFile, cfg); err != nil {
        log.Printf("Failed reload configuration file '%s': %s\n", configFile, err)
        return false
    }

    return true
}

func reloadConfigs(cfg interface{}, ncfg interface{}, ucfg interface{}) bool {
    return reloadConfigFile(configFile, cfg) &&
            reloadConfigFile(config.Main.Nodes, ncfg) &&
            reloadConfigFile(config.Main.Usage, ucfg)
}

func reloadConfig() {
    if debug {
        log.Println("Reload configuration")
    }

    var cfg ConfigStruct
    var ncfg smarthome.NodesConfigStruct
    var ucfg smarthome.UsageConfigStruct
    if !reloadConfigs(&cfg, &ncfg, &ucfg) {
        return
    }

    nodeUnsubscribe()
    time.Sleep(time.Duration(cfg.Main.Delay) * time.Second)
    mqtt.Disconnect()

    influx.Disconnect()

    config = cfg
    initMainConfig()

    initDebug()

    sharedData.Lock()
    sharedData.nodes = map[string]NodeStruct{}
    initTypes()
    sharedData.Unlock()

    nodeConfig = ncfg
    usageConfig = ucfg
    initUsageConfig()

    initNodes()

    initInfluxDB()
    mqttConnect()
    nodeSubscribe()
}
