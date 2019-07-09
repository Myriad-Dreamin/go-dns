package dnssrv

import (
	"fmt"
	"net"
	"time"

	msg "github.com/Myriad-Dreamin/go-dns/msg"
	mredis "github.com/Myriad-Dreamin/go-dns/redis"
	log "github.com/sirupsen/logrus"
)

type UDPDispatcher struct {
	logger  *log.Entry
	connUDP *net.UDPConn

	bytesPool  *BytesPool
	bufferPool *BufferPool

	remoteUDPConn *net.UDPConn
	udpConnected  bool

	// stateless udp limited by chan resource
	UDPRoutineLimit chan uint16
	UDPBuffer       [][]byte

	// stateless udp limited by chan resource
	UDPReadRoutineLimit chan uint16
	UDPReadBuffer       [][]byte
	UDPReadBytesChan    []chan uint16

	tidL     uint16
	tidR     uint16
	udpRange uint16
	closing  bool
}

func NewUDPDispatcher(
	logger *log.Entry,
	maxSize int64,
	idRangeL, idRangeR, udpRange uint16,
) (up *UDPDispatcher) {
	if udpRange != idRangeR-idRangeL {
		return
	}
	up = &UDPDispatcher{
		logger:           logger,
		bytesPool:        NewBytesPool(maxSize),
		bufferPool:       NewBufferPool(),
		tidL:             idRangeL,
		tidR:             idRangeR,
		udpRange:         udpRange,
		UDPBuffer:        make([][]byte, udpRange),
		UDPReadBuffer:    make([][]byte, udpRange),
		UDPReadBytesChan: make([]chan uint16, udpRange),
	}
	return
}

func (udpDispatcher *UDPDispatcher) tryConnectToRemoteDNSServer(network string, host *net.UDPAddr) (err error) {
	if err != nil {
		udpDispatcher.logger.Errorf("error occurred when resolving remote dns server ip: %v\n", err)
	}
	udpDispatcher.remoteUDPConn, err = net.DialUDP(network, nil, host)
	if err != nil {
		udpDispatcher.logger.Errorf("error occurred when dial remote udp DNS Server: %v\n", err)
	}
	return
}

func (udpDispatcher *UDPDispatcher) tryDisonnectFromRemoteDNSServer() error {
	if !udpDispatcher.udpConnected {
		return nil
	}

	udpDispatcher.logger.Infof("disconnected from remote udp DNS Server")
	return udpDispatcher.remoteUDPConn.Close()
}

func (udpDispatcher *UDPDispatcher) listenUDP() (err error) {
	var udpAddr *net.UDPAddr
	udpAddr, err = net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		udpDispatcher.logger.Errorf("resolve local udp server address error: %v", err)
		return
	}
	udpDispatcher.connUDP, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		udpDispatcher.logger.Errorf("setup local udp server error: %v", err)
		return
	}
	return
}

func (udpDispatcher *UDPDispatcher) Prepare(network string, host *net.UDPAddr) (err error) {

	// set up udp connection(dial test)
	if err = udpDispatcher.tryConnectToRemoteDNSServer(network, host); err != nil {
		udpDispatcher.logger.Errorf("udp socket set up failed, error: %v", err)
		return
	} else {
		udpDispatcher.logger.Infof("udp socket set up successfully")
		udpDispatcher.udpConnected = true
	}

	err = udpDispatcher.listenUDP()
	if err != nil {
		udpDispatcher.logger.Errorf("udp server set up failed, error: %v", err)
		return
	}
	udpDispatcher.UDPRoutineLimit = make(chan uint16, UDPRange)
	udpDispatcher.UDPReadRoutineLimit = make(chan uint16, UDPRange)

	for idx := uint16(0); idx < UDPRange; idx++ {
		udpDispatcher.UDPRoutineLimit <- idx
		udpDispatcher.UDPReadRoutineLimit <- idx
		udpDispatcher.UDPBuffer[idx] = make([]byte, UDPBufferSize)
		udpDispatcher.UDPReadBuffer[idx] = make([]byte, UDPBufferSize)
		udpDispatcher.UDPReadBytesChan[idx] = make(chan uint16, 1)
	}
	return nil
}

