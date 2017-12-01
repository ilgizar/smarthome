package main

import (
    "regexp"

    "github.com/ilgizar/smarthome/libs/smarthome"
)


const (
    defaultNodesConfig = "nodes.conf"
    defaultUsageConfig = "usage.conf"
)


var debug          bool
var cmdDebug       bool
var configFile     string
var sharedData     DataStruct
var config         ConfigStruct
var nodeConfig     smarthome.NodesConfigStruct
var usageConfig    smarthome.UsageConfigStruct
var variableRE     *regexp.Regexp
