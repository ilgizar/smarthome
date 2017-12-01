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
    state          bool
    changed        bool
    hold           bool
    online         int
    offline        int
}


func checkNodeState(
        node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct,
        online bool) bool {
    if node.hold {
        return true
    }
    if !node.changed || !node.hold || node.state != online {
        return false
    }

    currentState := node.offline
    previousState := node.online
    if online {
        currentState, previousState = previousState, currentState
    }

    now := int(time.Now().Unix())

    return currentState + cond.Pause * 60 < now &&
            (cond.After == 0 ||
                previousState + cond.After * 60 < now) &&
            (cond.Before == 0 ||
                currentState + cond.Before * 60 > now)
}

func checkOnlineState(
        node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) bool {
    return checkNodeState(node, cond, true)
}

func checkOfflineState(
        node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) bool {
    return checkNodeState(node, cond, false)
}

func offChangedState(node string) {
    m := sharedData.nodes[node]
    m.changed = false
    m.hold = true
    sharedData.Lock()
    sharedData.nodes[node] = m
    sharedData.Unlock()
}

func offHoldState(node string) {
    m := sharedData.nodes[node]
    m.hold = false
    sharedData.Lock()
    sharedData.nodes[node] = m
    sharedData.Unlock()
}
