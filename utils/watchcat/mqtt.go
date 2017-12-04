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
        value := string(message)
        parts := strings.Split(string(topic), "/")
        attribute := parts[len(parts) - 1]
        service := parts[len(parts) - 2]
        node := parts[len(parts) - 3]
        if _, ok := sharedData.nodes[node]; ok {
            switch service {
                case "ping":
                    switch attribute {
                        case "droprate":
                            v, err := strconv.ParseInt(value, 10, 64)
                            if err == nil {
                                m := sharedData.nodes[node]
                                state := v != 100
                                if (state && m.state != "online") || (!state && m.state != "offline") {
                                    now := int(time.Now().Unix())
                                    if state {
                                        log.Printf("Change node '%s' state '%s' -> 'online'", node, m.state)
                                        m.state = "online"
                                        m.online = now
                                    } else {
                                        log.Printf("Change node '%s' state '%s' -> 'offline'", node, m.state)
                                        m.state = "offline"
                                        m.offline = now
                                    }
                                    m.changed = true

                                    sharedData.Lock()
                                    sharedData.nodes[node] = m
                                    sharedData.Unlock()

                                    clearNodeActions(node)

                                    checkUsage(m.state, node)
                                }
                            } else if debug {
                                log.Printf("Failed convert to int value for node %s: %s", node, value)
                            }
                        default:
                            logUnattended(node, service, attribute, value)
                    }
                default:
                    logUnattended(node, service, attribute, value)
            }
        } else if debug {
            log.Printf("Unknown node: %s", node)
        }
    })
}

func nodeUnsubscribe() {
    mqtt.Unsubscribe(config.MQTT.Topic)
}
