package main

import (
    "strconv"
    "strings"
    "time"

    "github.com/ilgizar/smarthome/libs/system"
)


func parseDateTime(value string) (time.Time, error) {
    return time.Parse("02.01.2006 15:04:05", value)
}

func dateToTimestamp(value string, dawn bool) int {
    time := "00:00:00"
    if !dawn {
        time = "23:59:59"
    }
    t, err := parseDateTime(value + " " + time)
    if err == nil {
        return int(t.Unix())
    }

    return -1
}

func weekdayToInt(weekday string) int {
    for i, wd := range system.Days {
        if wd == weekday {
            return i
        }
    }

    return -1
}

func timeToMinutes(time string) int {
    parts := strings.Split(time, ":")
    hour, err := strconv.ParseInt(parts[0], 10, 64)
    if err == nil && hour >=0 && hour <= 23 {
        min, err := strconv.ParseInt(parts[1], 10, 64)
        if err == nil && min >= 0 && min <= 59 {
            return int(hour * 60 + min)
        }
    }

    return -1
}
