package main

import (
    "fmt"
    "log"
    "reflect"
    "time"

    "github.com/ilgizar/smarthome/libs/smarthome"
)


func getConditionStruct(c interface{}) (smarthome.UsageConfigConditionStruct, bool) {
    ok := true

    t := reflect.TypeOf(c).String()
    var cond smarthome.UsageConfigConditionStruct
    switch t {
        case "smarthome.UsageConfigConditionStruct":
            cond = c.(smarthome.UsageConfigConditionStruct)
        case "smarthome.UsageConfigLimitedStruct":
            cond = c.(smarthome.UsageConfigLimitedStruct).UsageConfigConditionStruct
        default:
            ok = false
    }

    return cond, ok
}

func checkCondition(c interface{}) bool {
    cond, ok := getConditionStruct(c)
    if !ok {
        return false
    }

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

func loopCondition(rule smarthome.UsageConfigRuleStruct, c interface{}, nodeName string, state string) {
    mode := getModeNameByState(state)

    cond, ok := getConditionStruct(c)
    if !ok {
        return
    }

    ok = checkCondition(c)
    if !(ok || (mode == "permit" && state != "limited")) {
        return
    }

    log.Printf("loopCondition: state(%s)\n", state)
    for _, node := range rule.Nodes {
        if (nodeName != "" && nodeName != node) || !nodeExists(node) || sharedData.nodes[node].modes[mode].prepared {
            continue
        }

        log.Printf("loopCondition: node(%s) mode(%s)\n", node, mode)
        if mode == "permit" {
            var st string
            if ok {
                st = state
            } else if sharedData.nodes[node].modes[mode].state == state {
                st = "limited"
            } else {
                continue
            }

            if checkChangeState(node, st) {
                actionNode(node, mode, "end")
                if ok {
                    setNodeState(node, st)
                }
            }
        }

        if checkNodeState(node, cond, state) {
            initNodeActions(node, cond, state)
        }

        if sharedData.nodes[node].active {
            actionNode(node, mode, "begin")
        }

        if ok {
            setPreparedState(node, mode, true)
        }
    }
}
