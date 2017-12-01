package main

import (
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
        t := now.Hour() * 60 + now.Minute()
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

func checkUsage() {
    for _, rule := range usageConfig.Rule {
        if rule.Enabled {
            for _, cond := range rule.Offline {
                if checkCondition(cond) {
                    for _, node := range rule.Nodes {
                        if n, ok := sharedData.nodes[node]; ok {
                            if checkOfflineState(node, cond) {
                                initNodeActions(node, cond)
                            }
                            if n.active {
                                actionNode(node)
                            }
                        }
                    }
                    break
                }
            }

            for _, cond := range rule.Online {
                if checkCondition(cond) {
                    for _, node := range rule.Nodes {
                        if n, ok := sharedData.nodes[node]; ok {
                            if checkOnlineState(node, cond) {
                                initNodeActions(node, cond)
                            }
                            if n.active {
                                actionNode(node)
                            }
                        }
                    }
                    break
                }
            }
        }
    }
}
