package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/ilgizar/smarthome/libs/influx"
    "github.com/influxdata/influxdb/client/v2"
)

type UsageStat struct {
    State bool
    Count int
    On    int
    Last  int
    Pause int
}

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

func parseTime(value string) (time.Time, error) {
    return time.Parse("02.01.2006 15:04:05", value)
}

func time2timestamp(value string, pattern string, def string) string {
    now := time.Now()
    _, z := now.Zone()

    if value != pattern {
        t, err := parseTime(value)
        if err != nil {
            t, err = time.Parse("15:04", value)
            if (err == nil) {
                t, err = parseTime(now.Format("02.01.2006") + " " + value + ":00")
            }
        }
        if err != nil {
            t, err = time.Parse("15:04:05", value)
            if (err == nil) {
                t, err = parseTime(now.Format("02.01.2006") + " " + value)
            }
        }
        if err == nil {
            value = strconv.FormatInt((t.Unix() - int64(z)) * 1000000000, 10)
        } else {
            value = pattern
        }
    }

    if value == pattern {
        value = def
    }

    return value
}

func getDataFromDB(deviceName string, periodBegin string, periodEnd string) ([]client.Result, error) {
    query := fmt.Sprintf(
        `SELECT mean("percent_packet_loss") FROM "ping" WHERE ("url" =~ /^%s$/) AND time > %s AND time < %s GROUP BY time(1m) fill(100)`,
        deviceName, periodBegin, periodEnd)
    if debug {
        log.Println(query)
    }

    return influx.QueryDB(query)
}

func getUsageStat() (UsageStat, error) {
    var err error
    var stat UsageStat

    periodEnd = time2timestamp(periodEnd, "now()", "now()")
    periodBegin = time2timestamp(periodBegin, "", fmt.Sprintf(`%s - %d%s`, periodEnd, periodValue, "m"))

    res, err := getDataFromDB(deviceName, periodBegin, periodEnd)
    if err != nil {
        return stat, err
    }

    count := 0
    switchOn := true
    var begin  int64   = 0
    var end    int64   = 0
    var pause  float64 = 0
    var entire float64 = 0
    var on     float64 = 0
    var value  float64
    var t      time.Time

    if intervalValue > 0 {
        t = time.Now()
        end = t.Unix()
        begin =  end - int64(intervalValue) * 60
    }

    for i, row := range res[0].Series[0].Values {
        if (i > 0) {
            t, err = time.Parse(time.RFC3339, row[0].(string))
            if err == nil {
                value, err = row[1].(json.Number).Float64()
                if err == nil {
                    value = 1 - value / 100
                    entire = entire + value

                    if debug {
                        fmt.Printf("[%s] %f\n", t, value)
                    }

                    if (intervalValue == 0 || (t.Unix() >= begin && t.Unix() <= end)) {
                        if value == 0 {
                            pause++
                        } else {
                            on = on + value
                        }
                    }

                    if (intervalValue == 0) {
                        if value == 0 {
                            on = 0
                        } else {
                            pause = 0
                        }
                    }

                    if value == 0 {
                        switchOn = true
                    } else if (switchOn) {
                        count++
                        switchOn = false
                    }
                }
            }
        }
    }

    stat = UsageStat {
        State: !switchOn,
        Count: int(count),
        On:    int(entire),
        Last:  int(on),
        Pause: int(pause),
    }

    return stat, nil
}

func out(stat UsageStat) {
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

    influx.ConnectDB(InfluxHost, InfluxUser, InfluxPasswd, InfluxDB)
    stat, err := getUsageStat()
    if (err != nil) {
        log.Fatal(err)
    }

    out(stat)
}
