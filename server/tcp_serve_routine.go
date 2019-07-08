package dnssrv

import "bytes"

type TCPUserRoutine struct {
	*sharedSpace
	Buffer      *bytes.Buffer
	QuitRequest chan bool
	MessageChan chan *bytes.Buffer
	RequestChan chan *bytes.Buffer
	tid         uint16
	readNumber  uint16
}

func NewTCPUserRoutine(
	sharedSpace *sharedSpace,
	tid uint16,
) *TCPUserRoutine {
	return &TCPUserRoutine{
		tid:         tid,
		sharedSpace: sharedSpace,
		Buffer:      new(bytes.Buffer),
		MessageChan: make(chan *bytes.Buffer),
		RequestChan: make(chan *bytes.Buffer),
		QuitRequest: make(chan bool, 1),
	}
}
