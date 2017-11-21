package files

import (
    "log"
    "os"

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
