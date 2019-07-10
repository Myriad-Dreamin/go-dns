package msg

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	rtype "github.com/Myriad-Dreamin/go-dns/msg/rec/rtype"
	mdnet "github.com/Myriad-Dreamin/go-dns/net"
)

/*
                                        1  1  1  1  1  1
          0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                                               |
        /                                               /
        /                      NAME                     /
        |                                               |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                      TYPE                     |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                     CLASS                     |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                      TTL                      |
        |                                               |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                   RDLENGTH                    |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
        /                     RDATA                     /
        /                                               /
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

where:

NAME            a domain name to which this resource record pertains.

TYPE            two octets containing one of the RR type codes.  This
                field specifies the meaning of the data in the RDATA
                field.

CLASS           two octets which specify the class of the data in the
                RDATA field.

TTL             a 32 bit unsigned integer that specifies the time
                interval (in seconds) that the resource record may be
                cached before it should be discarded.  Zero values are
                interpreted to mean that the RR can only be used for the
                transaction in progress, and should not be cached.

RDLENGTH        an unsigned 16 bit integer that specifies the length in
                octets of the RDATA field.

RDATA           a variable length string of octets that describes the
                resource.  The format of this information varies
                according to the TYPE and CLASS of the resource record.
                For example, the if the TYPE is A and the CLASS is IN,
                the RDATA field is a 4 octet ARPA Internet address.
*/
type DNSAnswer struct {
	Name     []byte
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RDData   interface{}
}

type SOA struct {
	PrimaryNS       []byte
	MailTo          []byte
	SerialNumber    uint32
	RefreshInterval uint32
	RetryInterval   uint32
	ExpireLimit     uint32
	MinimumTTL      uint32
}

type MX struct {
	Preference   uint16
	MailExchange []byte
}

func (soa *SOA) Len() int {
	return len(soa.PrimaryNS) + len(soa.MailTo) + 20
}

func (mx *MX) Len() int {
	return len(mx.MailExchange) + 2
}

func (a DNSAnswer) Size() uint16 {
	switch rdata := a.RDData.(type) {
	case []byte:
		return uint16(len(a.Name) + len(rdata) + 10)
	case SOA:
		return uint16(len(a.Name) + rdata.Len() + 10)
	default:
		panic("???? RDData Must...")
	}
}

func InitReply(q DNSQuestion) *DNSAnswer {
	a := new(DNSAnswer)
	a.Name = q.Name
	a.Type = q.Type
	a.Class = q.Class
	return a
}

func (a *DNSAnswer) SetTTL(ttl uint32) bool { // big/little endian problem
	a.TTL = ttl
	return true
}

func (a *DNSAnswer) ReadFrom(bs []byte, offset int) (int, error) {
	var cnt, l int
	var b []byte
	var err error
	a.Name, l, err = GetStringFullName(bs, offset)
	if err != nil {
		return 0, err
	}
	cnt += l
	buffer := bytes.NewBuffer(bs[offset+cnt:])
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	a.Type = binary.BigEndian.Uint16(b)

	// a.Type
	cnt += 2

	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	a.Class = binary.BigEndian.Uint16(b)

	// a.Class
	cnt += 2

	if b, err = ReadnBytes(buffer, 4); err != nil {
		return 0, err
	}
	a.TTL = binary.BigEndian.Uint32(b)

	// a.TTL
	cnt += 4
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	a.RDLength = binary.BigEndian.Uint16(b)

	// a.RDLength
	cnt += 2

	switch a.Type {
	case rtype.NS, rtype.CNAME:
		a.RDData, _, err = GetFullName(bs, offset+cnt)
		if err != nil {
			return 0, err
		}
	case rtype.SOA:
		var l2, l3 int
		rdata := new(SOA)
		rdata.PrimaryNS, l, err = GetFullName(bs, offset+cnt)
		if err != nil {
			return 0, err
		}
		rdata.MailTo, l2, err = GetFullName(bs, offset+cnt+l)
		if err != nil {
			return 0, err
		}
		l3 = offset + cnt + l + l2
		if l3+20 > len(bs) {
			return 0, errors.New("overflow when decoding soa..")
		}
		rw := mdnet.NewIO()
		rw.Write(bs[l3 : l3+20])
		rw.Read(&rdata.SerialNumber)
		rw.Read(&rdata.RefreshInterval)
		rw.Read(&rdata.RetryInterval)
		rw.Read(&rdata.ExpireLimit)
		rw.Read(&rdata.MinimumTTL)
		// a.RDLength = uint16(l2 + l3 + 20)
		a.RDLength = uint16(len(rdata.PrimaryNS) + len(rdata.MailTo) + 20)
		a.RDData = rdata
	case rtype.MX:
		rw := mdnet.NewIO()
		rdata := new(MX)
		rw.Write(bs[offset+cnt : offset+cnt+2])
		rw.Read(&rdata.Preference)
		rdata.MailExchange, _, err = GetFullName(bs, offset+cnt+2)
		if err != nil {
			return 0, err
		}
		a.RDLength = uint16(len(rdata.MailExchange) + 2)
		a.RDData = rdata
	case rtype.A, rtype.AAAA, rtype.TXT:
		fallthrough
	default:
		a.RDData, err = ReadnBytes(buffer, int(a.RDLength))
		if err != nil {
			return 0, err
		}
		// return 0, errors.New("Resource type not supported")
	}
	cnt += int(a.RDLength)
	return cnt, nil
}

