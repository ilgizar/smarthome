package files

import (
    "io/ioutil"
    "log"
    "os"

    "github.com/influxdata/toml"
    "github.com/spf13/viper"
)

func OpenFile(filename string) *os.File {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }

    return file
}

func ReadConfig(file string, path string, filetype string) map[string]interface{} {
    viper.SetConfigName(file)
    viper.SetConfigType(filetype)
    viper.AddConfigPath(path)

    err := viper.ReadInConfig()
    if err != nil {
        log.Println(err)
        return nil
    }

    return viper.AllSettings()
}

func ReadTypedConfig(filename string, config interface{}) error {
    f, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    buf, err := ioutil.ReadAll(f)
    if err != nil {
        return err
    }

    err = toml.Unmarshal(buf, config)

    return err
}
