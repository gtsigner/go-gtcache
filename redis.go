package cacher

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
	"zhaojunlike/common"
)

func NewRedisConn() redis.Conn {
	var opt = redis.DialPassword("zhaojunlike")
	c, err := redis.Dial("tcp", "127.0.0.1:6379", opt)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return nil
	}
	return c
}

type HkCache struct {
	Conn redis.Conn
	lock sync.Mutex
}

func NewHkCache() *HkCache {
	var conn = NewRedisConn()
	var cache = &HkCache{
		Conn: conn,
	}
	return cache
}

//设置JSON
func (cache *HkCache) SetJson(key string, data interface{}) error {
	var str, err = common.JSONStringify(data)
	if err != nil {
		return err
	}
	_, err = cache.ConDo("SET", key, str)
	return err
}

//获取JSON并且解析成json对象
func (cache *HkCache) GetJsonObj(key string, v interface{}) error {
	res, err := redis.String(cache.ConDo("GET", key))
	if err != nil {
		return err
	}
	err = common.JSONParse(res, v)
	if err != nil {
		return err
	}
	return nil
}

//并发执行
func (cache *HkCache) ConDo(commandName string, args ...interface{}) (reply interface{}, err error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	reply, err = cache.Conn.Do(commandName, args...)
	return reply, err
}

func (cache *HkCache) Log(tag string, args ...interface{}) {
	if tag == "" {
		return
	}
	var str = "[" + common.TimeString() + "][" + tag + "]"
	var dstr = common.DateString() + ":" + tag
	for _, v := range args {
		str += fmt.Sprintf("%v", v)
	}
	_, _ = cache.ConDo("lpush", dstr, str)
}
