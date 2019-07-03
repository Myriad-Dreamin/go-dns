package mdnet

import (
	"encoding/hex"
	"fmt"
	"github.com/Myriad-Dreamin/go-dns/msg"
	"net"
	"testing"
)

// Header
// 59ad81800001000300000000
// Question
// 0377777705626169647503636f6d0000010001
// Answer0
// c00c0005000100000005000f0377777701610673686966656ec016

func TestUdp(t *testing.T) {
	var hexData = "59ad010000010000000000000377777705626169647503636f6d0000050001"
	// var hexData = "59ad818000010003000000000377777705626169647503636f6d0000010001c00c0005000100000005000f0377777701610673686966656ec016c02b000100010000000500040ed7b127c02b000100010000000500040ed7b126"
	var bytesData, err = hex.DecodeString(hexData)
	if err != nil {
		t.Error(err)
		return
	}
	var msgMessage msg.DNSMessage
	var recvMessage msg.DNSMessage

	msgMessage.Read(bytesData)

	dns := "192.168.1.1:53"
	// dns := "192.168.0.1:53"
	local := "127.0.0.1:55555"
	_, err = net.ResolveUDPAddr("udp", local)
	// sAddr, err := net.ResolveUDPAddr("udp", dns)

	// /* Now listen at selected port */
	// serverconn, err := net.ListenUDP("udp", lAddr)
	conn, err := net.Dial("udp", dns)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	for {

		n, err := conn.Write(msgMessage.ToBytes())
		if err != nil {
			fmt.Println("Error: ", err)
		}

		buf := make([]byte, 1024)
		n, err = conn.Read(buf)
		recvMessage.Read(buf[:n])
		recvMessage.Print()
		break
	}
	// fmt.Printf("%x\n", msgMessage.ToBytes())

}
