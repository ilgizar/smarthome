package main

import (
    "log"
    "regexp"
    "time"

    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var variableRE     *regexp.Regexp


func checkActionState(action smarthome.UsageConfigActionStruct, mode NodeModeStruct) bool {
    if !action.Enabled || !mode.active {
        return false
    }

    now := int(time.Now().Unix())

    switch mode.state {
        case "online", "offline":
            s := "online"
            if mode.state == "online" {
                s = "offline"
            }
            currentState := mode.eventtime[mode.state]
            previousState := mode.eventtime[s]

            return currentState + action.Pause <= now &&
                (action.After == 0 || previousState + action.After <= now) &&
                (action.Before == 0 || currentState + action.Before > now)
        case "allowed", "denied", "limited":
            return mode.eventtime[mode.state] + action.Pause <= now
        default:
            return false
    }
}

func convertActionValue(value string, node NodeStruct) string {
    for res := variableRE.FindStringSubmatch(value); res != nil; res = variableRE.FindStringSubmatch(value) {
        val := ""
        switch res[1] {
            case "name":
                val = node.name
            case "title":
                val = node.title
            case "ip":
                val = node.ip
        }
        value = variableRE.ReplaceAllString(value, val)
    }

    return value
}

func offActionState(action *smarthome.UsageConfigActionStruct) {
    sharedData.Lock()
    action.Enabled = false
    sharedData.Unlock()
}

func loopAction(nodeName string, actionInx int, mode string) {
    log.Printf("loopAction('%s', '%d', '%s')\n", nodeName, actionInx, mode)
    node := sharedData.nodes[nodeName]
    action := sharedData.nodes[nodeName].modes[mode].actions[actionInx]
    if checkActionState(action, node.modes[mode]) {
        offActionState(&sharedData.nodes[nodeName].modes[mode].actions[actionInx])
        action.Value = convertActionValue(action.Value, node)
        action.Destination = convertActionValue(action.Destination, node)
        if debug {
            log.Printf("action node '%s' type '%s' destination '%s' value '%s'\n", nodeName, action.Type, action.Destination, action.Value)
        }
        switch action.Type {
            case "say", "telegram", "controller":
                mqtt.Publish(config.Main.Topic + "/util/" + action.Type + "/" + action.Destination, action.Value, false)
            case "mqtt":
                mqtt.Publish(action.Destination, action.Value, false)
        }
    }
}

func checkActionActive(mode NodeModeStruct) bool {
    log.Printf("checkActionActive: active(%v) prepared(%v)\n", mode.active, mode.prepared)
    return mode.active && !mode.prepared
}

func actionNode(nodeName, mode string) {
    if debug {
        log.Printf("actionNode('%s', '%s')\n", nodeName, mode)
    }

    m := sharedData.nodes[nodeName].modes[mode]
    if checkActionActive(m) {
        setPreparedState(nodeName, mode, true)
        for a, _ := range m.actions {
            loopAction(nodeName, a, mode)
        }
    }
}
