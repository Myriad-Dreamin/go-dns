package msg

import (
	"bytes"
	"encoding/binary"
	"fmt"

	QClass "github.com/Myriad-Dreamin/go-dns/msg/rec/qclass"
)

/*
                                        1  1  1  1  1  1
          0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                                               |
        /                     QNAME                     /
        /                                               /
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                     QTYPE                     |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                     QCLASS                    |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

where:

QNAME           a domain name represented as a sequence of labels, where
                each label consists of a length octet followed by that
                number of octets.  The domain name terminates with the
                zero length octet for the null label of the root.  Note
                that this field may be an odd number of octets; no
                padding is used.

QTYPE           a two octet code which specifies the type of the query.
                The values for this field include all codes valid for a
                TYPE field, together with some more general codes which
                can match more than one type of RR.

QCLASS          a two octet code that specifies the class of the query.
                For example, the QCLASS field is IN for the Internet.
*/

type DNSQuestion struct {
	Name  []byte
	Type  uint16
	Class uint16
}

func (q DNSQuestion) Size() uint16 {
	return uint16(len(q.Name) + 4)
}

// assuming len(qnames) == len(qtypes)
func Quest(qnames [][]byte, qtypes []uint16) (ds []DNSQuestion) {
	for idx, qname := range qnames {
		ds = append(ds, DNSQuestion{qname, qtypes[idx], QClass.IN})
	}
	return
}

func (q *DNSQuestion) ReadFrom(bs []byte, offset int) (int, error) {
	var cnt, l int
	var b []byte
	var err error
	if q.Name, l, err = GetFullName(bs, offset); err != nil {
		return 0, err
	}
	cnt += l
	buffer := bytes.NewBuffer(bs[offset+cnt:])
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	q.Type = binary.BigEndian.Uint16(b)
	cnt += 2
	if b, err = ReadnBytes(buffer, 2); err != nil {
		return 0, err
	}
	q.Class = binary.BigEndian.Uint16(b)
	cnt += 2
	return cnt, nil
}

func (q *DNSQuestion) Print() {
	// fmt.Printf(
	// 	"QuestionInfo:\nName:%x\nType:%d\nClass:%d\n\n",
	// 	q.Name,
	// 	q.Type,
	// 	q.Class,
	// )
	fmt.Printf(
		"QuestionInfo:\nName:%s\nType:%d\nClass:%d\n\n",
		q.SName(),
		q.Type,
		q.Class,
	)
}

func (q *DNSQuestion) SName() string {
	var s string
	var n, flag int
	for i := 0; ; i++ {
		n = int(q.Name[i])
		if n == 0 {
			break
		} else {
			if flag == 0 {
				flag = 1
			} else {
				s += string('.')
			}
			for j := 0; j < n; j++ {
				s += string(q.Name[i+1+j])
			}
			i = i + n
		}
	}
	return s
}

func (q *DNSQuestion) ToBytes() []byte {
	var buf bytes.Buffer
	tmp := make([]byte, 2)
	buf.Write(q.Name)
	binary.BigEndian.PutUint16(tmp, q.Type)
	buf.Write(tmp)
	binary.BigEndian.PutUint16(tmp, q.Class)
	buf.Write(tmp)
	return buf.Bytes()
}
