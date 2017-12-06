package main

import (
    "log"
)


func checkUsage(state string, nodeName string) {
    if debug {
        log.Printf("checkUsage('%s', '%s')\n", state, nodeName)
    }

    if nodeName != "" {
        if nodeExists(nodeName) {
            setPreparedState(nodeName, "state", false)
        } else {
            return
        }
    } else {
        for n, _ := range sharedData.nodes {
            setPreparedState(n, "", false)
        }
    }

    for _, rule := range usageConfig.Rule {
        if !rule.Enabled {
            continue
        }

        if state == "" || state == "online" {
            for _, cond := range rule.Online {
                loopCondition(rule, cond, nodeName, "online")
            }
        }

        if state == "" || state == "offline" {
            for _, cond := range rule.Offline {
                loopCondition(rule, cond, nodeName, "offline")
            }
        }

        if state != "" {
            continue
        }

        for _, cond := range rule.Allowed {
            loopCondition(rule, cond, nodeName, "allowed")
        }

        for _, cond := range rule.Denied {
            loopCondition(rule, cond, nodeName, "denied")
        }

        for _, cond := range rule.Limited {
            loopCondition(rule, cond, nodeName, "limited")
        }
    }
}
