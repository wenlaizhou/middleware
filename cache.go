package middleware

import (
	"sync"
	"time"
)

// CacheData 缓存数据
type CacheData struct {
	Time time.Time
	Data interface{}
}

// Cache 缓存对象
type Cache struct {
	Expire time.Duration
	Data   map[string]CacheData
	Lock   sync.RWMutex
}

// NewCache 创建新缓存
func NewCache(expire time.Duration) *Cache {
	res := &Cache{
		Expire: expire,
		Data:   make(map[string]CacheData),
		Lock:   sync.RWMutex{},
	}
	go func(cache *Cache) {
		for {
			time.Sleep(cache.Expire)
			cache.CacheClean()
		}
	}(res)
	return res
}

// InsertData 插入数据
func (cache *Cache) InsertData(key string, data interface{}) {
	t := time.Now().Add(cache.Expire)
	cache.Lock.Lock()
	defer cache.Lock.Unlock()
	cache.Data[key] = CacheData{
		Time: t,
		Data: data,
	}
}

// GetData 获取缓存数据
func (cache *Cache) GetData(key string) interface{} {
	cache.Lock.RLock()
	defer cache.Lock.RUnlock()
	value, hasData := cache.Data[key]
	if !hasData {
		return nil
	}
	return value.Data
}

// GetAllKeys 获取全部缓存数据
func (cache *Cache) GetAllKeys() []string {
	var res []string
	cache.Lock.RLock()
	defer cache.Lock.RUnlock()
	for k, _ := range cache.Data {
		res = append(res, k)
	}
	return res
}

// CacheClean 清除过期数据
func (cache *Cache) CacheClean() {
	cache.Lock.Lock()
	for k, v := range cache.Data {
		if time.Now().After(v.Time.Add(cache.Expire)) {
			delete(cache.Data, k)
		}
	}
	cache.Lock.Unlock()
}
