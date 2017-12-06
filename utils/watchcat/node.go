package main

import (
    "log"
    "time"

    "github.com/ilgizar/smarthome/libs/smarthome"
)

type NodeModeStruct struct {
    active         bool
    changed        bool
    prepared       bool
    state          string
    eventtime      map[string]int
    actions        []smarthome.UsageConfigActionStruct
}

type NodeStruct struct {
    name           string
    title          string
    ip             string
    proto          string

    active         bool
    modes          map[string]NodeModeStruct
}


func nodeExists(nodeName string) bool {
    _, ok := sharedData.nodes[nodeName]

    return ok
}

func clearNodeActions(nodeName string, modeName string) {
    if !nodeExists(nodeName) {
        return
    }

    sharedData.Lock()
    n := sharedData.nodes[nodeName]
    clear := true
    if modeName != "" {
        m := "state"
        if modeName == "state" {
            m = "permit"
        }
        clear = len(n.modes[m].actions) == 0
    }
    if clear {
        n.active = false
    }

    modes := []string{}
    if modeName == "" {
        modes = append(modes, "state", "permit")
    } else {
        modes = append(modes, modeName)
    }
    for _, mode := range modes {
        m := n.modes[mode]
        m.actions = m.actions[:0]
        m.active = false
        m.changed = false
        n.modes[mode] = m
    }

    sharedData.nodes[nodeName] = n
    sharedData.Unlock()
}

func checkNodeState(nodeName string, cond smarthome.UsageConfigConditionStruct, state string) bool {
    log.Printf("checkNodeState('%s', '%s')", nodeName, state)
    if !nodeExists(nodeName) {
        return false
    }

    modeName := getModeNameByState(state)
    node := sharedData.nodes[nodeName]
    changedState := node.modes[modeName].changed && node.modes[modeName].state == state
    log.Printf("changedState: %v %+v\n", changedState, node)

    if node.active {
        for _, mode := range []string{"state", "permit"} {
            if !node.modes[mode].active {
                continue
            }

            res := false
            for _, a := range node.modes[mode].actions {
                if a.Enabled {
                    res = true
                    break
                }
            }
            if !res {
                clearNodeActions(nodeName, mode)
            }
        }
    }

    return changedState
}

func initNodeActions(nodeName string, cond smarthome.UsageConfigConditionStruct, state string) {
    if debug {
        log.Printf("initNodeActions('%s', '%s')\n", nodeName, state)
    }

    if !nodeExists(nodeName) {
        return
    }

    modeName := getModeNameByState(state)

    sharedData.Lock()
    n := sharedData.nodes[nodeName]

    mode := n.modes[modeName]
    mode.actions = make([]smarthome.UsageConfigActionStruct, len(cond.Action))
    copy(mode.actions, cond.Action)
    mode.eventtime[state] = int(time.Now().Unix())
    mode.active = true
    mode.changed = false

    n.modes[modeName] = mode
    n.active = true

    sharedData.nodes[nodeName] = n
    sharedData.Unlock()
}
