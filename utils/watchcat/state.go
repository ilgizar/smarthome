package main

import (
    "log"
)


func checkChangeState(nodeName, state string) {
    log.Printf("checkChangeState('%s', '%s')\n", nodeName, state);
    sharedData.Lock()
    mode := sharedData.nodes[nodeName].modes["permit"]
    if mode.changed = mode.state != state; mode.changed {
        log.Printf("State changed node(%s) state(%s)", nodeName, state)
        mode.state = state
        sharedData.nodes[nodeName].modes["permit"] = mode
    }
    sharedData.Unlock()
}

func getModesList(modeName string) []string {
    modes := []string{}
    if modeName == "" {
        modes = append(modes, "state", "permit")
    } else {
        modes = append(modes, modeName)
    }

    return modes
}

func setPreparedState(nodeName, modeName string, value bool) {
    for _, mode := range getModesList(modeName) {
        sharedData.Lock()
        m := sharedData.nodes[nodeName].modes[mode]
        m.prepared = value
        sharedData.nodes[nodeName].modes[mode] = m
        sharedData.Unlock()
    }
}

func getModeByState(state string) string {
    mode := "permit"
    if state == "online" || state == "offline" {
        mode = "state"
    }

    return mode
}
