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

type DNSAnswer struct {
	Name     []byte
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RDData   []byte
}

func (a *DNSAnswer) Read(bs []byte) (int, error) {
	buffer := bytes.NewBuffer(bs)
	var tempBuf bytes.Buffer
	var cnt int
	for {
		temp, _ := buffer.ReadByte()
		if temp == 0xc0 {
			tempBuf.WriteByte(temp)
			temp, _ := buffer.ReadByte()
			tempBuf.WriteByte(temp)
			break
		} else {
			tempBuf.WriteByte(temp)
			if temp == 0 {
				break
			}
			for i := 0; i < int(temp); i++ {
				b, _ := buffer.ReadByte()
				tempBuf.WriteByte(b)
			}
		}
	}
	a.Name = tempBuf.Bytes()
	cnt += len(a.Name)
	a.Type = uint16(BytesToInt(ReadnBytes(buffer, 2)))
	cnt += 2
	a.Class = uint16(BytesToInt(ReadnBytes(buffer, 2)))
	cnt += 2
	a.TTL = uint32(BytesToInt(ReadnBytes(buffer, 4)))
	cnt += 4
	a.RDLength = uint16(BytesToInt(ReadnBytes(buffer, 2)))
	cnt += 2
	a.RDData, _ = ReadnBytes(buffer, int(a.RDLength))
	cnt += int(a.RDLength)
	return cnt, nil
}

func (a *DNSAnswer) Print() {
	fmt.Printf(
		"AnswerInfo:\nName:%x\nType:%d\nClass:%d\nTLL:%d\nRDLength:%d\nRDData:%x\n\n",
		a.Name,
		a.Type,
		a.Class,
		a.TTL,
		a.RDLength,
		a.RDData,
	)
}
