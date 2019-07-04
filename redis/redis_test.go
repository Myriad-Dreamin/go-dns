package mredis

import (
	"fmt"
	// "github.com/go-redis/redis"
	"encoding/hex"
	"testing"
	"time"

	"github.com/Myriad-Dreamin/go-dns/msg"
	"github.com/garyburd/redigo/redis"
)

// rediskey:
// www.google.Com:AAAA
//
// Get key:
// AnswerInfo:
// Name:www.google.Com
// Type:28
// Class:1
// TLL:5
// RDLength:16
// RDData:2400cb0020480001000000006814224e

// redis get failed: redigo: nil returned

func TestRedis(t *testing.T) {
	var hexData = "12b6818000010001000000000377777706676f6f676c6503436f6d00001c0001c00c001c00010000000500102400cb0020480001000000006814224e"
	var msgMessage msg.DNSMessage
	var bytesData, err = hex.DecodeString(hexData)

	_, err = msgMessage.Read(bytesData)
	if err != nil {
		fmt.Println(err)
		return
	}

	key, err := msgMessage.Answer[0].RedisKey()
	val, err := msgMessage.Answer[0].ToBytes()
	if err != nil {
		fmt.Println(err)
		return
	}
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}
	defer c.Close()
	_, err = c.Do("SET", key, val, "EX", msgMessage.Answer[0].TTL) // 5s
	if err != nil {
		fmt.Println("redis set failed:", err)
	}
	keyval, err := redis.Bytes(c.Do("GET", key))
	var redismsg msg.DNSAnswer
	redismsg.ReadFrom(keyval, 0)
	if err != nil {
		fmt.Println("redis get failed:", err)
	} else {
		fmt.Printf("Get key:\n")
		redismsg.Print()
	}
	time.Sleep(6 * time.Second)
	keyval, err = redis.Bytes(c.Do("GET", key))
	redismsg.ReadFrom(keyval, 0)
	if err != nil {
		fmt.Println("redis get failed:", err)
	} else {
		fmt.Printf("Get key:\n")
		redismsg.Print()
	}

}
