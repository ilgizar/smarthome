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
        ClientID: []byte("smarthome-pushconfigs"),
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

func Clear(section string) {
    subscribeTopic := []byte(section)

    err := cli.Subscribe(&client.SubscribeOptions{
        SubReqs: []*client.SubReq{
            &client.SubReq{
                TopicFilter: subscribeTopic,
                Handler: func(topicName, message []byte) {
                    if (string(message) != "") {
                        Publish(string(topicName), "", true)
                    }
                },
            },
        },
    })

    if err != nil {
        log.Println(err)
    }

    time.Sleep(1000 * time.Millisecond)

    err = cli.Unsubscribe(&client.UnsubscribeOptions{
        TopicFilters: [][]byte{
            subscribeTopic,
        },
    })
    if err != nil {
        log.Println(err)
    }

    time.Sleep(1000 * time.Millisecond)
}
