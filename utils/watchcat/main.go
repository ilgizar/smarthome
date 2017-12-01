package main

import (
    "flag"
    "log"
    "sync"
    "time"

    "github.com/ilgizar/smarthome/libs/files"
)

type DataStruct struct {
    sync.Mutex
    types          map[string]bool
    nodes          map[string]NodeStruct
}

var sharedData     DataStruct


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
