package main

import (
    "log"

    "github.com/ilgizar/smarthome/libs/smarthome"
)


func convertActionValue(value string,
        node NodeStruct) string {
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

func actionNode(node NodeStruct,
        cond smarthome.UsageConfigOnlineStruct) {
    for _, a := range cond.Action {
        if a.Value != "" {
            a.Value = convertActionValue(a.Value, node)
            switch a.Type {
                case "say":
log.Printf("Say on %s: %s\n", a.Destination, a.Value)
                case "telegram":
log.Printf("Telegram to %s: %s\n", a.Destination, a.Value)
                case "command":
                    a.Value = convertActionValue(a.Value, node)
log.Printf("Command to %s: %s\n", a.Destination, a.Value)
                case "mqtt":
log.Printf("MQTT publish to %s: %s\n", a.Destination, a.Value)
            }
        }
    }
}
