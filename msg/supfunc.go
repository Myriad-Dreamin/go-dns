package msg

import (
	"bytes"
	"errors"
	//"fmt"
	"io"
	"strings"

	mdnet "github.com/Myriad-Dreamin/go-dns/net"
)

var Typename = map[uint16]string{
	1:  "A",
	2:  "NS",
	3:  "MD",
	4:  "MF",
	5:  "CNAME",
	6:  "SOA",
	7:  "MB",
	8:  "MG",
	9:  "MR",
	10: "NULL",
	11: "WKS",
	12: "PTR",
	13: "HINFO",
	14: "MINFO",
	15: "MX",
	16: "TXT",
	28: "AAAA",
}

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
			return nil, 0, errors.New("Format error when en/decoding: index out of range")
		}
		for i := offset; ; i++ {
			n = int(bs[i])
			if n&0xc0 == 0xc0 {
				if i+1 >= len(bs) {
					return nil, 0, errors.New("Format error when en/decoding: index out of range")
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
					return nil, 0, errors.New("Format error when en/decoding: index out of range")
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

func GetStringFullName(bs []byte, offset int) ([]byte, int, error) {
	var buf bytes.Buffer
	var n, flag int
	var cnt int
	for k := 0; k < 1<<10; k++ {
		if offset < 0 || offset >= len(bs) {
			return nil, 0, errors.New("Format error when en/decoding: index out of range")
		}
		for i := offset; ; i++ {
			n = int(bs[i])
			if n&0xc0 == 0xc0 {
				if i+1 >= len(bs) {
					return nil, 0, errors.New("Format error when en/decoding: index out of range")
				}
				offset = n&0x3f<<8 + int(bs[i+1])
				if k == 0 {
					cnt += 2
				}
				break
			} else {
				//buf.WriteByte(bs[i])
				if k == 0 {
					cnt += n + 1
				}
				if n == 0 {
					return buf.Bytes(), cnt, nil
				}
				if flag == 0 {
					flag = 1
				} else {
					buf.WriteByte(byte('.'))
				}
				if i+n+1 >= len(bs) {
					return nil, 0, errors.New("Format error when en/decoding: index out of range")
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
	if dnm == nil {
		return []byte{0}, nil
	}
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

func CompressName(buf *bytes.Buffer, sufpos map[string]int, bytename []byte) error {
	name := strings.Split(string(bytename), ".")
	var (
		suffix string
		trunc  int
		nxoff  int
		offset int
		prelen int
		flag   bool
	)
	suffix = name[len(name)-1]
	// trunc = len(name)
	offset = buf.Len()
	prelen = len(bytename)
	for j := len(name) - 1; j >= 0; j-- {
		if j == len(name)-1 {
			suffix = name[j]
			prelen -= len(name[j])
		} else {
			suffix = name[j] + "." + suffix
			prelen -= len(name[j]) + 1
		}
		if _, ok := sufpos[suffix]; ok == false {
			sufpos[suffix] = offset + prelen
		} else {
			flag = true
			trunc = j
			nxoff = sufpos[suffix]
		}
	}
	if flag {
		for j := 0; j < trunc; j++ {
			buf.WriteByte(byte(len(name[j])))
			buf.Write([]byte(name[j]))
		}
		tmp := make([]byte, 2)
		if nxoff > 0x3fff {
			return errors.New("Offset out of range")
		}
		tmp[0] = uint8(0xc0 | (nxoff>>8)&0xff)
		tmp[1] = uint8(nxoff & 0xff)
		buf.Write(tmp)
	} else {
		for j := 0; j < len(name); j++ {
			buf.WriteByte(byte(len(name[j])))
			buf.Write([]byte(name[j]))
		}
		buf.WriteByte(byte(0))
	}
	return nil
}
