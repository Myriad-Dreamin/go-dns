package msg

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
import "bytes"
import "fmt"
import "encoding/binary"

type DNSAnswer struct {
	Name     []byte
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RDData   []byte
}

func (a *DNSAnswer) ReadFrom(bs []byte, offset int) (int, error) {
	var cnt, l int
	var b []byte
	var err error
	a.Name, l, err = GetFullName(bs, offset)
	if err != nil {
		return 0, err
	}
	cnt += l
	buffer := bytes.NewBuffer(bs[offset+cnt:])
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	a.Type = binary.BigEndian.Uint16(b)
	cnt += 2
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	a.Class = binary.BigEndian.Uint16(b)
	cnt += 2
	if b, err = ReadnBytes(buffer, 4); err != nil {
		return 0, err
	}
	a.TTL = binary.BigEndian.Uint32(b)
	cnt += 4
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	a.RDLength = binary.BigEndian.Uint16(b)
	cnt += 2
	a.RDData, err = ReadnBytes(buffer, int(a.RDLength))
	if err != nil {
		return 0, err
	}
	cnt += int(a.RDLength)
	return cnt, nil
}

func (a *DNSAnswer) Print() {
	// fmt.Printf(
	// 	"AnswerInfo:\nName:%x\nType:%d\nClass:%d\nTLL:%d\nRDLength:%d\nRDData:%x\n\n",
	// 	a.Name,
	// 	a.Type,
	// 	a.Class,
	// 	a.TTL,
	// 	a.RDLength,
	// 	a.RDData,
	// )
	fmt.Printf(
		"AnswerInfo:\nName:%s\nType:%d\nClass:%d\nTLL:%d\nRDLength:%d\nRDData:%x\n\n",
		a.SName(),
		a.Type,
		a.Class,
		a.TTL,
		a.RDLength,
		a.RDData,
	)
}

func (a *DNSAnswer) SName() string {
	var s string
	var n, flag int
	for i := 0; ; i++ {
		n = int(a.Name[i])
		if n == 0 {
			break
		} else {
			if flag == 0 {
				flag = 1
			} else {
				s += string('.')
			}
			for j := 0; j < n; j++ {
				s += string(a.Name[i+1+j])
			}
			i = i + n
		}
	}
	return s
}

func (a *DNSAnswer) ToBytes() []byte {
	var buf bytes.Buffer
	tmp2 := make([]byte, 2)
	tmp4 := make([]byte, 4)
	buf.Write(a.Name)
	binary.BigEndian.PutUint16(tmp2, a.Type)
	buf.Write(tmp2)
	binary.BigEndian.PutUint16(tmp2, a.Class)
	buf.Write(tmp2)
	binary.BigEndian.PutUint32(tmp4, a.TTL)
	buf.Write(tmp4)
	binary.BigEndian.PutUint16(tmp2, a.RDLength)
	buf.Write(tmp2)
	buf.Write(a.RDData)
	return buf.Bytes()
}
