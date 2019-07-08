package dnssrv

import (
	"bytes"
	"net"
)

type TCPRemoteServerRoutine struct {
	*sharedSpace
	Buffer        *bytes.Buffer
	MessageChan   chan *bytes.Buffer
	QuitRequest   chan bool
	remoteTCPConn *net.TCPConn
	readNumber    uint16
	network       string
	remoteIP      *net.TCPAddr
}

func NewTCPRemoteServerRoutine(
	sharedSpace *sharedSpace,
	network string,
	addr *net.TCPAddr,
) (gor *TCPRemoteServerRoutine) {
	gor = &TCPRemoteServerRoutine{
		sharedSpace: sharedSpace,
		Buffer:      new(bytes.Buffer),
		MessageChan: make(chan *bytes.Buffer),
		QuitRequest: make(chan bool, 1),
		network:     network,
		remoteIP:    addr,
	}
	var err error
	if gor.remoteTCPConn, err = net.DialTCP(gor.network, nil, gor.remoteIP); err != nil {
		gor.logger.Errorf(
			"error occurred when dial remote dns server: %v\n",
			err,
		)
		return nil
	}
	return gor
}
