package main

import (
    "flag"
    "log"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/ilgizar/smarthome/libs/mqtt"
    "github.com/ilgizar/smarthome/libs/smarthome"
)

var dateRangeLimit int
var subscribeDelay int
var debug          bool
var initialized    bool
var mqttHost       string
var mqttRoot       string
var dateFormat     string
var stateRefreshed time.Time
var weekEnd        string
var weekEndDays    []string

type Map struct {
    sync.Mutex
    types          map[string]bool
    exists         map[string]bool
    dates          map[string]interface{}
}

var sharedData     Map
var types          map[string]bool
var exists         map[string]bool
var dates          map[string]interface{}

func init() {
    flag.BoolVar(&debug,         "debug",   false,             "debug mode")
    flag.StringVar(&mqttHost,    "mqtt",    "localhost:1883",  "MQTT server, can set without port number")
    flag.StringVar(&mqttRoot,    "root",    "smarthome",       "MQTT root section")
    flag.StringVar(&dateFormat,  "format",  "02.01.2006",      "date format")
    flag.StringVar(&weekEnd,     "weekend", "Saturday,Sunday", "set weekend days separated commas of list: Monday,Tuesday,Wednesday,Thursday,Friday,Saturday,Sunday")
    flag.IntVar(&dateRangeLimit, "limit",   365,               "date range limit by days")
    flag.IntVar(&subscribeDelay, "delay",   1,                 "delay between MQTT subscribes on starting time by seconds")

    sharedData = Map{
        types:  make(map[string]bool),
        exists: make(map[string]bool),
        dates:  make(map[string]interface{}),
    }

    for _, t := range smarthome.TypesOfDays {
        sharedData.types[t] = false
        sharedData.exists[t] = false
        sharedData.dates[t] = []time.Time{}
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
    for _, d := range sharedData.dates[t].([]time.Time) {
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

func publicState(t string) {
    if !initialized {
        sharedData.Lock()
        sharedData.exists[t] = true
        sharedData.Unlock()
    }

    val := "false"
    if sharedData.types[t] {
        val = "true"
    }

    if debug {
        log.Printf("Publish state '%s' to %s\n", t, val)
    }

    mqtt.Publish(mqttRoot + "/calendar/" + t + "/state", val, true)
}

func checkTodayByType(t string) {
    state := checkDateType(t)

    if state != sharedData.types[t] {
        sharedData.Lock()
        sharedData.types[t] = state
        sharedData.Unlock()
        publicState(t)
    }
}

func checkToday() {
    for _, t := range smarthome.TypesOfDays {
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


    mqtt.Subscribe(mqttRoot + "/calendar/+/state", func(topic, message []byte) {
        parts := strings.Split(string(topic), "/")
        section := parts[len(parts) - 2]
        if _, ok:= sharedData.types[section]; ok {
            if !initialized {
                sharedData.Lock()
                sharedData.exists[section] = true
                sharedData.Unlock()
            }
            state, err := strconv.ParseBool(string(message))
            if (err == nil && sharedData.types[section] != state) {
                sharedData.Lock()
                sharedData.types[section] = state
                sharedData.Unlock()
                if debug {
                    msg := "Change"
                    if !initialized {
                        msg = "Init"
                    }
                    log.Printf("%s state '%s' to %v\n", msg, section, state)
                }
            }
        }
    })

    time.Sleep(time.Duration(subscribeDelay) * time.Second)

    mqtt.Subscribe(mqttRoot + "/calendar/+/days", func(topic, message []byte) {
        parts := strings.Split(string(topic), "/")
        section := parts[len(parts) - 2]
        if _, ok:= sharedData.types[section]; ok {
            sharedData.Lock()
            sharedData.dates[section] = forgetPast(getDateArray(string2map(string(message))))
            sharedData.Unlock()
            checkTodayByType(section)
        }
    })

    time.Sleep(time.Duration(subscribeDelay) * time.Second)

    initialized = true

    for _, t := range smarthome.TypesOfDays {
        if !sharedData.exists[t] {
            publicState(t)
        }
    }

    stateRefreshed = time.Now().Add(-1 * time.Hour)
    c := time.Tick(time.Second)
    for now := range c {
        if checkMidnight(now) || checkMoreThanDay(now) {
            stateRefreshed = now
            checkToday()
        }
    }
}