func (a *DNSAnswer) Print() {
	fmt.Printf(
		"AnswerInfo:\nName:%s\nType:%d\nClass:%d\nTLL:%d\nRDLength:%d\nRDData:%x\n\n",
		a.Name,
		a.Type,
		a.Class,
		a.TTL,
		a.RDLength,
		a.RDData,
	)
	// sname, err := a.SName()
	// if err != nil {
	// 	fmt.Println("Wrong DNSAnswer format")
	// 	return
	// }
	// fmt.Printf(
	// 	"AnswerInfo:\nName:%s\nType:%d\nClass:%d\nTLL:%d\nRDLength:%d\nRDData:%x\n\n",
	// 	sname,
	// 	a.Type,
	// 	a.Class,
	// 	a.TTL,
	// 	a.RDLength,
	// 	a.RDData,
	// )
}

func (a *DNSAnswer) SName() (string, error) {
	var s string
	var n, l, flag int
	l = len(a.Name)
	for i := 0; ; i++ {
		if i >= l {
			return "", errors.New("Index out of range")
		}
		n = int(a.Name[i])
		if n == 0 {
			break
		} else {
			if flag == 0 {
				flag = 1
			} else {
				s += string('.')
			}
			if i+n+1 >= l {
				return "", errors.New("Index out of range")
			}
			for j := 0; j < n; j++ {
				s += string(a.Name[i+1+j])
			}
			i = i + n
		}
	}
	return s, nil
}

func (a *DNSAnswer) ToBytes() ([]byte, error) {
	var buf = mdnet.NewIO()
	b, err := ToDNSDomainName(a.Name)
	if err != nil {
		return nil, err
	}
	buf.Write(b)
	buf.Write(a.Type)
	buf.Write(a.Class)
	buf.Write(a.TTL)
	buf.Write(a.RDLength)
	switch a.Type {
	case rtype.SOA:
		buf.Write(a.RDData.(*SOA).PrimaryNS)
		buf.Write(a.RDData.(*SOA).MailTo)
		buf.Write(a.RDData.(*SOA).SerialNumber)
		buf.Write(a.RDData.(*SOA).RefreshInterval)
		buf.Write(a.RDData.(*SOA).RetryInterval)
		buf.Write(a.RDData.(*SOA).ExpireLimit)
		buf.Write(a.RDData.(*SOA).MinimumTTL)
	case rtype.MX:
		buf.Write(a.RDData.(*MX).Preference)
		buf.Write(a.RDData.(*MX).MailExchange)
	default:
		buf.Write(a.RDData.([]byte))
	}
	return buf.Bytes(), nil
}

func (a *DNSAnswer) SType() (string, error) {
	stype, suc := Typename[a.Type]
	if suc != true {
		return "", errors.New("No such RR type")
	}
	return stype, nil
}

func (a *DNSAnswer) RedisKey() (string, error) {
	// sname, err := a.SName()
	// if err != nil {
	// 	return "", err
	// }
	sname := string(a.Name)
	stype, err := a.SType()
	if err != nil {
		return "", err
	}
	return sname + ":" + stype, nil
}

func (a *DNSAnswer) RedisRandomKey() (string, error) {
	rand.Seed(time.Now().UnixNano())
	hash := rand.Int31()
	rkey, err := a.RedisKey()
	if err != nil {
		return "", err
	}
	return rkey + ":" + strconv.Itoa(int(hash)), nil
}

func (a *DNSAnswer) RedisHashKey() (string, error) {
	var bs []byte
	hash := md5.New()
	rw := mdnet.NewIO()
	rw.Write(a.Name)
	rw.Write(a.Type)
	rw.Write(a.RDData)
	bs = rw.Bytes()
	if _, err := hash.Write(bs); err != nil {
		return "", err
	}
	rkey, err := a.RedisKey()
	if err != nil {
		return "", err
	}
	return rkey + ":" + hex.EncodeToString(hash.Sum(nil)), nil
}

func (a *DNSAnswer) RedisAuthorityHashKey(qt uint16) (string, error) {
	var bs []byte
	var rkey string
	hash := md5.New()
	rw := mdnet.NewIO()
	rw.Write(a.Name)
	rw.Write(qt)
	rw.Write(a.Type)
	rw.Write(a.RDData)
	bs = rw.Bytes()
	if _, err := hash.Write(bs); err != nil {
		return "", err
	}
	rkey, err := a.RedisKey()
	if err != nil {
		return "", err
	}
	rkey += ":" + Typename[qt]
	return rkey + ":" + hex.EncodeToString(hash.Sum(nil)), nil
}
