package redis

import "time"

type RedisConf struct {
	DBIndex        int
	Addr           string
	Password       string
	MaxIdle        int
	IdleTimeoutSec time.Duration
}
