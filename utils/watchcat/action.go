package main

import (
    "regexp"

    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var variableRE     *regexp.Regexp


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

func actionNode(node NodeStruct, actions *[]smarthome.UsageConfigActionStruct) {
    for _, a := range *actions {
        if a.Value != "" && a.Destination != "" {
            a.Value = convertActionValue(a.Value, node)
            a.Destination = convertActionValue(a.Destination, node)
            switch a.Type {
                case "say", "telegram", "controller":
                    mqtt.Publish("smarthome/util/" + a.Type + "/" + a.Destination, a.Value, false)
                case "mqtt":
                    mqtt.Publish(a.Destination, a.Value, false)
            }
        }
    }
}
