package cache

import (
	"git.inke.cn/BackendPlatform/go-tools/cast"
	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/sync/singleflight"
	"time"
)

type GetValFunc func(key string) (interface{}, error)

type Cache struct {
	cache       *lru.Cache
	getValFunc  GetValFunc
	singleFight singleflight.Group
	expireTime  time.Duration
}

type cacheVal struct {
	expireAt int64
	Val      interface{}
}

func NewCache(size int, expire time.Duration, getValFunc GetValFunc) *Cache {
	if size <= 0 || expire <= 0 || getValFunc == nil {
		panic("getValFunc is nil")
	}
	return &Cache{
		getValFunc: getValFunc,
		expireTime: expire,
		cache: func() (cache *lru.Cache) {
			var err error
			if cache, err = lru.New(size); err != nil {
				panic("cache param err") //这里出错就是 size <= 0
			}
			return
		}(),
	}
}

func (c *Cache) add(key string, val interface{}) {
	c.cache.Add(key, cacheVal{
		expireAt: time.Now().Unix(),
		Val:      val,
	})
}
func (c *Cache) Remove(key string) {
	c.cache.Remove(key)
}

func (c *Cache) Get(key string) (result interface{}, err error) {
	val, ok := c.cache.Get(key)
	if !ok || time.Now().Add(-c.expireTime).Unix() > val.(cacheVal).expireAt {
		lockKey := cast.ToString(time.Now().Minute())
		defer c.singleFight.Forget(lockKey)
		if result, err, _ = c.singleFight.Do(lockKey, func() (interface{}, error) {
			return c.getValFunc(key)
		}); err != nil {
			if ok {
				result = val.(cacheVal).Val
			}
			return
		}
		c.add(key, result)
		return
	}

	return val.(cacheVal).Val, err
}
