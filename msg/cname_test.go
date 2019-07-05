package msg

import (
	"encoding/hex"
	"fmt"
	"net"
	"testing"
)

func TestCNAME(t *testing.T) {
	var hexData = "c61781800001000900000001037777770331363303636f6d0000010001c00c00050001000000050017037777770331363303636f6d083136336a69617375c014c02900050001000000050017037777770331363303636f6d06627367736c6202636e00c04c0005000100000005000d087a313633697076360176c058c06f00010001000000050004dcaab589c06f00010001000000050004dcaab58ec06f00010001000000050004dcaab58ac06f00010001000000050004dcaab591c06f00010001000000050004dcaab58dc06f00010001000000050004dcaab58c0000291000000000050000"

	var bytesData, err = hex.DecodeString(hexData)
	if err != nil {
		t.Error(err)
		return
	}
	var msgMessage DNSMessage

	msgMessage.Read(bytesData)

	//msgMessage.Print()

	dns := "192.168.1.1:53"
	local := "127.0.0.1:55555"
	_, err = net.ResolveUDPAddr("udp", local)
	conn, err := net.Dial("udp", dns)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	for {

		// bs, err := msgMessage.CompressToBytes()
		// if err != nil {
		// 	fmt.Println("Error: ", err)
		// }
		// _, err = conn.Write(bs)

		bs, err := msgMessage.ToBytes()
		if err != nil {
			fmt.Println("Error: ", err)
		}
		_, err = conn.Write(bs)
		break
	}
}
