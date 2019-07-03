package mdnet

import "fmt"
import "testing"
import "encoding/hex"

func TestIO(t *testing.T) {
	rw := NewIO()
	if err := rw.Write(int16(0x3ff3)); err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%v\n", hex.EncodeToString(rw.Bytes()))
}
