package cache

import (
	"time"

	redism "github.com/techidea8/restctl/pkg/redis"

	"github.com/garyburd/redigo/redis"
)

type RedisCache struct {
	redisClient *redism.RedisClient
	name        string
}

func NewRedisCache(conf *redism.RedisConf) *RedisCache {
	return &RedisCache{
		redisClient: redism.NewRedisClient(redism.NewRedisPool(conf)),
		name:        "rediscache",
	}
}
func (fc *RedisCache) WriteCache(key string, v string) (err error) {
	if _, err = fc.redisClient.Set(key, v); err != nil {
		return
	}

	return err
}
func (fc *RedisCache) Name() string {
	return fc.name
}
func (fc *RedisCache) WriteCacheWithTTL(key string, v string, ttl time.Duration) (err error) {
	fc.redisClient.Set(key, v)
	fc.Expire(key, ttl)
	return
}

func (fc *RedisCache) KeyExists(key string) (exists bool) {
	return fc.redisClient.Exists(key)
}

func (fc *RedisCache) ReadCache(key string) (value string, e error) {
	return redis.String(fc.redisClient.Get(key))
}
func (fc *RedisCache) Expire(key string, duration time.Duration) (err error) {
	_, err = fc.redisClient.SetKeyExpireSecondLater(key, int(duration.Seconds()))
	return
}

func (fc *RedisCache) RemoveCache(key string) (err error) {
	return fc.redisClient.DelKey(key)
}
