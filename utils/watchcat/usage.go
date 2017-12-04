package main

import (
    "fmt"
    "log"
    "time"

    "github.com/ilgizar/smarthome/libs/smarthome"
)


func checkCondition(cond smarthome.UsageConfigConditionStruct) bool {
    res := true

    now := time.Now()
    if len(cond.DatePeriod) > 0 {
        ts := int(now.Unix())
        res = false
        for _, v := range cond.DatePeriod {
            if res = ts >= v.Begin && ts <= v.End; res {
                break
            }
        }

        if !res {
            return false
        }
    }

    if len(cond.Weekdays) > 0 {
        wd := now.Weekday().String()
        res = false
        for _, v := range cond.Weekdays {
            if res = v == wd; res {
                break
            }
        }

        if !res {
            return false
        }
    }

    if len(cond.TimePeriod) > 0 {
        t := timeToMinutes(fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute()))
        res = false
        for _, v := range cond.TimePeriod {
            if res = t >= v.Begin && t <= v.End; res {
                break
            }
        }

        if !res {
            return false
        }
    }

    return res
}

func checkUsage(state string, nodeName string) {
    if debug {
        log.Printf("checkUsage('%s', '%s')\n", state, nodeName)
    }

    for _, rule := range usageConfig.Rule {
        if rule.Enabled {
            if state == "" || state == "offline" {
                for _, cond := range rule.Offline {
                    if checkCondition(cond) {
                        for _, node := range rule.Nodes {
                            if nodeName != "" && nodeName != node {
                                return
                            }
                            if n, ok := sharedData.nodes[node]; ok {
                                if checkOfflineState(node, cond) {
                                    initNodeActions(node, cond)
                                }
                                if n.active {
                                    actionNode(node)
                                }
                            }
                        }
                    }
                }
            }

            if state == "" || state == "online" {
                for _, cond := range rule.Online {
                    if checkCondition(cond) {
                        for _, node := range rule.Nodes {
                            if nodeName != "" && nodeName != node {
                                return
                            }
                            if n, ok := sharedData.nodes[node]; ok {
                                if checkOnlineState(node, cond) {
                                    initNodeActions(node, cond)
                                }
                                if n.active {
                                    actionNode(node)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
