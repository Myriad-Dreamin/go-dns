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
		cnt, _ := m.Question[i].Read(bs[offset:])
		offset += cnt
	}
	m.Answer = make([]DNSAnswer, m.Header.ANCount)
	for i := 0; i < int(m.Header.ANCount); i++ {
		cnt, _ := m.Answer[i].Read(bs[offset:])
		offset += cnt
	}
	m.Authority = make([]DNSAnswer, m.Header.NSCount)
	for i := 0; i < int(m.Header.NSCount); i++ {
		cnt, _ := m.Authority[i].Read(bs[offset:])
		offset += cnt
	}
	m.Additional = make([]DNSAnswer, m.Header.ARCount)
	for i := 0; i < int(m.Header.ARCount); i++ {
		cnt, _ := m.Additional[i].Read(bs[offset:])
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
