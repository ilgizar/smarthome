package main

import (
    "flag"
    "regexp"

    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var debug          bool
var cmdDebug       bool
var configFile     string


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
