package mredis

import (
	"github.com/Myriad-Dreamin/go-dns/msg"
	"github.com/garyburd/redigo/redis"
)

func PushToRedis(answers []msg.DNSAnswer, client redis.Conn) error {
	for _, ans := range answers {
		key, err := ans.RedisRandomKey()
		if err != nil {
			return err
		}
		//b := ans.RDData
		b, err := ans.ToBytes()
		if err != nil {
			return err
		}
		client.Do("set", key, b, "EX", ans.TTL)
	}
	return nil
}
