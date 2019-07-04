package mredis

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

type RedisCache struct {
	pool *redis.Pool
}

var (
	pool             *redis.Pool
	RedisCacheClient *RedisCache
)

func init() {
	pool = newPool("127.0.0.1:6379", "")
	RedisCacheClient = &RedisCache{pool: pool}
}

func newPool(server string, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1024,
		IdleTimeout: 600 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp", server, redis.DialPassword(password),
				redis.DialReadTimeout(5000*time.
					Millisecond), redis.DialWriteTimeout(5000*time.Millisecond), redis.DialDatabase(10))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
