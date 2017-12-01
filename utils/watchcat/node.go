package main

import (
    "time"

    "github.com/ilgizar/smarthome/libs/smarthome"
)

type NodeStruct struct {
    name           string
    title          string
    ip             string
    proto          string

    online         int
    offline        int
    eventtime      int
    active         bool
    changed        bool
    state          string
    actions        []smarthome.UsageConfigActionStruct
}


func clearNodeActions(node string) {
    m := sharedData.nodes[node]
    m.active = false
    m.actions = []smarthome.UsageConfigActionStruct{}

    sharedData.Lock()
    sharedData.nodes[node] = m
    sharedData.Unlock()
}

func checkNodeState(
        n string,
        cond smarthome.UsageConfigConditionStruct,
        state string) bool {
    node := sharedData.nodes[n]
    if node.active {
        res := false
        for _, a := range node.actions {
            if a.Enabled {
                res = true
                break
            }
        }

        if !res {
            clearNodeActions(n)
        }

        return res
    }

    if !node.changed || node.state != state {
        return false
    }

    return true
}

func checkOnlineState(
        node string,
        cond smarthome.UsageConfigConditionStruct) bool {
    return checkNodeState(node, cond, "online")
}

func checkOfflineState(
        node string,
        cond smarthome.UsageConfigConditionStruct) bool {
    return checkNodeState(node, cond, "offline")
}

func initNodeActions(node string, cond smarthome.UsageConfigConditionStruct) {
    m := sharedData.nodes[node]
    m.changed = false
    m.active = true
    a := cond.Action
    m.actions = a
    m.eventtime = int(time.Now().Unix())

    sharedData.Lock()
    sharedData.nodes[node] = m
    sharedData.Unlock()
}
