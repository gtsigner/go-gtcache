package cacher

import (
    "fmt"
    "github.com/go-redis/redis/v7"
    "testing"
)

func TestNewGtCache(t *testing.T) {
    var rds = NewGtCache(&redis.Options{
        Addr:     "127.0.0.1:6379",
        Password: "zhaojunlike",
        DB:       1,
    })
    rds.Client.Subscribe()
    err := rds.Client.Close()
    fmt.Println(err)
    err = rds.Client.Close()
    fmt.Println(err)
    res, err := rds.Client.Ping().Result()
    fmt.Println(res, err)
}
