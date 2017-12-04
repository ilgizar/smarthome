package main

import (
    "flag"
    "log"
    "sync"
    "time"
)

type DataStruct struct {
    sync.Mutex
    types          map[string]bool
    nodes          map[string]NodeStruct
}

var sharedData     DataStruct


func main() {
    flag.Parse()

    readConfigs()

    initDebug()

    if (debug) {
        log.Println("Started")
    }

    initHUP()

    initNodes()

    mqttConnect()
    nodeSubscribe()

    c := time.Tick(time.Second)
    for _ = range c {
        checkUsage("", "")
    }
}
