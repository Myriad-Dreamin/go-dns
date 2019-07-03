package mdnet

import "fmt"
import "unsafe"
import "encoding/binary"
import "testing"
import "encoding/hex"
import "github.com/Myriad-Dreamin/go-dns/msg"

// Header
// 59ad81800001000300000000
// Question
// 0377777705626169647503636f6d0000010001
// Answer0
// c00c0005000100000005000f0377777701610673686966656ec016

func TestEncodeMessage(t *testing.T) {
	var hexData = "59ad818000010003000000000377777705626169647503636f6d0000010001c00c0005000100000005000f0377777701610673686966656ec016c02b000100010000000500040ed7b127c02b000100010000000500040ed7b126"
	var bytesData, err = hex.DecodeString(hexData)
	if err != nil {
		t.Error(err)
		return
	}
	var msgMessage msg.DNSMessage

	msgMessage.Read(bytesData)

	fmt.Printf("%x\n", msgMessage.ToBytes())
}

func testLittleEndian() {

	// 0000 0000 0000 0000   0000 0001 1111 1111
	var testInt int32 = 256
	fmt.Printf("%d use little endian: \n", testInt)
	var testBytes []byte = make([]byte, 4)
	binary.LittleEndian.PutUint32(testBytes, uint32(testInt))
	fmt.Println("int32 to bytes:", testBytes)

	convInt := binary.LittleEndian.Uint32(testBytes)
	fmt.Printf("bytes to int32: %d\n\n", convInt)
}

func testBigEndian() {
	s := int16(0x1234)
	b := int8(s)
	fmt.Println("int16字节大小为", unsafe.Sizeof(s)) //结果为2
	if 0x34 == b {
		fmt.Println("little endian")
	} else {
		fmt.Println("big endian")
	}

	// 0000 0000 0000 0000   0000 0001 1111 1111
	var testInt int32 = 256
	fmt.Printf("%d use big endian: \n", testInt)
	var testBytes []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(testBytes, uint32(testInt))
	fmt.Println("int32 to bytes:", testBytes)

	convInt := binary.BigEndian.Uint32(testBytes)
	fmt.Printf("bytes to int32: %d\n\n", convInt)
}

func IsLittleEndian() bool {
	n := 0x1234
	return *(*byte)(unsafe.Pointer(&n)) == 0x34
}
