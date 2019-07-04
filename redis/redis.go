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
		// client.Do("set", key, b, "EX", ans.TTL)
		client.Send("set", key, b, "EX", ans.TTL)
	}
	client.Flush()
	for i := 0; i < len(answers); i++ {
		if _, err := client.Receive(); err != nil {
			return err
		}
	}
	return nil
}
