package system

import (
    "bytes"
    "os"
    "runtime"
    "strconv"
    "time"
)

import p "path"


func GetGID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)
    return n
}

func Loop() {
    c := time.Tick(time.Second)
    for _ = range c {}
}

func GetAppName() string {
    return p.Base(os.Args[0])
}
