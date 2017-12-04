package smarthome

import (
    "encoding/json"
    "fmt"
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

    return influx.Query(query)
}

func GetUsageStat(deviceName string, periodValue int, periodBegin string, periodEnd string, intervalValue int) (UsageStat, error) {
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
