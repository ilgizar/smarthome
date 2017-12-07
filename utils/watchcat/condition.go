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

    var cond smarthome.UsageConfigConditionStruct
    switch reflect.TypeOf(c).String() {
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

func enableCheckLimited(node NodeStruct) bool {
    log.Printf("enableCheckLimited: %+v", node)
    if node.modes["permit"].prepared {
        return false
    }

    if node.modes["state"].state == "online" && node.modes["permit"].eventtime["limited"] < node.modes["state"].eventtime["online"] {
        return true
    }

    now := time.Now()
    return now.Second() == 0 || getDeltaTimestamp(node.modes["permit"].eventtime["limited"]) > 60
}

func setEventTime(nodeName, modeName string) {
    sharedData.Lock()
    mode := sharedData.nodes[nodeName].modes[modeName]
    mode.eventtime[mode.state] = int(time.Now().Unix())
    sharedData.nodes[nodeName].modes[modeName] = mode
    sharedData.Unlock()
}

func getLimitedAction(nodeName string, cond smarthome.UsageConfigLimitedStruct) string {
    node := sharedData.nodes[nodeName]
    if !enableCheckLimited(node)  {
        return ""
    }

    setEventTime(nodeName, "permit")

    stat, err := smarthome.GetUsageStat(nodeName, cond.Period, cond.Begin, cond.End, cond.Using + cond.Pause)
    log.Printf("%+v %+v", stat, err)
    if err != nil {
        return ""
    }

    if node.modes["permit"].event == "block" || node.modes["permit"].event == "begin" {
        if stat.Pause >= cond.Pause && stat.On < cond.Overall {
            return "open"
        }
    }

    if node.modes["permit"].event == "open" || node.modes["permit"].event == "begin" {
        if stat.Last >= cond.Using || stat.On >= cond.Overall {
            return "block"
        }
    }

    return ""
}

func loopCondition(rule smarthome.UsageConfigRuleStruct, c interface{}, nodeName string, state string) {
    modeName := getModeNameByState(state)

    cond, ok := getConditionStruct(c)
    if !ok {
        return
    }

    ok = checkCondition(c)
    if !(ok || (modeName == "permit" && state != "limited")) {
        return
    }

    log.Printf("loopCondition: state(%s)\n", state)
    for _, n := range rule.Nodes {
        mode := sharedData.nodes[n].modes[modeName]
        if (nodeName != "" && nodeName != n) || !nodeExists(n) || mode.prepared {
            continue
        }

        log.Printf("loopCondition: node(%s) mode(%s)\n", n, modeName)
        if modeName == "permit" {
            var st string
            if ok {
                st = state
            } else if mode.state == state {
                st = "limited"
            } else {
                continue
            }

            if checkChangeState(n, st) {
                actionNode(n, modeName, "end")
                if ok {
                    setNodeState(n, st)
                }
            }
        }

        if checkNodeState(n, cond, state) {
            initNodeActions(n, cond, state)
        }

        node := sharedData.nodes[n]
        if node.active {
            actionNode(n, modeName, "begin")
            switch reflect.TypeOf(c).String() {
                case "smarthome.UsageConfigLimitedStruct":
                    if act := getLimitedAction(n, c.(smarthome.UsageConfigLimitedStruct)); act != "" {
                        actionNode(n, modeName, act)
                    }
            }
        }

        if ok {
            setPreparedState(n, modeName, true)
        }
    }
}
