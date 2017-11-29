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

type MainConfigStruct struct {
    Main struct {
        Debug    bool
        Nodes    string
    }

    MQTT struct {
        Host     string
        User     string
        Password string
        Topic    string
    }
}


type OptionConfigStruct struct {
    Name         string
    Value        string
}

type NodeConfigStruct struct {
    Name         string
    Title        string
    IP           string
    Protocol     string
    Option       []OptionConfigStruct
}

type NodesConfigStruct struct {
    Node         []NodeConfigStruct
}

var debug          bool
var cmdDebug       bool
var configFile     string
var sharedData     DataStruct
var config         MainConfigStruct
var nodeConfig     NodesConfigStruct



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

func refreshNodes() {
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

func main() {
    flag.Parse()

    if err := files.ReadTypedConfig(configFile, &config); err != nil {
        log.Fatal(err)
    }

    if err := files.ReadTypedConfig(config.Main.Nodes, &nodeConfig); err != nil {
        log.Fatal(err)
    }

    initDebug()

    if (debug) {
        log.Println("Started")
    }

    initNodes()

    mqttConnect()

    nodeSubscribe()

    c := time.Tick(time.Second)
    for _ = range c {
        refreshNodes()
    }
}
