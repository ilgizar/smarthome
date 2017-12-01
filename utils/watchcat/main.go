package main

import (
    "flag"
    "log"
    "time"

    "github.com/ilgizar/smarthome/libs/files"
)


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
