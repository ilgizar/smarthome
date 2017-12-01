package main

import (
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

            return currentState + action.Pause * 60 < now &&
                (action.After == 0 ||
                    previousState + action.After * 60 < now) &&
                (action.Before == 0 ||
                    currentState + action.Before * 60 > now)
        default:
            return node.eventtime + action.Pause * 60 < now
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

func actionNode(n string) {
    node := sharedData.nodes[n]
    for a, action := range node.actions {
        if checkActionState(node, action) {
            offActionState(&sharedData.nodes[n].actions[a])
            action.Value = convertActionValue(action.Value, node)
            action.Destination = convertActionValue(action.Destination, node)
            switch action.Type {
                case "say", "telegram", "controller":
                    mqtt.Publish(config.Main.Topic + "/util/" + action.Type + "/" + action.Destination, action.Value, false)
                case "mqtt":
                    mqtt.Publish(action.Destination, action.Value, false)
            }
        }
    }
}
