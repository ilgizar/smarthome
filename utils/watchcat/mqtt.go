package main

import (
    "log"
    "strconv"
    "strings"
    "time"

    "github.com/ilgizar/smarthome/libs/mqtt"
)


func logUnattended(node, service, attribute, value string) {
    if debug {
        log.Printf("Unattended node(%s) service(%s) attribute(%s) value(%s)\n", node, service, attribute, value)
    }
}

func nodeSubscribe() {
    mqtt.Subscribe(config.MQTT.Topic, func(topic, message []byte) {
        parts := strings.Split(string(topic), "/")
        node := parts[len(parts) - 3]
        if !nodeExists(node) {
            if debug {
                log.Printf("Unknown node: %s", node)
            }

            return
        }

        value := string(message)
        attribute := parts[len(parts) - 1]
        service := parts[len(parts) - 2]

        switch service {
            case "ping":
                switch attribute {
                    case "droprate":
                        v, err := strconv.ParseInt(value, 10, 64)
                        if err != nil {
                            if debug {
                                log.Printf("Failed convert value (%s) to int for node %s: %s", value, node, err)
                            }

                            return
                        }

                        state := v != 100
                        mode := "state"
                        if (state && sharedData.nodes[node].modes[mode].state != "online") ||
                                (!state && sharedData.nodes[node].modes[mode].state != "offline") {
                            actionNode(node, "state", "end")
                            clearNodeActions(node, mode)

                            sharedData.Lock()
                            n := sharedData.nodes[node]

                            m := n.modes[mode]
                            s := "offline"
                            if state {
                                s = "online"
                            }
                            log.Printf("Change node '%s' state '%s' -> '%s'", node, n.modes[mode].state, s)
                            m.state = s
                            m.changed = true
                            m.eventtime[s] = int(time.Now().Unix())
                            n.modes[mode] = m

                            sharedData.nodes[node] = n
                            sharedData.Unlock()

                            checkUsage(s, node)
                        }
                    default:
                        logUnattended(node, service, attribute, value)
                }
            default:
                logUnattended(node, service, attribute, value)
        }
    })
}

func nodeUnsubscribe() {
    mqtt.Unsubscribe(config.MQTT.Topic)
}
