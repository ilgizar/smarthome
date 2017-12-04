package main

import (
    "log"
    "regexp"
    "time"

    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var variableRE     *regexp.Regexp


func checkActionState(node NodeStruct, action smarthome.UsageConfigActionStruct) bool {
    if !action.Enabled {
        return false
    }

    now := int(time.Now().Unix())

    switch node.state {
        case "online", "offline":
            currentState := node.offline
            previousState := node.online
            if node.state == "online" {
                currentState, previousState = previousState, currentState
            }

            return currentState + action.Pause < now &&
                (action.After == 0 ||
                    previousState + action.After < now) &&
                (action.Before == 0 ||
                    currentState + action.Before > now)
        default:
            return node.eventtime + action.Pause < now
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

func actionNode(nodeName string) {
    if debug {
        log.Printf("actionNode(%s)\n", nodeName)
    }

    node := sharedData.nodes[nodeName]
    for a, action := range node.actions {
        if checkActionState(node, action) {
            offActionState(&sharedData.nodes[nodeName].actions[a])
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
}
