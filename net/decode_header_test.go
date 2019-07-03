package mdnet

import "fmt"
import "testing"
import "encoding/hex"
import "encoding/binary"
import "github.com/Myriad-Dreamin/go-dns/msg"

/*
59ad81800001000300000000
err? <nil>.
{22957 33152 1 3 0 0}
ID:59ad
Flags:8180
quc:1
anc:3
auc:0
adc:0
59ad81800001000300000000
err? <nil>.
{44377 32897 256 768 0 0}
ID:ad59
Flags:8081
quc:100
anc:300
auc:0
adc:0
*/

func TestDecodeHeader(t *testing.T) {
	var hexData = "59ad81800001000300000000"
	var bytesData, err = hex.DecodeString(hexData)
	if err != nil {
		t.Error(err)
		return
	}
	var msgHeader msg.DNSHeader
	rw := NewIO()
	if err := rw.Write(bytesData); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%v\nerr? %v.\n%v\n", hex.EncodeToString(rw.Bytes()), rw.Read(&msgHeader), msgHeader)
	fmt.Printf(
		"ID:%x\nFlags:%x\nquc:%x\nanc:%x\nauc:%x\nadc:%x\n",
		msgHeader.ID,
		msgHeader.Flags,
		msgHeader.QDCount,
		msgHeader.ANCount,
		msgHeader.NSCount,
		msgHeader.ARCount,
	)

	if err := rw.Write(bytesData); err != nil {
		t.Error(err)
		return
	}
	rw.SetByteOrder(binary.LittleEndian)

	fmt.Printf("%v\nerr? %v.\n%v\n", hex.EncodeToString(rw.Bytes()), rw.Read(&msgHeader), msgHeader)
	fmt.Printf(
		"ID:%x\nFlags:%x\nquc:%x\nanc:%x\nauc:%x\nadc:%x\n",
		msgHeader.ID,
		msgHeader.Flags,
		msgHeader.QDCount,
		msgHeader.ANCount,
		msgHeader.NSCount,
		msgHeader.ARCount,
	)
}
