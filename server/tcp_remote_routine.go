package dnssrv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
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
	Repeat        int64
	RepeatTimeOut time.Duration
}

func (rt *TCPRemoteServerRoutine) RequestQuit() bool {
	if rt == nil {
		return false
	}
	rt.QuitRequest <- true
	return true
}

func NewTCPRemoteServerRoutine(
	sharedSpace *sharedSpace,
	network string,
	addr *net.TCPAddr,
) (rt *TCPRemoteServerRoutine) {
	rt = &TCPRemoteServerRoutine{
		sharedSpace:   sharedSpace,
		Buffer:        sharedSpace.bufferPool.Get().(*bytes.Buffer),
		MessageChan:   make(chan *bytes.Buffer),
		QuitRequest:   make(chan bool, 1),
		network:       network,
		remoteIP:      addr,
		Repeat:        2,
		RepeatTimeOut: 2 * time.Second,
	}
	var err error
	if rt.remoteTCPConn, err = net.DialTCP(rt.network, nil, rt.remoteIP); err != nil {
		rt.logger.Errorf(
			"error occurred when dial remote dns server: %v\n",
			err,
		)
		return nil
	}
	return rt
}

func (rt *TCPRemoteServerRoutine) reDial() (err error) {
	if rt.remoteTCPConn, err = net.DialTCP(rt.network, nil, rt.remoteIP); err != nil {
		rt.logger.Errorf(
			"error occurred when dial remote dns server: %v\n",
			err,
		)
	}
	return
}

func (rt *TCPRemoteServerRoutine) tryReDial() (err error) {
	var qwq = rt.Repeat
	for ; qwq > 0; qwq-- {
		if err := rt.reDial(); err == nil {
			break
		}
		time.Sleep(rt.RepeatTimeOut)
	}
	return
}

func (rt *TCPRemoteServerRoutine) Run() {
	if rt == nil {
		return
	}
	b := rt.bytesPool.Get().([]byte)
	rt.Buffer.Reset()
	for {
		select {
		case <-rt.QuitRequest:
			rt.quit <- true
			fmt.Println("TCPRemoteServerRoutine")
			return
		case bmsg := <-rt.MessageChan:
			fmt.Println("ready to write")
			rt.remoteTCPConn.SetWriteDeadline(time.Now().Add(1 * time.Second))
			err := binary.Write(rt.remoteTCPConn, binary.BigEndian, uint16(bmsg.Len()))
			if err != nil {
				rt.logger.Errorf("write len error: %v", err)
				rt.dispatcher.messageChan <- bmsg
			}
			_, err = rt.remoteTCPConn.Write(bmsg.Bytes())
			if err != nil {
				rt.logger.Errorf("write bmsg error: %v", err)
				rt.dispatcher.messageChan <- bmsg
			}
			rt.bufferPool.Put(bmsg)
		default:

			rt.remoteTCPConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			_, err := rt.remoteTCPConn.Read(b)
			if err != nil {
				if err == io.EOF {
					if err = rt.tryReDial(); err != nil {
						rt.logger.Errorf("redial failed, error: %v", err)
						return
					} else {
						continue
					}
				}
				if er, ok := err.(net.Error); !ok {
					rt.logger.Errorf("failed when reading message, error: %v", err)
				} else if er.Timeout() {
					if len(rt.QuitRequest) != 0 {
						rt.quit <- true
						return
					}
					rt.remoteTCPConn.SetDeadline(time.Now().Add(1 * time.Second))
				} else {
					rt.logger.Errorf("failed when reading message, error: %v %v", err, er.Temporary())
				}
				continue
			}
			_, err = rt.Buffer.Write(b)
			if err != nil {
				rt.logger.Errorf("buffering error: %v", err)
				rt.remoteTCPConn.Close()
				if err = rt.tryReDial(); err != nil {
					rt.logger.Errorf("redial failed, error: %v", err)
					return
				}
				continue
			}
			for {
				if rt.readNumber != 0 {
					if rt.Buffer.Len() >= int(rt.readNumber) {
						bb := rt.bufferPool.Get().(*bytes.Buffer)
						bb.Reset()
						var tid uint16
						binary.Read(rt.Buffer, binary.BigEndian, &tid)
						binary.Write(bb, binary.BigEndian, &tid)
						_, err := io.TeeReader(io.LimitReader(rt.Buffer, int64(rt.readNumber-2)), bb).Read(b)

						fmt.Println("getting...", tid, bb)
						rt.readNumber = 0
						if err != nil {
							rt.logger.Errorf("trans buffering error: %v", err)
							continue
						}

						rt.dispatcher.tcpUserRoutine[tid].MessageChan <- bb

					} else {
						break
					}
				} else if rt.Buffer.Len() > 1 {
					binary.Read(rt.Buffer, binary.BigEndian, &rt.readNumber)
				} else {
					break
				}
			}
			if rt.Buffer.Cap() < 600 {
				var t = rt.bufferPool.Get().(*bytes.Buffer)
				_, err = rt.Buffer.WriteTo(t)
				if err != nil {
					rt.logger.Errorf("convert buffering error: %v", err)
					rt.bufferPool.Put(t)
					continue
				} else {
					rt.Buffer.Reset()
					rt.bufferPool.Put(rt.Buffer)
					rt.Buffer = t
				}
			}

			// case idx := <-srv.TCPRoutineLimit:
			// 	go srv.ServeTCPFromOut(idx, srv.TCPBuffer[idx-UDPRange])
			// case idx := <-srv.TCPWriteRoutineLimit:
			// 	go srv.ServeTCPWriteToOut(idx)
			// case idx := <-srv.TCPReadRoutineLimit:
			// 	go srv.ServeTCPReadFromOut(idx, srv.TCPReadBuffer[idx-UDPRange])
		}
	}
}
