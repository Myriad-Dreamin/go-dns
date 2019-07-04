package mredis

import (
	"fmt"
	// "github.com/go-redis/redis"
	"encoding/hex"
	"github.com/Myriad-Dreamin/go-dns/msg"
	"github.com/garyburd/redigo/redis"
	"testing"
	"time"
)

// Get key: google.com:TXT:770906257
// AnswerInfo:
// Name:google.com
// Type:16
// Class:1
// TLL:5
// RDLength:46
// RDData:2d646f63757369676e3d30353935383438382d343735322d346566322d393565622d616137626138613362643065

// Get key: google.com:TXT:1137553654
// AnswerInfo:
// Name:google.com
// Type:16
// Class:1
// TLL:5
// RDLength:36
// RDData:23763d7370663120696e636c7564653a5f7370662e676f6f676c652e636f6d207e616c6c

// Get key: google.com:TXT:869102121
// AnswerInfo:
// Name:google.com
// Type:16
// Class:1
// TLL:5
// RDLength:46
// RDData:2d646f63757369676e3d31623061363735342d343962312d346462352d383534302d643263313236363462323839

// Get key: google.com:TXT:679595641
// AnswerInfo:
// Name:google.com
// Type:16
// Class:1
// TLL:5
// RDLength:65
// RDData:40676c6f62616c7369676e2d736d696d652d64763d434459582b584648557732776d6c362f4762382b353942734833314b7a55723663316c32425076714b58383d

// Get key: google.com:TXT:1590375190
// AnswerInfo:
// Name:google.com
// Type:16
// Class:1
// TLL:5
// RDLength:46
// RDData:2d646f63757369676e3d30353935383438382d343735322d346566322d393565622d616137626138613362643065

// redis get failed: redigo: nil returned

func TestRedis(t *testing.T) {

	var hexData = "25818180000100050000000006676f6f676c6503636f6d0000100001c00c0010000100000005002e2d646f63757369676e3d30353935383438382d343735322d346566322d393565622d616137626138613362643065c00c0010000100000005002423763d7370663120696e636c7564653a5f7370662e676f6f676c652e636f6d207e616c6cc00c0010000100000005003c3b66616365626f6f6b2d646f6d61696e2d766572696669636174696f6e3d3232726d3535316375346b3061623062787377353336746c647334683935c00c0010000100000005004140676c6f62616c7369676e2d736d696d652d64763d434459582b584648557732776d6c362f4762382b353942734833314b7a55723663316c32425076714b58383dc00c0010000100000005002e2d646f63757369676e3d31623061363735342d343962312d346462352d383534302d643263313236363462323839"
	var msgMessage msg.DNSMessage
	var bytesData, err = hex.DecodeString(hexData)

	_, err = msgMessage.Read(bytesData)
	if err != nil {
		fmt.Println(err)
		return
	}

	key, err := msgMessage.Answer[0].RedisRandomKey()
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
	_, err = c.Do("set", key, val, "EX", msgMessage.Answer[0].TTL) // 5s
	// _, err = c.Do("set", key, val) // 5s
	if err != nil {
		fmt.Println("redis set failed:", err)
	}

	if err = PushToRedis(msgMessage.Answer, c); err != nil {
		fmt.Println(err)
		return
	}
	if err = PushToRedis(msgMessage.Authority, c); err != nil {
		fmt.Println(err)
		return
	}
	if err = PushToRedis(msgMessage.Additional, c); err != nil {
		fmt.Println(err)
		return
	}

	searchkey, err := msgMessage.Question[0].RedisKey()

	fmt.Println(searchkey)

	keys, err := redis.Strings(c.Do("keys", searchkey+":*"))

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, k := range keys {

		keyval, err := redis.Bytes(c.Do("GET", k))

		var redismsg msg.DNSAnswer
		redismsg.ReadFrom(keyval, 0)
		if err != nil {
			fmt.Println("redis get failed:", err)
		} else {
			fmt.Printf("Get key: %s \n", k)
			redismsg.Print()
		}
		time.Sleep(1 * time.Second)
	}
}
