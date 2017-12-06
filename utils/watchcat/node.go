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

func clearNodeActions(nodeName string, mode string) {
    if !nodeExists(nodeName) {
        return
    }

    sharedData.Lock()
    n := sharedData.nodes[nodeName]
    clear := true
    if mode != "" {
        m := "state"
        if mode == "state" {
            m = "permit"
        }
        clear = len(n.modes[m].actions) > 0
    }
    if clear {
        n.active = false
    }

    modes := []string{}
    if mode == "" {
        modes = append(modes, "state", "permit")
    } else {
        modes = append(modes, mode)
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

    mode := getModeByState(state)
    node := sharedData.nodes[nodeName]
    changedState := node.modes[mode].changed && node.modes[mode].state == state
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

        if !node.modes["state"].active && !node.modes["permit"].active {
            node.active = false
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

    mode := getModeByState(state)

    sharedData.Lock()
    n := sharedData.nodes[nodeName]

    m := n.modes[mode]
    m.actions = make([]smarthome.UsageConfigActionStruct, len(cond.Action))
    copy(m.actions, cond.Action)
    m.eventtime[state] = int(time.Now().Unix())
    m.active = true
    m.changed = false

    n.modes[mode] = m
    n.active = true

    sharedData.nodes[nodeName] = n
    sharedData.Unlock()
}
