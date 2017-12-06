package main

import (
    "log"
)


func checkChangeState(nodeName, state string) bool {
    log.Printf("checkChangeState('%s', '%s')\n", nodeName, state);
    mode := sharedData.nodes[nodeName].modes["permit"]
    changed := mode.state != state
    if changed {
        log.Printf("State changed node(%s) state(%s -> %s)", nodeName, mode.state, state)
    }

    return changed
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

func getModeNameByState(state string) string {
    modeName := "permit"
    if state == "online" || state == "offline" {
        modeName = "state"
    }

    return modeName
}

func setNodeState(nodeName, state string) {
    sharedData.Lock()
    mode := sharedData.nodes[nodeName].modes["permit"]
    log.Printf("setNodeState('%s', '%s -> %s')", nodeName, mode.state, state)
    mode.changed = mode.changed || mode.state != state
    mode.state = state
    sharedData.nodes[nodeName].modes["permit"] = mode
    sharedData.Unlock()
}
