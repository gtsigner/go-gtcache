package cacher

import (
    "errors"
    "fmt"
    "github.com/go-redis/redis/v7"
    "log"
    "reflect"
    "sync"
    "time"
    "zhaojunlike/common"
)

func NewGtRedisConn(opt *redis.Options) *redis.Client {
    client := redis.NewClient(opt)
    return client
}

type GtCache struct {
    Client *redis.Client
    lock   sync.Mutex
}

func NewGtCache(opt *redis.Options) *GtCache {
    var client = NewGtRedisConn(opt)
    _, err := client.Ping().Result()
    if err != nil {
        log.Fatalln("连接redis失败:", err)
        return nil
    }
    var cache = &GtCache{
        Client: client,
    }
    return cache
}

func (cache *GtCache) RPush(key string, data interface{}) error {
    var str, err = common.JSONStringify(data)
    if err != nil {
        return err
    }
    err = cache.Client.RPush(key, str).Err()
    return err
}

func (cache *GtCache) LPop(key string, v interface{}) error {
    res, err := cache.Client.LPop(key).Result()
    if err != nil {
        return err
    }
    err = common.JSONParse(res, v)
    if err != nil {
        return err
    }
    return nil
}

//设置JSON
func (cache *GtCache) SetJson(key string, data interface{}, expiration time.Duration) error {
    var str, err = common.JSONStringify(data)
    if err != nil {
        return err
    }
    err = cache.Client.Set(key, str, expiration).Err()
    return err
}
func (cache *GtCache) SetJsonNX(key string, data interface{}, expiration time.Duration) (bool, error) {
    var str, err = common.JSONStringify(data)
    if err != nil {
        return false, err
    }
    return cache.Client.SetNX(key, str, expiration).Result()
}

//获取JSON并且解析成json对象 v 必须是一个指针
func (cache *GtCache) GetJsonObj(key string, v interface{}) error {
    res, err := cache.Client.Get(key).Result()
    if err != nil {
        return err
    }
    err = common.JSONParse(res, v)
    if err != nil {
        return err
    }
    return nil
}

//设置JSON
func (cache *GtCache) HSetJson(key string, field string, data interface{}) error {
    var str, err = common.JSONStringify(data)
    if err != nil {
        return err
    }
    err = cache.Client.HSet(key, field, str).Err()
    return err
}

//如果不存在设置JSON
func (cache *GtCache) HSetJsonNX(key string, field string, data interface{}) (bool, error) {
    var str, err = common.JSONStringify(data)
    if err != nil {
        return false, err
    }
    return cache.Client.HSetNX(key, field, str).Result()
}

//获取JSON并且解析成json对象 v 必须是一个指针
func (cache *GtCache) HGetJsonObj(key string, field string, v interface{}) error {
    res, err := cache.Client.HGet(key, field).Result()
    if err != nil {
        return err
    }
    err = common.JSONParse(res, v)
    if err != nil {
        return err
    }
    return nil
}

//log
func (cache *GtCache) Log(tag string, args ...interface{}) {
    if tag == "" {
        return
    }
    var str = "[" + common.TimeString() + "][" + tag + "]"
    var dstr = common.DateString() + ":" + tag
    for _, v := range args {
        str += fmt.Sprintf("%v", v)
    }
    _ = cache.Client.LPush(dstr, str)
}

func (cache *GtCache) Publish(channel string, data interface{}) *redis.IntCmd {
    var tp = reflect.TypeOf(data).String()
    if tp == "string" {
        return cache.Client.Publish(channel, data.(string))
    }
    var js, _ = common.JSONStringify(data)
    return cache.Client.Publish(channel, js)
}

//获取redis锁
func (cache *GtCache) GetLock(key string, timeout time.Duration) (bool, error) {
    var t = timeout
    var lockKey = "locker:" + key
    for {
        var ok, _ = cache.Client.SetNX(lockKey, 1, timeout).Result()
        if ok {
            return true, nil
        }
        time.Sleep(time.Second)
        t -= time.Second
        if t <= 0 {
            break
        }
    }
    return false, errors.New("get lock timeout")
}
func (cache *GtCache) FreeLock(key string) error {
    var lockKey = "locker:" + key
    var _, err = cache.Client.Del(lockKey).Result()
    return err
}