func (udpDispatcher *UDPDispatcher) Start(qc *chan bool) (err error) {
	for !udpDispatcher.closing {
		select {
		case idx := <-udpDispatcher.UDPRoutineLimit:
			// fmt.Println("r", idx)
			go udpDispatcher.ServeUDPFromOut(idx, udpDispatcher.UDPBuffer[idx])
		case idx := <-udpDispatcher.UDPReadRoutineLimit:
			// fmt.Println("s", idx)
			go udpDispatcher.ServeUDPReadFromOut(idx, udpDispatcher.UDPReadBuffer[idx])
			// case idx := <-srv.TCPRoutineLimit:
			// 	go srv.ServeTCPFromOut(idx, srv.TCPBuffer[idx-UDPRange])
			// case idx := <-srv.TCPWriteRoutineLimit:
			// 	go srv.ServeTCPWriteToOut(idx)
			// case idx := <-srv.TCPReadRoutineLimit:
			// 	go srv.ServeTCPReadFromOut(idx, srv.TCPReadBuffer[idx-UDPRange])
		}
	}
	return
}

func (udpDispatcher *UDPDispatcher) AtExit() {
	if err := udpDispatcher.Stop(); err != nil {
		udpDispatcher.logger.Errorf("error occurred when stopping  dispatcher, error: %v", err)
	}
	udpDispatcher.logger.Infof("udp server stop successfully")
}

func (udpDispatcher *UDPDispatcher) Stop() (err error) {
	udpDispatcher.closing = true
	var aa, bb = 0, 0
	for i := int32(udpDispatcher.udpRange) * 2; i > 0; i-- {
		select {
		case <-udpDispatcher.UDPRoutineLimit:
			aa++
			fmt.Println("UDPRoutineLimit", i, aa)
		case <-udpDispatcher.UDPReadRoutineLimit:
			bb++
			fmt.Println("UDPReadRoutineLimit", i, bb)
		}
	}
	err = udpDispatcher.connUDP.Close()
	if err != nil {
		udpDispatcher.logger.Errorf("error occurred when close udp-server connection, error: %v", err)
	}
	return udpDispatcher.tryDisonnectFromRemoteDNSServer()
}

func (udpDispatcher *UDPDispatcher) ReleaseUDPRoutine(tid uint16) {
	udpDispatcher.UDPRoutineLimit <- tid
}

func (udpDispatcher *UDPDispatcher) ReleaseUDPReadRoutine(tid uint16) {
	udpDispatcher.UDPReadRoutineLimit <- tid
}

func (udpDispatcher *UDPDispatcher) ServeUDPReadFromOut(tid uint16, b []byte) {
	udpDispatcher.remoteUDPConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	for {
		if udpDispatcher.closing {
			udpDispatcher.ReleaseUDPReadRoutine(tid)
			return
		}
		_, err := udpDispatcher.remoteUDPConn.Read(b)
		// udpDispatcher.udpDispatcherMutex.Lock()
		// defer udpDispatcher.udpDispatcherMutex.Unlock()
		if err != nil {
			if er, ok := err.(net.Error); !ok {
				udpDispatcher.logger.Errorf("read error: %v", err)
				return
			} else {
				if udpDispatcher.closing {
					udpDispatcher.ReleaseUDPReadRoutine(tid)
					return
				}
				if er.Timeout() {
					udpDispatcher.remoteUDPConn.SetReadDeadline(time.Now().Add(1 * time.Second))
					continue
				}
			}
		}
		// fast extract id from message
		udpDispatcher.UDPReadBytesChan[(uint16(b[0])<<8)+uint16(b[1])] <- tid
		return
	}
}

