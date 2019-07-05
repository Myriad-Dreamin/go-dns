package msg

import (
	"bytes"
	flags "github.com/Myriad-Dreamin/go-dns/msg/flags"
)

const (
	// reserve 62 bytes for Name Segment
	maxMessageSize = 450
)

/*
   +---------------------+
   |        Header       |
   +---------------------+
   |       Question      |
   +---------------------+
   |        Answer       |
   +---------------------+
   |      Authority      |
   +---------------------+
   |      Additional     |
   +---------------------+
*/

type DNSMessage struct {
	Header     DNSHeader
	Question   []DNSQuestion
	Answer     []DNSAnswer
	Authority  []DNSAnswer
	Additional []DNSAnswer
}

// Assuming Empty Message
func (m *DNSMessage) InitQuery(msgid uint16) {
	m.Header.ID = msgid
}

// Assuming Empty Message
func (m *DNSMessage) InitRecursivelyQuery(msgid uint16) {
	m.Header.ID = msgid
	m.Header.Flags = flags.RD
}

func NewEmptyDNSMessageQuery(msgid uint16) (m *DNSMessage) {
	m = new(DNSMessage)
	m.InitQuery(msgid)
	return
}

func NewDNSMessageQuery(msgid uint16, que []DNSQuestion) (n int, m *DNSMessage) {
	c := new(QuestContext)
	c.Message.InitQuery(msgid)
	c.additionSize = c.Message.Header.Size()
	n = c.PacketQuestion(que)
	m = &c.Message
	return
}

func NewEmptyDNSMessageRecursivelyQuery(msgid uint16) (m *DNSMessage) {
	m = new(DNSMessage)
	m.InitRecursivelyQuery(msgid)
	return
}

func NewDNSMessageRecursivelyQuery(msgid uint16, que []DNSQuestion) (n int, m *DNSMessage) {
	c := new(QuestContext)
	c.Message.InitRecursivelyQuery(msgid)
	c.additionSize = c.Message.Header.Size()
	n = c.PacketQuestion(que)
	m = &c.Message
	return
}

func (m *DNSMessage) InsertQuestion(que ...DNSQuestion) {
	m.Question = append(m.Question, que...)
	m.Header.QDCount += uint16(len(que))
}

func (m *DNSMessage) InsertAnswer(ans ...DNSAnswer) {
	m.Answer = append(m.Answer, ans...)
	m.Header.ANCount += uint16(len(ans))
}

func (m *DNSMessage) InsertAuthority(ans ...DNSAnswer) {
	m.Authority = append(m.Authority, ans...)
	m.Header.NSCount += uint16(len(ans))
}

func (m *DNSMessage) InsertAdditional(ans ...DNSAnswer) {
	m.Additional = append(m.Additional, ans...)
	m.Header.ARCount += uint16(len(ans))
}

func (m *DNSMessage) Read(bs []byte) (int, error) {
	var offset int
	cnt, err := m.Header.Read(bs[offset:])
	if err != nil {
		return offset, err
	}
	offset += cnt
	m.Question = make([]DNSQuestion, m.Header.QDCount)
	for i := 0; i < int(m.Header.QDCount); i++ {
		cnt, err := m.Question[i].ReadFrom(bs, offset)
		if err != nil {
			return offset, err
		}
		offset += cnt
	}
	m.Answer = make([]DNSAnswer, m.Header.ANCount)
	for i := 0; i < int(m.Header.ANCount); i++ {
		cnt, err := m.Answer[i].ReadFrom(bs, offset)
		if err != nil {
			return offset, err
		}
		offset += cnt
	}
	m.Authority = make([]DNSAnswer, m.Header.NSCount)
	for i := 0; i < int(m.Header.NSCount); i++ {
		cnt, err := m.Authority[i].ReadFrom(bs, offset)
		if err != nil {
			return offset, err
		}
		offset += cnt
	}
	m.Additional = make([]DNSAnswer, m.Header.ARCount)
	for i := 0; i < int(m.Header.ARCount); i++ {
		cnt, err := m.Additional[i].ReadFrom(bs, offset)
		if err != nil {
			return offset, err
		}
		offset += cnt
	}
	return offset, nil
}

func (m *DNSMessage) Print() {
	m.Header.Print()
	for _, it := range m.Question {
		it.Print()
	}
	for _, it := range m.Answer {
		it.Print()
	}
	for _, it := range m.Authority {
		it.Print()
	}
	for _, it := range m.Additional {
		it.Print()
	}
}

func (m *DNSMessage) ToBytes() (b []byte, err error) {
	var buf bytes.Buffer
	buf.Write(m.Header.ToBytes())
	for i := 0; i < int(m.Header.QDCount); i++ {
		b, err = m.Question[i].ToBytes()
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	for i := 0; i < int(m.Header.ANCount); i++ {
		b, err := m.Answer[i].ToBytes()
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	for i := 0; i < int(m.Header.NSCount); i++ {
		b, err := m.Authority[i].ToBytes()
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	for i := 0; i < int(m.Header.ARCount); i++ {
		b, err := m.Additional[i].ToBytes()
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	return buf.Bytes(), nil
}

func (m *DNSMessage) CompressToBytes() ([]byte, error) {
	return NewReplyContext(m).Bytes()
}
