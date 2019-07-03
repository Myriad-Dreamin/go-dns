package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unsafe"
)

type Query struct {
	Name  []byte
	Type  int
	Class int
}

type ResourceRecord struct {
	Name       []byte
	Type       int
	Class      int
	TTL        int
	DataLength int
	Data       []byte
}

func ReadName(dns, bs []byte) string {
	var s string
	var n int
	for i := 0; ; i++ {
		n = int(bs[i])
		if n == 0xc0 {
			offset := int(bs[i+1])
			return s + ReadName(dns, dns[offset:])
		} else {
			if n == 0 {
				return s
			}
			for j := 0; j < n; j++ {
				// fmt.Printf("%s", string(bs[i+1+j]))
				s += string(bs[i+1+j])
			}
			i = i + n
			s += "."
		}
	}
	return s
}

func ReadQuery(buffer *bytes.Buffer) Query {
	var qry Query
	var tempBuf bytes.Buffer
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
	qry.Name = tempBuf.Bytes()
	fmt.Println(tempBuf.String())
	b, _ := ReadnBytes(buffer, 2)
	qry.Type = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 2)
	qry.Class = BytesToInt(b)
	return qry
}

func ReadResourceRecord(buffer *bytes.Buffer) ResourceRecord {
	var rr ResourceRecord
	var tempBuf bytes.Buffer
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
	rr.Name = tempBuf.Bytes()
	b, _ := ReadnBytes(buffer, 2)
	rr.Type = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 2)
	rr.Class = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 4)
	rr.TTL = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 2)
	rr.DataLength = BytesToInt(b)
	rr.Data, _ = ReadnBytes(buffer, rr.DataLength)
	return rr
}

type DNSInfo struct {
	TransactionID                       int
	FlagsResponse                       int
	FlagsOpcode                         int
	FlagsAuthoritative                  int
	FlagsTruncated                      int
	FlagsRecursionDesired               int
	FlagsRecursionAvailable             int
	FlagsReserved                       int
	FlagsAnswerAuthenticated            int
	FlagsNonauthenticatedDataAcceptable int
	FlagsReplyCode                      int
	Questions                           int
	AnswerRRs                           int
	AuthorityRRs                        int
	AdditionalRRs                       int
	Queries                             []Query
	Answers                             []ResourceRecord
	AuthoritativeNameservers            []ResourceRecord
	AdditionalRecords                   []ResourceRecord
}

func IsLittleEndian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func ReadnBytes(buffer *bytes.Buffer, n int) ([]byte, error) {
	res := new(bytes.Buffer)
	for i := 0; i < n; i++ {
		c, _ := buffer.ReadByte()
		res.WriteByte(c)
	}
	return res.Bytes(), nil
}

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func BytesToInt(bs []byte) int {
	var res int32
	if IsLittleEndian() {
		for i := 0; i < len(bs); i++ {
			x := uint8(bs[i])
			res = (res << 8) + int32(x)
		}
	} else {
		for i := len(bs) - 1; i >= 0; i-- {
			x := uint8(bs[i])
			res = (res << 8) + int32(x)
		}
	}
	return int(res)
}

func BufferToDNSInfo(byteBuffer []byte) DNSInfo {
	var dns DNSInfo
	buffer := bytes.NewBuffer(byteBuffer)
	b, _ := ReadnBytes(buffer, 2)
	dns.TransactionID = BytesToInt(b)

	b, _ = ReadnBytes(buffer, 2)
	Flags := BytesToInt(b)
	dns.FlagsResponse = Flags >> 15 & 1
	dns.FlagsOpcode = Flags >> 11 & 0x1
	dns.FlagsAuthoritative = Flags >> 10 & 1
	dns.FlagsTruncated = Flags >> 9 & 1
	dns.FlagsRecursionDesired = Flags >> 8 & 1
	dns.FlagsRecursionAvailable = Flags >> 7 & 1
	dns.FlagsReserved = Flags >> 6 & 1
	dns.FlagsAnswerAuthenticated = Flags >> 5 & 1
	dns.FlagsNonauthenticatedDataAcceptable = Flags >> 4 & 1
	dns.FlagsReplyCode = Flags & 0x1

	b, _ = ReadnBytes(buffer, 2)
	dns.Questions = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 2)
	dns.AnswerRRs = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 2)
	dns.AuthorityRRs = BytesToInt(b)
	b, _ = ReadnBytes(buffer, 2)
	dns.AdditionalRRs = BytesToInt(b)

	fmt.Printf("%d %d %d %d\n", dns.Questions, dns.AnswerRRs, dns.AuthorityRRs, dns.AdditionalRRs)

	dns.Queries = make([]Query, dns.Questions)
	for i := 0; i < dns.Questions; i++ {
		dns.Queries[i] = ReadQuery(buffer)
	}
	dns.Answers = make([]ResourceRecord, dns.AnswerRRs)
	for i := 0; i < dns.AnswerRRs; i++ {
		dns.Answers[i] = ReadResourceRecord(buffer)
	}
	dns.AuthoritativeNameservers = make([]ResourceRecord, dns.AuthorityRRs)
	for i := 0; i < dns.AuthorityRRs; i++ {
		dns.AuthoritativeNameservers[i] = ReadResourceRecord(buffer)
	}
	dns.AdditionalRecords = make([]ResourceRecord, dns.AdditionalRRs)
	for i := 0; i < dns.AdditionalRRs; i++ {
		dns.AdditionalRecords[i] = ReadResourceRecord(buffer)
	}
	return dns
}

func main() {
	// hexstr := "7c4c010000010000000000000377777706676f6f676c6503636f6d0000010001"
	hexstr := "59ad818000010003000000000377777705626169647503636f6d0000010001c00c0005000100000005000f0377777701610673686966656ec016c02b000100010000000500040ed7b127c02b000100010000000500040ed7b126"
	byteSlice, _ := hex.DecodeString(hexstr)
	fmt.Println(byteSlice)
	// buffer := bytes.NewBuffer(byteSlice) // 创建20字节缓冲区 len = 20 off = 0
	res := BufferToDNSInfo(byteSlice)
	// fmt.Printf("%x\n", res.TransactionID)
	// fmt.Println(IsLittleEndian())
	s := ReadName(byteSlice, res.Answers[0].Name)
	fmt.Println(s)
}
