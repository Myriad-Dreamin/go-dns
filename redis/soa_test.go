package mredis

import (
	"fmt"
	// "github.com/go-redis/redis"
	"encoding/hex"
	"github.com/Myriad-Dreamin/go-dns/msg"
	"testing"
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

func TestSOA(t *testing.T) {

	var hexData = "4a5b818000010000000100000377777701610673686966656e03636f6d00001c0001c01000060001000000050033036e7331c0101062616964755f646e735f6d6173746572056261696475c01971aad0e7000000050000000500278d0000000e10"
	var msgMessage msg.DNSMessage
	var bytesData, err = hex.DecodeString(hexData)

	_, err = msgMessage.Read(bytesData)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(msgMessage.Answer[0].RedisHashKey())
	conn := pool.Get()
	MessageToRedis(msgMessage, conn)
}
