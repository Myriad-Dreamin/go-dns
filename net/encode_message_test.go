package mdnet

import (
	"encoding/hex"
	"fmt"
	"github.com/Myriad-Dreamin/go-dns/msg"
	"testing"
)

// Header
// 59ad81800001000300000000
// Question
// 0377777705626169647503636f6d0000010001
// Answer0
// c00c0005000100000005000f0377777701610673686966656ec016

func TestEncodeMessage(t *testing.T) {
	var hexData = "4a5b818000010000000100000377777701610673686966656e03636f6d00001c0001c01000060001000000050033036e7331c0101062616964755f646e735f6d6173746572056261696475c01971aad0e7000000050000000500278d0000000e10"
	// var hexData = "59ad818000010003000000000377777705626169647503636f6d0000010001c00c0005000100000005000f0377777701610673686966656ec016c02b000100010000000500040ed7b127c02b000100010000000500040ed7b126"

	ip := "2400:cb00:2048:1::6814:224e"

	var bytesData, err = hex.DecodeString(hexData)
	if err != nil {
		t.Error(err)
		return
	}
	var msgMessage msg.DNSMessage

	_, err = msgMessage.Read(bytesData)
	if err != nil {
		fmt.Println(err)
		return
	}
	msgMessage.Print()
	// fmt.Printf("%x\n", msgMessage.ToBytes())
}
