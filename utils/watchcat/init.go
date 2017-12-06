package main

import (
    "flag"
    "os"
    "os/signal"
    "regexp"
    "syscall"

    "github.com/ilgizar/smarthome/libs/influx"
    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var debug          bool
var cmdDebug       bool
var configFile     string


func initTypes() {
    for _, t := range smarthome.TypesOfDays {
        sharedData.types[t] = false
    }
}

func init() {
    flag.BoolVar(&cmdDebug,      "debug",   false,             "debug mode")
    flag.StringVar(&configFile,  "config",  "watchcat.conf",   "path to config file")

    sharedData = DataStruct{
        types:  make(map[string]bool),
        nodes:  make(map[string]NodeStruct),
    }

    initTypes()

    variableRE = regexp.MustCompile(`\$([a-z0-9-]+)`)
}

func initDebug() {
    debug = cmdDebug || config.Main.Debug
}

func mqttConnect() {
    mqtt.Connect(config.MQTT.Host)
}

func initModes() map[string]NodeModeStruct {
    modes := make(map[string]NodeModeStruct)

    modes["state"] = NodeModeStruct{
        active:    false,
        changed:   false,
        state:     "",
        eventtime: make(map[string]int),
        actions:   []smarthome.UsageConfigActionStruct{},
    }
    modes["state"].eventtime["online"] = 0
    modes["state"].eventtime["offline"] = 0

    modes["permit"] = NodeModeStruct{
        active:    false,
        changed:   false,
        state:     "",
        eventtime: make(map[string]int),
        actions:   []smarthome.UsageConfigActionStruct{},
    }
    modes["permit"].eventtime["allowed"] = 0
    modes["permit"].eventtime["denied"] = 0
    modes["permit"].eventtime["limited"] = 0

    return modes
}

func initNodes() {
    sharedData.Lock()
    for _, node := range nodeConfig.Node {
        sharedData.nodes[node.Name] = NodeStruct{
            name:    node.Name,
            title:   node.Title,
            ip:      node.IP,
            proto:   node.Protocol,
            active:  false,
            modes:   initModes(),
        }
    }
    sharedData.Unlock()
}

func initHUP() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGHUP)

    go func(){
        for _ = range c {
            reloadConfig()
        }
    }()
}

func initInfluxDB() {
    influx.Connect(config.Influx.Host, config.Influx.User, config.Influx.Password, config.Influx.DB)
}
