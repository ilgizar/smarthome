package mqtt

import (
    "log"
    "time"

    "github.com/yosssi/gmq/mqtt/client"
)

var cli *client.Client


func Connect(host string) *client.Client {
    cli = client.New(&client.Options{
        ErrorHandler: func(err error) {
            log.Println(err)
        },
    })
    defer cli.Terminate()

    err := cli.Connect(&client.ConnectOptions{
        Network:  "tcp",
        Address:  host,
        ClientID: []byte("smarthome-go"),
    })
    if err != nil {
        log.Fatal(err)
    }

    return cli
}

func Publish(topic string, message string, retain bool) {
    err := cli.Publish(&client.PublishOptions{
        TopicName: []byte(topic),
        Message:   []byte(message),
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
