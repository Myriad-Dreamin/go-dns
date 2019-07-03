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
	cnt, _ := m.Header.Read(bs[offset:])
	offset += cnt
	m.Question = make([]DNSQuestion, m.Header.QDCount)
	for i := 0; i < int(m.Header.QDCount); i++ {
		cnt, _ := m.Question[i].ReadFrom(bs, offset)
		offset += cnt
	}
	m.Answer = make([]DNSAnswer, m.Header.ANCount)
	for i := 0; i < int(m.Header.ANCount); i++ {
		cnt, _ := m.Answer[i].ReadFrom(bs, offset)
		offset += cnt
	}
	m.Authority = make([]DNSAnswer, m.Header.NSCount)
	for i := 0; i < int(m.Header.NSCount); i++ {
		cnt, _ := m.Authority[i].ReadFrom(bs, offset)
		offset += cnt
	}
	m.Additional = make([]DNSAnswer, m.Header.ARCount)
	for i := 0; i < int(m.Header.ARCount); i++ {
		cnt, _ := m.Additional[i].ReadFrom(bs, offset)
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
<<<<<<< HEAD

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
=======
>>>>>>> da5afc8a2ac9b75c25edb7df63d7597e3c247518
