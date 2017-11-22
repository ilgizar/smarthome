package main

import (
    "flag"
    "log"
    "regexp"
    "strings"
    "time"

    "github.com/ilgizar/smarthome/libs/mqtt"
)

var types = map[string]string{
    "leave":    "false",        // отпуск
    "vacation": "false",        // каникулы
    "holiday":  "false",        // праздник
    "workday":  "false",        // рабочий день
}

var dateRangeLimit int
var subscribeDelay int
var debug          bool
var mqttHost       string
var mqttRoot       string
var dateFormat     string
var stateRefreshed time.Time
var weekEnd        string
var weekEndDays    []string

var dates          map[string]interface{}


func init() {
    flag.BoolVar(&debug,         "debug",   false,             "debug mode")
    flag.StringVar(&mqttHost,    "mqtt",    "localhost:1883",  "MQTT server, can set without port number")
    flag.StringVar(&mqttRoot,    "root",    "smarthome",       "MQTT root section")
    flag.StringVar(&dateFormat,  "format",  "02.01.2006",      "date format")
    flag.StringVar(&weekEnd,     "weekend", "Saturday,Sunday", "set weekend days separated commas of list: Monday,Tuesday,Wednesday,Thursday,Friday,Saturday,Sunday")
    flag.IntVar(&dateRangeLimit, "limit",   365,               "date range limit by days")
    flag.IntVar(&subscribeDelay, "delay",   1,                 "delay between MQTT subscribes on starting time by seconds")

    dates = make(map[string]interface{})
    for t, _ := range types {
        dates[t] = []time.Time{}
    }
}

func string2map(value string) ([]string) {
    clearRE := regexp.MustCompile(`(^\s*\[\s*"\s*|\s*(")\s*(,)\s*(")\s*|\s*"\s*\]\s*$)`)

    value = clearRE.ReplaceAllString(value, "$2$3$4")

    return strings.Split(value, "\",\"")
}

func getDateRange(begin, end string) []time.Time {
    result := []time.Time{}

    start, err := time.Parse(dateFormat, begin)
    if (err == nil) {
        stop, err := time.Parse(dateFormat, end)
        if (err == nil) {
            stop = stop.Add(24 * time.Hour)
            for ((start != stop) && (len(result) < dateRangeLimit)) {
                result = append(result, start)
                start = start.Add(24 * time.Hour)
            }
            if len(result) == dateRangeLimit {
                result = []time.Time{}
            }
        }
    }

    return result
}

func getDateArray(array []string) ([]time.Time) {
    result := []time.Time{}

    dateRE := regexp.MustCompile(`^\s*(\d{2}\.\d{2}\.\d{4})\s*$`)
    rangeRE := regexp.MustCompile(`^\s*(\d{2}\.\d{2}\.\d{4})\s*-\s*(\d{2}\.\d{2}\.\d{4})\s*$`)
    for _, v := range array {
        if dateRE.MatchString(v) {
            t, err := time.Parse(dateFormat, v)
            if err == nil {
                result = append(result, t)
            }
        } else if rangeRE.MatchString(v) {
            borders := rangeRE.FindStringSubmatch(v)
            result = append(result, getDateRange(borders[1], borders[2])...)
        }
    }

    return result
}

func forgetPast(list []time.Time) ([]time.Time) {
    result := []time.Time{}

    now := time.Now()
    now = now.Truncate(24 * time.Hour)
    for _, t := range list {
        if (t.After(now) || t.Equal(now)) {
            result = append(result, t)
        }
    }

    return result
}

func checkDateType(t string) bool {
    res := false

    now := time.Now()
    now = now.Truncate(24 * time.Hour)
    for _, d := range dates[t].([]time.Time) {
        if d.Equal(now) {
            res = true
            break
        }
    }

    if (!res && (t == "workday")) {
        weekday := now.Weekday().String()
        res = true
        for _, w := range weekEndDays {
            if w == weekday {
                res = false
                break
            }
        }
    }

    return res
}

func checkTodayByType(t string) {
    state := "false"
    if checkDateType(t) {
        state = "true"
    }
    if state != types[t] {
        types[t] = state
        mqtt.Publish(mqttRoot + "/calendar/" + t + "/state", state, true)
    }
}

func checkToday() {
    for t, _ := range types {
        checkTodayByType(t)
    }
}

func checkMidnight(now time.Time) bool {
    if (now.Hour() == 0) {
        return now.Sub(stateRefreshed) > time.Hour
    }

    return false
}

func checkMoreThanDay(now time.Time) bool {
    return now.Sub(stateRefreshed) > 24 * time.Hour
}

func main() {
    flag.Parse()

    if debug {
        log.Printf("Started")
    }

    weekEndDays = strings.Split(weekEnd, ",")

    mqtt.Connect(mqttHost)

    initialized := false

    mqtt.Subscribe(mqttRoot + "/calendar/+/state", func(topic, message []byte) {
        parts := strings.Split(string(topic), "/")
        section := parts[len(parts) - 2]
        if _, ok:= types[section]; ok {
            state := string(message)
            if (types[section] != state) {
                types[section] = state
                if debug {
                    msg := "Change"
                    if !initialized {
                        msg = "Init"
                    }
                    log.Printf("%s state '%s' to %v\n", msg, section, types[section])
                }
            }
        }
    })

    time.Sleep(time.Duration(subscribeDelay) * time.Second)

    initialized = true

    mqtt.Subscribe(mqttRoot + "/calendar/+/days", func(topic, message []byte) {
        parts := strings.Split(string(topic), "/")
        section := parts[len(parts) - 2]
        if _, ok:= types[section]; ok {
            dates[section] = forgetPast(getDateArray(string2map(string(message))))
            checkTodayByType(section)
        }
    })

    stateRefreshed = time.Now().Add(-1 * time.Hour)
    c := time.Tick(time.Second)
    for now := range c {
        if checkMidnight(now) || checkMoreThanDay(now) {
            stateRefreshed = now
            checkToday()
        }
    }
}
