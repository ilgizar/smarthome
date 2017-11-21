package main

import (
    "flag"
    "fmt"
    "os"
    "reflect"
    "strings"
    "time"

    "github.com/ilgizar/smarthome/libs/files"
    "github.com/ilgizar/smarthome/libs/mqtt"
)

var mqttHost string
var mqttRoot string
var path     string
var configs  string
var debug    bool
var format   string


func init() {
    flag.BoolVar(&debug,      "debug",  false,            "debug mode")
    flag.StringVar(&path,     "path",   ".",              "path to config directory")
    flag.StringVar(&mqttHost, "mqtt",   "localhost:1883", "MQTT server")
    flag.StringVar(&mqttRoot, "root",   "smarthome",      "MQTT root section")
    flag.StringVar(&configs,  "config", "",               "list of config files without extensions separated commas")
    flag.StringVar(&format,   "format", "toml",           "config file format")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s -config <list of configs> [options]\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
    }
}

func convertConfig(config string, data map[string]interface{}) (map[string]string) {
    result := map[string]string{}

    var message string
    for node, v := range data {
        for key, value := range v.(map[string]interface{}) {
            k := fmt.Sprintf("%s/%s/%s/%s", mqttRoot, config, node, key)
            typeName := reflect.TypeOf(value).String()
            message = ""
            if typeName == "string" {
                message = value.(string)
            } else {
                items := []string{}
                for _, item := range value.([]interface{}) {
                    items = append(items, fmt.Sprintf("\"%s\"", item))
                }
                message = "[" + strings.Join(items, ",") + "]"
            }
            result[k] = message
        }
    }

    return result
}

func exportToMQTT(list map[string]string) {
    for topic, msg := range list {
        mqtt.Publish(topic, msg, true)
        if (debug) {
            fmt.Printf("Send MQTT message: %36s %s\n", topic, msg)
        }
    }
}

func main() {
    flag.Parse()

    if (configs == "") {
        flag.Usage()
        return
    }

    mqtt.Connect(mqttHost)

    for _, file := range strings.Split(configs, ",") {
        data := files.ReadConfig(file, path, format)
        list := convertConfig(file, data)
        mqtt.Clear(mqttRoot + "/" + file + "/#", 1)
        exportToMQTT(list)
    }

    time.Sleep(1000 * time.Millisecond)
}
