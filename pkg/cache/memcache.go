package cache

import (
	"errors"
	"sync"
	"time"

	log "github.com/techidea8/restctl/pkg/log"
)

type MemCache struct {
	ticker  *time.Ticker
	datamap *sync.Map
	name    string
}
type dataitem struct {
	ExpireAt int64
	Data     any
}

func NewMemCache() *MemCache {
	instance := &MemCache{
		ticker:  time.NewTicker(time.Second * 1),
		datamap: new(sync.Map),
		name:    "memcache",
	}
	go instance.loop()
	return instance
}

func (fc *MemCache) WriteCacheWithTTL(key string, value string, duration time.Duration) error {
	fc.datamap.Store(key, dataitem{
		ExpireAt: time.Now().Unix() + int64(duration.Seconds()),
		Data:     value,
	})
	return nil
}
func (fc *MemCache) Name() string {
	return fc.name
}

func (fc *MemCache) KeyExists(key string) (exists bool) {
	_, exists = fc.datamap.Load(key)
	return
}
func (fc *MemCache) loop() {
	for {
		select {
		case <-fc.ticker.C:
			fc.clearcache()
		}

	}
}
func (fc *MemCache) clearcache() {
	fc.datamap.Range(func(key, value any) bool {
		data := value.(dataitem)
		if data.ExpireAt > 0 && data.ExpireAt < time.Now().Unix() {
			log.Debugf("remove key %s", key)
			fc.datamap.Delete(key)
		}
		return true
	})
}
func (fc *MemCache) WriteCache(key string, v string) (err error) {
	fc.datamap.Store(key, dataitem{
		ExpireAt: -1,
		Data:     v,
	})
	return
}

func (fc *MemCache) ReadCache(key string) (value string, e error) {
	v, ok := fc.datamap.Load(key)
	if ok {
		value = v.(dataitem).Data.(string)
	} else {
		e = errors.New("not exist")
	}
	return
}
func (fc *MemCache) Expire(key string, duration time.Duration) (err error) {
	v, ok := fc.datamap.Load(key)
	if !ok {
		err = errors.New("not exist")
		return
	} else {
		value := v.(dataitem)
		value.ExpireAt = int64(time.Duration(time.Now().Unix()) + duration*time.Second)
		fc.datamap.Store(key, value)
	}
	return
}

func (fc *MemCache) RemoveCache(key string) (err error) {
	fc.datamap.Delete(key)
	log.Debugf("remove key %s", key)
	return
}
