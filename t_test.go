package cache

import (
	"github.com/spf13/cast"
	"sync"
	"testing"
	"time"
)

var cach = NewCache(100000000, 10*time.Second, func(key string) (interface{}, error) {

	return time.Now().UnixNano(), nil
}).WithLockPrecision(time.Second)

func TestCache(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	for idx := 0; idx < 5; idx++ {
		go func(idx int) {
			defer wg.Done()
			for i := 0; i < 10000; i++ {
				t.Log(cach.Get(cast.ToString(i)))
			}
		}(idx)
	}
	wg.Wait()

}
