package redis

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

func NewRedisPool(redisCfg *RedisConf) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     redisCfg.MaxIdle,
		IdleTimeout: redisCfg.IdleTimeoutSec * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(
				redisCfg.Addr,
				redis.DialDatabase(redisCfg.DBIndex),
				redis.DialPassword(redisCfg.Password),
			)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				log.Fatalf("ping redis error: %s", err)
				return err
			}
			return nil
		},
	}
}
