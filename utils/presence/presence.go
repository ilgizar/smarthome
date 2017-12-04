package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/ilgizar/smarthome/libs/influx"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var deviceName    string
var periodValue   int
var periodBegin   string
var periodEnd     string
var intervalValue int
var debug         bool
var outputFormat  string
var InfluxHost    string
var InfluxDB      string
var InfluxUser    string
var InfluxPasswd  string


func init() {
    flag.StringVar(&deviceName,   "device",   "localhost",             "device name")
    flag.IntVar(&periodValue,     "period",   60,                      "period value in minutes")
    flag.StringVar(&periodBegin,  "begin",    "",                      "period begin (in GMT). format: YYYY-MM-DDThh:mm:ss.msZ (default <end> - <period>)")
    flag.StringVar(&periodEnd,    "end",      "now()",                 "period end (in GMT)")
    flag.IntVar(&intervalValue,   "interval", 0,                       "calculating interval in minutes")
    flag.StringVar(&outputFormat, "output",   "json",                  "output format: json, text")
    flag.BoolVar(&debug,          "debug",    false,                   "debug mode")
    flag.StringVar(&InfluxHost,   "host",     "http://127.0.0.1:8086", "InfluxDB server address")
    flag.StringVar(&InfluxDB,     "db",       "telegraf",              "InfluxDB database name")
    flag.StringVar(&InfluxUser,   "user",     "",                      "InfluxDB account")
    flag.StringVar(&InfluxPasswd, "password", "telegraf",              "InfluxDB password")
}

func out(stat smarthome.UsageStat) {
    if (outputFormat == "json") {
        b, err := json.Marshal(stat)
        if err != nil {
            log.Fatal(err)
        }
        os.Stdout.Write(b)
    } else if (outputFormat == "text") {
        state := 0
        if stat.State {
            state = 1
        }
        fmt.Printf("%d %d %d %d %d", state, stat.On, stat.Last, stat.Pause, stat.Count)
    } else {
        log.Fatal(fmt.Sprintf("Unknown output format: %s\n", outputFormat))
    }
}

func main() {
    flag.Parse()

    influx.Connect(InfluxHost, InfluxUser, InfluxPasswd, InfluxDB)
    stat, err := smarthome.GetUsageStat(deviceName, periodValue, periodBegin, periodEnd, intervalValue)
    if (err != nil) {
        log.Fatal(err)
    }

    out(stat)
}
