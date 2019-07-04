package msg

import (
	"bytes"
	"errors"
	"io"

	mdnet "github.com/Myriad-Dreamin/go-dns/net"
)

func BytesToInt(bs []byte, val interface{}) int {
	var res int32
	for i := 0; i < len(bs); i++ {
		x := uint8(bs[i])
		res = (res << 8) + int32(x)
	}
	return int(res)
}

func ReadnBytes(buffer *bytes.Buffer, n int) ([]byte, error) {
	res := new(bytes.Buffer)
	if buffer.Len() < n {
		return nil, errors.New("No Enough bytes in bytes.Buffer")
	}
	for i := 0; i < n; i++ {
		c, _ := buffer.ReadByte()
		res.WriteByte(c)
	}
	return res.Bytes(), nil
}

func GetFullName(bs []byte, offset int) ([]byte, int, error) {
	var buf bytes.Buffer
	var n int
	var cnt int
	for k := 0; k < 1<<10; k++ {
		if offset < 0 || offset >= len(bs) {
			return nil, 0, errors.New("Index out of range")
		}
		for i := offset; ; i++ {
			n = int(bs[i])
			if n&0xc0 == 0xc0 {
				if i+1 >= len(bs) {
					return nil, 0, errors.New("Index out of range")
				}
				offset = n&0x3f<<8 + int(bs[i+1])
				if k == 0 {
					cnt += 2
				}
				break
			} else {
				buf.WriteByte(bs[i])
				if k == 0 {
					cnt += n + 1
				}
				if n == 0 {
					return buf.Bytes(), cnt, nil
				}
				if i+n+1 >= len(bs) {
					return nil, 0, errors.New("Index out of range")
				}
				for j := 0; j < n; j++ {
					buf.WriteByte(bs[i+1+j])
				}
				i = i + n
			}
		}
	}
	return nil, 0, errors.New("Wrong Domain Name Format")
}

// todo: ignoring the case of '\.'
func ToDNSDomainName(dnm []byte) ([]byte, error) {
	var rw = mdnet.NewIO()
	var bf = bytes.NewBuffer(dnm)
	for {
		d, err := bf.ReadBytes(byte('.'))
		if err == io.EOF {
			if d == nil {
				return nil, errors.New("nil domain name is not allowed")
			}
			rw.Write(byte(len(d)))
			rw.Write(d)
			rw.Write(byte(0))
			return rw.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}
		if len(d) < 2 {
			return nil, errors.New("nil domain name is not allowed")
		}
		d = d[:len(d)-1]
		rw.Write(byte(len(d)))
		rw.Write(d)
	}
}