func (udpDispatcher *UDPDispatcher) ServeUDPFromOut(tid uint16, b []byte) {
	defer udpDispatcher.ReleaseUDPRoutine(tid)
	udpDispatcher.connUDP.SetReadDeadline(time.Now().Add(1 * time.Second))
	for {

		for len(udpDispatcher.UDPReadBytesChan[tid]) > 0 {
			<-udpDispatcher.UDPReadBytesChan[tid]
			fmt.Println("QAQ")
		}
		if udpDispatcher.closing {
			return
		}
		_, servingAddr, err := udpDispatcher.connUDP.ReadFromUDP(b)
		if err != nil {
			if er, ok := err.(net.Error); !ok {
				udpDispatcher.logger.Errorf("failed when reading udp msg, error: %v", err)
				return
			} else {
				if udpDispatcher.closing {
					return
				}
				if er.Timeout() {
					udpDispatcher.connUDP.SetReadDeadline(time.Now().Add(1 * time.Second))
					continue
				}
			}
		}
		udpDispatcher.connUDP.SetReadDeadline(time.Now().Add(1 * time.Second))

		var message msg.DNSMessage
		_, err = message.Read(b)
		if err != nil {
			udpDispatcher.logger.Errorf("failed when decoding udp msg, error: %v", err)
			return
		}

		udpDispatcher.logger.Infof("new message incoming: id, address: %v, %v", message.Header.ID, servingAddr)

		conn := mredis.RedisCacheClient.Pool.Get()
		defer conn.Close()

		reply := msg.NewDNSMessageReply(message.Header.ID, message.Header.Flags, message.Question)
		if mredis.FindCache(reply, conn) {
			// reply.Print()
			b, err := reply.CompressToBytes()
			if err != nil {
				udpDispatcher.logger.Errorf("get redis cache error: %v", err)
				return
			}
			_, err = udpDispatcher.connUDP.WriteToUDP(b, servingAddr)
			if err != nil {
				udpDispatcher.logger.Errorf("write to client error: %v", err)
				return
			}

			udpDispatcher.logger.Infof("using redis cache reply to address: %v, %v", message.Header.ID, servingAddr)
			return
		} else {
			fid := message.Header.ID
			message.Header.ID = tid
			b, err = message.CompressToBytes()
			// b[0] = byte(tid >> 8)
			// b[1] = byte(tid & 0xff)
			if err != nil {
				udpDispatcher.logger.Errorf("convert error: %v", err)
				return
			}

			if _, err := udpDispatcher.remoteUDPConn.Write(b); err != nil {
				udpDispatcher.logger.Errorf("write error: %v", err)
				return
			}

			select {
			case rid := <-udpDispatcher.UDPReadBytesChan[tid]:
				defer udpDispatcher.ReleaseUDPReadRoutine(rid)
				b = udpDispatcher.UDPReadBuffer[rid]
			case <-time.After(time.Second * 1):
				//todo: reply...
			}

			_, err = message.Read(b)
			if err != nil {
				udpDispatcher.logger.Errorf("read error: %v", err)
				return
			}
			// if tid != message.Header.ID {
			// 	udpDispatcher.logger.Errorf("not matching..., serving %v", servingAddr)
			// }
			// message.Print()
			message.Header.ID = fid

			mredis.MessageToRedis(message, conn)

			b, err = message.CompressToBytes()

			// b[0] = byte(fid >> 8)
			// b[1] = byte(fid & 0xff)
			if err != nil {
				udpDispatcher.logger.Errorf("convert error: %v", err)
				return
			}
			_, err = udpDispatcher.connUDP.WriteToUDP(b, servingAddr)
			if err != nil {
				udpDispatcher.logger.Errorf("write to client error: %v", err)
				return
			}

			udpDispatcher.logger.Infof("reply to address: %v, %v", message.Header.ID, servingAddr)
			return
		}
	}
}
