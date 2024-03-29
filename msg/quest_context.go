package msg

type QuestContext struct {
	Message      DNSMessage
	additionSize uint16
}

func (m *QuestContext) PacketQuestion(question []DNSQuestion) int {
	for i, q := range question {
		if !m.InsertQuestion(q) {
			return i
		}
	}
	return len(question)
}

func (m *QuestContext) InsertQuestion(que DNSQuestion) bool {
	var qs = que.Size()
	if m.additionSize+qs > maxMessageSize {
		return false
	}
	m.Message.InsertQuestion(que)
	m.additionSize += qs
	return true
}

func (m *QuestContext) InsertAnswer(ans DNSAnswer) bool {
	var qs = ans.Size()
	if m.additionSize+qs > maxMessageSize {
		return false
	}
	m.Message.InsertAnswer(ans)
	m.additionSize += qs
	return true
}

func (m *QuestContext) InsertAuthority(ans DNSAnswer) bool {
	var qs = ans.Size()
	if m.additionSize+qs > maxMessageSize {
		return false
	}
	m.Message.InsertAuthority(ans)
	m.additionSize += qs
	return true
}

func (m *QuestContext) InsertAdditional(ans DNSAnswer) bool {
	var qs = ans.Size()
	if m.additionSize+qs > maxMessageSize {
		return false
	}
	m.Message.InsertAdditional(ans)
	m.additionSize += qs
	return true
}
