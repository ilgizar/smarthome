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

func loopCondition(rule smarthome.UsageConfigRuleStruct, cond smarthome.UsageConfigConditionStruct, nodeName string, state string) {
    mode := getModeByState(state)

    ok := checkCondition(cond)
    if !(ok || (mode == "permit" && state != "limited")) {
        return
    }

    log.Printf("loopCondition: state(%s)\n", state)
    for _, node := range rule.Nodes {
        if (nodeName != "" && nodeName != node) || !nodeExists(node) {
            continue
        }

        log.Printf("loopCondition: node(%s) mode(%s)\n", node, mode)
        if mode == "permit" {
            if ok {
                checkChangeState(node, state)
            } else if sharedData.nodes[node].modes[mode].state == state {
                checkChangeState(node, "limited")
            } else {
                continue
            }
        }

        if checkNodeState(node, cond, state) {
            initNodeActions(node, cond, state)
        }

        if sharedData.nodes[node].active {
            actionNode(node, mode)
        }
    }
}
