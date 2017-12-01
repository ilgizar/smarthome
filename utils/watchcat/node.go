package main

import (
    "time"

    "github.com/ilgizar/smarthome/libs/smarthome"
)


func checkNodeState(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct,
        online bool) bool {
    if !node.changed || node.state != online {
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

func checkOnlineState(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) bool {
    return checkNodeState(node, cond, true)
}

func checkOfflineState(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) bool {
    return checkNodeState(node, cond, false)
}

func offChangedState(node string) {
    m := sharedData.nodes[node]
    m.changed = false
    sharedData.Lock()
    sharedData.nodes[node] = m
    sharedData.Unlock()
}
