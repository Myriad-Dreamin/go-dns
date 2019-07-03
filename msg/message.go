package msg

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
import "bytes"

type DNSMessage struct {
	Header     DNSHeader
	Question   []DNSQuestion
	Answer     []DNSAnswer
	Authority  []DNSAnswer
	Additional []DNSAnswer
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

func (m *DNSMessage) ToBytes() []byte {
	var buf bytes.Buffer
	buf.Write(m.Header.ToBytes())
	for i := 0; i < int(m.Header.QDCount); i++ {
		buf.Write(m.Question[i].ToBytes())
	}
	for i := 0; i < int(m.Header.ANCount); i++ {
		buf.Write(m.Answer[i].ToBytes())
	}
	for i := 0; i < int(m.Header.NSCount); i++ {
		buf.Write(m.Authority[i].ToBytes())
	}
	for i := 0; i < int(m.Header.ARCount); i++ {
		buf.Write(m.Additional[i].ToBytes())
	}
	return buf.Bytes()
}
