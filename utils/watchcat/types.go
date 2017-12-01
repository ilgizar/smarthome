package main

import (
    "sync"
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
