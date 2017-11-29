package mqtt

import (
    "log"
    "regexp"
    "time"

    "github.com/ilgizar/smarthome/libs/system"
    "github.com/yosssi/gmq/mqtt/client"
)

var cli *client.Client


func Connect(host string) *client.Client {
    cli = client.New(&client.Options{
        ErrorHandler: func(err error) {
            log.Printf("MQTT client error: %s\n", err)
        },
    })
    defer cli.Terminate()

    addressRE := regexp.MustCompile(`:\d{1,5}$`)
    if !addressRE.MatchString(host) {
        host = host + ":1883"
    }

    err := cli.Connect(&client.ConnectOptions{
        Network:  "tcp",
        Address:  host,
        ClientID: []byte("smarthome-" + system.GetAppName()),
    })

    if err != nil {
        log.Fatal(err)
    }

    return cli
}

func Disconnect() {
    if cli != nil {
        if err := cli.Disconnect(); err != nil {
            log.Fatal(err)
        }
    }
}

func Publish(topic string, message string, retain bool) {
    err := cli.Publish(&client.PublishOptions{
        TopicName: []byte(topic),
        Message:   []byte(message),
        Retain:    retain,
    })

    if err != nil {
        log.Println(err)
    }
}

func Subscribe(topic string, handler client.MessageHandler) error {
    err := cli.Subscribe(&client.SubscribeOptions{
        SubReqs: []*client.SubReq{
            &client.SubReq{
                TopicFilter: []byte(topic),
                Handler: handler,
            },
        },
    })

    if err != nil {
        log.Println(err)
    }

    return err
}

func Unsubscribe(topic string) error {
    err := cli.Unsubscribe(&client.UnsubscribeOptions{
        TopicFilters: [][]byte{
            []byte(topic),
        },
    })

    if err != nil {
        log.Println(err)
    }

    return err
}

func Clear(topic string, delay time.Duration) {
    Subscribe(topic, func(topicName, message []byte) {
        if (string(message) != "") {
            Publish(string(topicName), "", true)
        }
    })
    time.Sleep(delay * time.Second)

    Unsubscribe(topic)
    time.Sleep(delay * time.Second)
}
