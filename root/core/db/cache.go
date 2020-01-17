package db

import (
	"github.com/astaxie/beego/cache"
	"time"
)

// 数据cache
type MyCache struct {
	cache cache.Cache
}

// 新建cache 其中config为json格式
func NewCache(engine, config string) *MyCache {
	mycache := &MyCache{}
	cache, err := cache.NewCache(engine, config)
	if err != nil {
		panic(err)
	}
	mycache.cache = cache
	return mycache
}

// 存储数据
func (self *MyCache) Save(prefix string, key string, value interface{}, timeout time.Duration) {
	err := self.cache.Put(prefix+key, value, time.Duration(timeout))
	if err != nil {
		panic(err)
	}
}

// 加载数据
func (self *MyCache) Load(prefix string, key string) interface{} {
	return self.cache.Get(prefix + key)
}

// 按key删除数据
func (self *MyCache) DeleteByKey(prefix string, key string) {
	self.cache.Delete(prefix + key)
}
