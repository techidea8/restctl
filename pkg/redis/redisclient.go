package redis

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	log "github.com/techidea8/restctl/pkg/log"
)

func NewRedisClient(p *redis.Pool) *RedisClient {

	return &RedisClient{
		redisPool: p,
	}
}

type RedisClient struct {
	redisPool *redis.Pool
}

func (this *RedisClient) Exec(cmd string, key interface{}, args ...interface{}) (interface{}, error) {
	con := this.redisPool.Get()
	if err := con.Err(); err != nil {
		log.Errorf("%s->%v,%s,", cmd, key, err.Error())
		return nil, err
	}
	defer con.Close()
	parmas := make([]interface{}, 0)
	parmas = append(parmas, key)

	if len(args) > 0 {
		for _, v := range args {
			parmas = append(parmas, v)
		}
	}
	replay, err := con.Do(cmd, parmas...)
	if err != nil {
		log.Errorf("%s->%v,%s,", cmd, key, err.Error())
		return nil, err
	} else {
		return replay, err
	}
}

func (this *RedisClient) Set(k, v interface{}) (interface{}, error) {
	return this.Exec("set", k, v)
}

func (this *RedisClient) HSet(k, f, v interface{}) (interface{}, error) {
	//HSET KEY_NAME FIELD VALUE
	return this.Exec("hset", k, f, v)
}
func (this *RedisClient) HGet(k, f string) (r string, err error) {
	//HSET KEY_NAME FIELD VALUE
	result, e := this.Exec("hget", k, f)
	if e != nil {
		return "", e
	}
	if result != nil {
		return fmt.Sprintf("%s", result), e
	} else {
		return "", nil
	}

}

func (this *RedisClient) Get(k string) (r interface{}, err error) {
	result, e := this.Exec("get", k)
	if e != nil {
		return "", e
	}
	if result != nil {
		return result, e
	} else {
		return nil, nil
	}
}

func (this *RedisClient) SetKeyExpire(k string, ex int) (interface{}, error) {

	return this.Exec("EXPIRE", k, ex)

}

func (this *RedisClient) SetKeyExpireHourLater(k string, ex int) (interface{}, error) {
	return this.SetKeyExpire(k, ex*3600)
}

func (this *RedisClient) SetKeyExpireMinitusLater(k string, ex int) (interface{}, error) {
	return this.SetKeyExpire(k, ex*60)
}
func (this *RedisClient) SetKeyExpireSecondLater(k string, ex int) (interface{}, error) {
	return this.SetKeyExpire(k, ex)
}
func (this *RedisClient) Exists(k string) bool {
	c := this.redisPool.Get()
	defer c.Close()
	exist, err := redis.Bool(c.Do("EXISTS", k))

	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return exist
	}
}

// 获得键值时间
func (this *RedisClient) Ttl(k string) int64 {
	r, err := this.Exec("TTL", k)
	if err != nil {
		return -1
	} else {
		return r.(int64)
	}
}

func (this *RedisClient) DelKey(k string) error {
	_, err := this.Exec("DEL", k)
	return err
}
