package msg

/*
                                        1  1  1  1  1  1
          0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                      ID                       |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                    QDCOUNT                    |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                    ANCOUNT                    |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                    NSCOUNT                    |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
        |                    ARCOUNT                    |
        +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

where:

ID              A 16 bit identifier assigned by the program that
                generates any kind of query.  This identifier is copied
                the corresponding reply and can be used by the requester
                to match up replies to outstanding queries.

QR              A one bit field that specifies whether this message is a
                query (0), or a response (1).

OPCODE          A four bit field that specifies kind of query in this
                message.  This value is set by the originator of a query
                and copied into the response.  The values are:

                0               a standard query (QUERY)

                1               an inverse query (IQUERY)

                2               a server status request (STATUS)

                3-15            reserved for future use

AA              Authoritative Answer - this bit is valid in responses,
                and specifies that the responding name server is an
                authority for the domain name in question section.

                Note that the contents of the answer section may have
                multiple owner names because of aliases.  The AA bit
                corresponds to the name which matches the query name, or
                the first owner name in the answer section.

TC              TrunCation - specifies that this message was truncated
                due to length greater than that permitted on the
                transmission channel.

RD              Recursion Desired - this bit may be set in a query and
                is copied into the response.  If RD is set, it directs
                the name server to pursue the query recursively.
                Recursive query support is optional.

RA              Recursion Available - this be is set or cleared in a
                response, and denotes whether recursive query support is
                available in the name server.

Z               Reserved for future use.  Must be zero in all queries
                and responses.

RCODE           Response code - this 4 bit field is set as part of
                responses.  The values have the following
                interpretation:

                0               No error condition

                1               Format error - The name server was
                                unable to interpret the query.

                2               Server failure - The name server was
                                unable to process this query due to a
                                problem with the name server.

                3               Name Error - Meaningful only for
                                responses from an authoritative name
                                server, this code signifies that the
                                domain name referenced in the query does
                                not exist.

                4               Not Implemented - The name server does
                                not support the requested kind of query.

                5               Refused - The name server refuses to
                                perform the specified operation for
                                policy reasons.  For example, a name
                                server may not wish to provide the
                                information to the particular requester,
                                or a name server may not wish to perform
                                a particular operation (e.g., zone
                                transfer) for particular data.

                6-15            Reserved for future use.

QDCOUNT         an unsigned 16 bit integer specifying the number of
                entries in the question section.

ANCOUNT         an unsigned 16 bit integer specifying the number of
                resource records in the answer section.

NSCOUNT         an unsigned 16 bit integer specifying the number of name
                server resource records in the authority records
                section.

ARCOUNT         an unsigned 16 bit integer specifying the number of
                resource records in the additional records section.
*/
import "bytes"
import "fmt"

type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

// func (h *DNSHeader) Read(r io.ReadWriter) error {
// 	var bs []byte
// 	r.Read(bs)
// 	buf := bytes.NewBuffer(bs)
// 	h.ID = uint16(BytesToInt(ReadnBytes(buf, 2)))
// 	h.Flags = uint16(BytesToInt(ReadnBytes(buf, 2)))
// 	h.QDCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
// 	h.ANCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
// 	h.NSCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
// 	h.ARCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
// 	return nil
// }

func (h *DNSHeader) Read(bs []byte) (int, error) {
	buf := bytes.NewBuffer(bs)
	h.ID = uint16(BytesToInt(ReadnBytes(buf, 2)))
	h.Flags = uint16(BytesToInt(ReadnBytes(buf, 2)))
	h.QDCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
	h.ANCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
	h.NSCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
	h.ARCount = uint16(BytesToInt(ReadnBytes(buf, 2)))
	return 12, nil
}

func (h *DNSHeader) Print() {
	fmt.Printf(
		"HeaderInfo:\nID:%x\nFlags:%x\nquc:%x\nanc:%x\nauc:%x\nadc:%x\n\n",
		h.ID,
		h.Flags,
		h.QDCount,
		h.ANCount,
		h.NSCount,
		h.ARCount,
	)
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
	for i := 0; i < n; i++ {
		c, _ := buffer.ReadByte()
		res.WriteByte(c)
	}
	return res.Bytes(), nil
}
