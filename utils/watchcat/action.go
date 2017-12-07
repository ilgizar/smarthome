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
    log.Printf("checkActionState: enabled(%v) active(%v) state(%s)", action.Enabled, mode.active, mode.state)
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
        case "allowed", "denied":
            return mode.eventtime[mode.state] + action.Pause <= now
        case "limited":
            return true
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

func loopAction(nodeName string, actionInx int, modeName string) {
    log.Printf("loopAction('%s', '%d', '%s')\n", nodeName, actionInx, modeName)
    node := sharedData.nodes[nodeName]
    action := sharedData.nodes[nodeName].modes[modeName].actions[actionInx]
    if checkActionState(action, node.modes[modeName]) {
        offActionState(&sharedData.nodes[nodeName].modes[modeName].actions[actionInx])
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

func setModeEvent(nodeName, modeName, event string) {
    if sharedData.nodes[nodeName].modes[modeName].event == event {
        return
    }

    sharedData.Lock()
    m := sharedData.nodes[nodeName].modes[modeName]
    m.event = event
    sharedData.nodes[nodeName].modes[modeName] = m
    sharedData.Unlock()
    log.Printf("setModeEvent('%s', '%s', '%s'): %+v", nodeName, modeName, event, m)
}

func actionNode(nodeName, modeName string, event string) {
    mode := sharedData.nodes[nodeName].modes[modeName]
    if debug {
        log.Printf("actionNode('%s', '%s', '%s')\n", nodeName, modeName, event)
    }

    log.Printf("checkActionActive: active(%v) prepared(%v)\n", mode.active, mode.prepared)
    if !mode.active {
        return
    }
    if event != "end" && mode.prepared {
        return
    }

    if event != "begin" || mode.event == "" || mode.event == "end" {
        setModeEvent(nodeName, modeName, event)
    }

    for a, action := range mode.actions {
        if action.Event != event {
            continue
        }

        loopAction(nodeName, a, modeName)
    }
}
