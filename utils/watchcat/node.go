package main

import (
    "log"
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


func nodeExists(nodeName string) bool {
    _, ok := sharedData.nodes[nodeName]

    return ok
}

func clearNodeActions(nodeName string) {
    if !nodeExists(nodeName) {
        return
    }

    m := sharedData.nodes[nodeName]
    m.active = false
    m.actions = m.actions[:0]

    sharedData.Lock()
    sharedData.nodes[nodeName] = m
    sharedData.Unlock()
}

func checkNodeState(
        nodeName string,
        cond smarthome.UsageConfigConditionStruct,
        state string) bool {
    if !nodeExists(nodeName) {
        return false
    }

    node := sharedData.nodes[nodeName]
    if node.active {
        res := false
        for _, a := range node.actions {
            if a.Enabled {
                res = true
                break
            }
        }

        if !res {
            clearNodeActions(nodeName)
        }

        return false
    }

    if !node.changed || node.state != state {
        return false
    }

    return true
}

func checkOnlineState(
        nodeName string,
        cond smarthome.UsageConfigConditionStruct) bool {
    return checkNodeState(nodeName, cond, "online")
}

func checkOfflineState(
        nodeName string,
        cond smarthome.UsageConfigConditionStruct) bool {
    return checkNodeState(nodeName, cond, "offline")
}

func initNodeActions(nodeName string, cond smarthome.UsageConfigConditionStruct) {
    if debug {
        log.Printf("initNodeActions(%s)\n", nodeName)
    }

    if !nodeExists(nodeName) {
        return
    }

    m := sharedData.nodes[nodeName]
    m.changed = false
    m.active = true
    m.actions = make([]smarthome.UsageConfigActionStruct, len(cond.Action))
    copy(m.actions, cond.Action)
    m.eventtime = int(time.Now().Unix())

    sharedData.Lock()
    sharedData.nodes[nodeName] = m
    sharedData.Unlock()
}
