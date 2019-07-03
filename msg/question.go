package msg

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
import "bytes"
import "fmt"

type DNSQuestion struct {
	Name  []byte
	Type  uint16
	Class uint16
}

func (q *DNSQuestion) ReadFrom(bs []byte, offset int) (int, error) {
	var cnt, l int
	q.Name, l = GetFullName(bs, offset)
	cnt += l
	buffer := bytes.NewBuffer(bs[offset+cnt:])
	q.Type = uint16(BytesToInt(ReadnBytes(buffer, 2)))
	cnt += 2
	q.Class = uint16(BytesToInt(ReadnBytes(buffer, 2)))
	cnt += 2
	return cnt, nil
}

func (q *DNSQuestion) Print() {
	fmt.Printf(
		"QuestionInfo:\nName:%x\nType:%d\nClass:%d\n\n",
		q.Name,
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
	fmt.Printf(
		"Domain name: %s\n", s)
	return s
}
