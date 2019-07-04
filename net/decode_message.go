package mdnet

import "testing"
import "encoding/hex"
import "github.com/Myriad-Dreamin/go-dns/msg"

// HeaderInfo:
// ID:59ad
// Flags:8180
// quc:1
// anc:3
// auc:0
// adc:0

// QuestionInfo:
// Name:0377777705626169647503636f6d00
// Type:1
// Class:1

// AnswerInfo:
// Name:c00c
// Type:5
// Class:1
// TLL:5
// RDLength:15
// RDData:0377777701610673686966656ec016

// AnswerInfo:
// Name:c02b
// Type:1
// Class:1
// TLL:5
// RDLength:4
// RDData:0ed7b127

// AnswerInfo:
// Name:c02b
// Type:1
// Class:1
// TLL:5
// RDLength:4
// RDData:0ed7b126

// Domain name: www.baidu.com
// Domain name: www.a.shifen.com

func TestDecodeMessage(t *testing.T) {
	var hexData = "59ad818000010003000000000377777705626169647503636f6d0000010001c00c0005000100000005000f0377777701610673686966656ec016c02b000100010000000500040ed7b127c02b000100010000000500040ed7b126"
	var bytesData, err = hex.DecodeString(hexData)
	if err != nil {
		t.Error(err)
		return
	}
	var msgMessage msg.DNSMessage

	msgMessage.Read(bytesData)
	msgMessage.Print()
	msgMessage.Answer[0].SName()
	msgMessage.Answer[1].SName()
}
