package msg

import "bytes"

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
	for i := 0; i < n; i++ {
		c, _ := buffer.ReadByte()
		res.WriteByte(c)
	}
	return res.Bytes(), nil
}

func GetFullName(bs []byte, offset int) ([]byte, int) {
	var buf bytes.Buffer
	var n int
	var cnt int
	for i := offset; ; i++ {
		n = int(bs[i])
		if n&0xc0 == 0xc0 {
			nof := n&0x3f<<8 + int(bs[i+1])
			b, _ := GetFullName(bs, nof)
			buf.Write(b)
			cnt += 2
			return buf.Bytes(), cnt
		} else {
			buf.WriteByte(bs[i])
			cnt += n + 1
			if n == 0 {
				return buf.Bytes(), cnt
			}
			for j := 0; j < n; j++ {
				buf.WriteByte(bs[i+1+j])
			}
			i = i + n
		}
	}
	return buf.Bytes(), cnt
}
