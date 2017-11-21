package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "reflect"
    "strings"
    "time"

    "github.com/spf13/viper"
    "github.com/yosssi/gmq/mqtt/client"
)

var mqttHost string
var mqttRoot string
var path     string
var configs  string
var debug    bool


func init() {
    flag.BoolVar(&debug,      "debug",  false,            "debug mode")
    flag.StringVar(&path,     "path",   ".",              "path to config directory")
    flag.StringVar(&mqttHost, "mqtt",   "localhost:1883", "MQTT server")
    flag.StringVar(&mqttRoot, "root",   "smarthome",      "MQTT root section")
    flag.StringVar(&configs,  "config", "",               "list of config files without extensions separated commas")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s -config <list of configs> [options]\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
    }
}

func readConfig(file string) map[string]interface{} {
    viper.SetConfigName(file)
    viper.SetConfigType("toml")
    viper.AddConfigPath(path)

    err := viper.ReadInConfig()
    if err != nil {
        log.Println(err)
        return nil
    }

    return viper.AllSettings()
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

func connectMQTT() *client.Client {
    cli := client.New(&client.Options{
        ErrorHandler: func(err error) {
            log.Println(err)
        },
    })
    defer cli.Terminate()

    err := cli.Connect(&client.ConnectOptions{
        Network:  "tcp",
        Address:  mqttHost,
        ClientID: []byte("smarthome-pushconfigs"),
    })
    if err != nil {
        log.Fatal(err)
    }

    return cli
}

func showMQTTvalue(k, v string) string {
    return fmt.Sprintf("%36s %s", k, v)
}

func sendMQTT(cli *client.Client, topic string, message string) {
    err := cli.Publish(&client.PublishOptions{
        TopicName: []byte(topic),
        Message:   []byte(message),
        Retain:    true,
    })

    if err != nil {
        log.Println(err)
    }

    if (debug) {
        if (message == "") {
            fmt.Println("Remove MQTT message: " + showMQTTvalue(topic, message))
        } else {
            fmt.Println("Send MQTT message: " + showMQTTvalue(topic, message))
        }
    }

}

func sleep(seconds time.Duration) {
    time.Sleep(seconds * 1000 * time.Millisecond)
}

func exportToMQTT(cli *client.Client, list map[string]string) {
    for k, v := range list {
        sendMQTT(cli, k, v)
    }

    sleep(1)
}

func clearMQTT(cli *client.Client, section string) {
    topic := []byte(mqttRoot + "/" + section + "/#")

    err := cli.Subscribe(&client.SubscribeOptions{
        SubReqs: []*client.SubReq{
            &client.SubReq{
                TopicFilter: topic,
                Handler: func(topicName, message []byte) {
                    if (string(message) != "") {
                        sendMQTT(cli, string(topicName), "")
                    }
                },
            },
        },
    })

    if err != nil {
        log.Println(err)
    }

    sleep(1)

    err = cli.Unsubscribe(&client.UnsubscribeOptions{
        TopicFilters: [][]byte{
            topic,
        },
    })
    if err != nil {
        log.Println(err)
    }

    sleep(1)
}

func main() {
    flag.Parse()

    if (configs == "") {
        flag.Usage()
        return
    }

    cli := connectMQTT()

    for _, file := range strings.Split(configs, ",") {
        data := readConfig(file)
        list := convertConfig(file, data)
        clearMQTT(cli, file)
        exportToMQTT(cli, list)
    }
}
