package influx

import (
    "log"

    "github.com/influxdata/influxdb/client/v2"
)

var influxDB string
var influxClient client.Client
var influxBP client.BatchPoints


func Connect(InfluxHost string, InfluxUser string, InfluxPass string, InfluxDB string) {
    var err error
    influxClient, err = client.NewHTTPClient(client.HTTPConfig{
        Addr:     InfluxHost,
        Username: InfluxUser,
        Password: InfluxPass,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer influxClient.Close()

    influxDB = InfluxDB
    influxBP, err = client.NewBatchPoints(client.BatchPointsConfig{
        Database:  InfluxDB,
    })
    if err != nil {
        log.Fatal(err)
    }
}

func Disconnect() {
    influxClient.Close()
}

func Query(cmd string) (res []client.Result, err error) {
    q := client.Query{
        Command:  cmd,
        Database: influxDB,
    }

    if response, err := influxClient.Query(q); err == nil {
        if response.Error() != nil {
            return res, response.Error()
        }
        res = response.Results
    } else {
        return res, err
    }

    return res, nil
}