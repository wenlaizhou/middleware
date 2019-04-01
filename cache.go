package middleware

import (
	"sync"
	"time"
)

// 缓存数据
type CacheData struct {
	Time time.Time
	Data interface{}
}

// 缓存对象
type Cache struct {
	Expire time.Duration
	Data   map[string]CacheData
	Lock   sync.RWMutex
}

// 创建新缓存
func NewCache(expire time.Duration) Cache {
	return Cache{
		Expire: expire,
		Data:   make(map[string]CacheData),
		Lock:   sync.RWMutex{},
	}
}

// 插入数据
func InsertData(cache Cache, key string, data interface{}) {
	t := time.Now().Add(cache.Expire)
	cache.Lock.Lock()
	defer cache.Lock.Unlock()
	cache.Data[key] = CacheData{
		Time: t,
		Data: data,
	}
}

// 获取缓存数据
func GetData(cache Cache, key string) interface{} {
	value, hasData := cache.Data[key]
	if !hasData {
		return nil
	}
	if time.Now().After(value.Time.Add(cache.Expire)) {
		cache.Lock.Lock()
		delete(cache.Data, key)
		cache.Lock.Unlock()
		return nil
	}
	return value.Data
}

// 清除过期数据
func CacheClean(cache Cache) {
	cache.Lock.Lock()
	for k, v := range cache.Data {
		if time.Now().After(v.Time.Add(cache.Expire)) {
			delete(cache.Data, k)
		}
	}
	cache.Lock.Unlock()
}
