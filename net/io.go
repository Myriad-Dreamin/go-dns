package mdnet

import "io"
import "bytes"
import "encoding/binary"

var (
	NetWorkEndian = binary.BigEndian
)

type PacketableBuffer interface {
	io.ReadWriter
	Bytes() []byte
	String() string
}

type IO struct {
	Buffer PacketableBuffer
	Endian binary.ByteOrder
}

func NewIO() (rw *IO) {
	return &IO{
		Buffer: new(bytes.Buffer),
		Endian: NetWorkEndian,
	}
}

func NewIOObj(b interface{}) (rw *IO) {
	rw = NewIO()
	if err := rw.Write(b); err != nil {
		panic(err)
	}
	return rw
}

func (rw *IO) Write(b interface{}) error {
	return binary.Write(rw.Buffer, rw.Endian, b)
}

func (rw *IO) Bytes() []byte {
	return rw.Buffer.Bytes()
}

func (rw *IO) String() string {
	return rw.Buffer.String()
}
