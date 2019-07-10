package dnssrv

import (
	"fmt"
	"net"
	"time"

	hosts "github.com/Myriad-Dreamin/go-dns/hosts"
	msg "github.com/Myriad-Dreamin/go-dns/msg"
	qtype "github.com/Myriad-Dreamin/go-dns/msg/rec/qtype"
	mredis "github.com/Myriad-Dreamin/go-dns/redis"
	log "github.com/sirupsen/logrus"
)

type UDPDispatcher struct {
	logger  *log.Entry
	connUDP *net.UDPConn

	bytesPool  *BytesPool
	bufferPool *BufferPool

	remoteUDPConn []*net.UDPConn
	udpConnected  []bool

	// stateless udp limited by chan resource
	UDPRoutineLimit chan int64
	UDPBuffer       [][]byte

	// stateless udp limited by chan resource
	UDPReadRoutineLimit chan int64
	UDPReadBuffer       [][]byte
	UDPReadBytesChan    []chan []byte

	maxRoutineCount int64
	closing         bool
}

func NewUDPDispatcher(
	logger *log.Entry,
	maxSize int64,
	maxRoutineCount int64,
) (up *UDPDispatcher) {
	var bp = NewBytesPool(maxSize)
	up = &UDPDispatcher{
		logger:              logger,
		bytesPool:           bp,
		bufferPool:          NewBufferPool(bp),
		maxRoutineCount:     maxRoutineCount,
		udpConnected:        make([]bool, maxRoutineCount),
		remoteUDPConn:       make([]*net.UDPConn, maxRoutineCount),
		UDPReadRoutineLimit: make(chan int64, maxRoutineCount),
		UDPRoutineLimit:     make(chan int64, maxRoutineCount),
		UDPBuffer:           make([][]byte, maxRoutineCount),
		UDPReadBuffer:       make([][]byte, maxRoutineCount),
		UDPReadBytesChan:    make([]chan []byte, maxRoutineCount),
	}
	return
}

func (udpDispatcher *UDPDispatcher) tryConnectToRemoteDNSServer(idx int64, network string, host *net.UDPAddr) (err error) {
	// if udpDispatcher.udpConnected[idx] {
	// 	return nil
	// }
	if err != nil {
		udpDispatcher.logger.Errorf("error occurred when resolving remote dns server ip: %v\n", err)
	}
	udpDispatcher.remoteUDPConn[idx], err = net.DialUDP(network, nil, host)
	if err != nil {
		udpDispatcher.logger.Errorf("error occurred when dial remote udp DNS Server: %v\n", err)
	}
	return
}

func (udpDispatcher *UDPDispatcher) tryDisonnectFromRemoteDNSServer(idx int64) error {
	if !udpDispatcher.udpConnected[idx] {
		return nil
	}

	udpDispatcher.logger.Infof("disconnected from remote udp DNS Server")
	if err := udpDispatcher.remoteUDPConn[idx].Close(); err != nil {
		return err
	}
	udpDispatcher.udpConnected[idx] = false
	return nil
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
	for idx := int64(0); idx < udpDispatcher.maxRoutineCount; idx++ {
		if err = udpDispatcher.tryConnectToRemoteDNSServer(idx, network, host); err != nil {
			udpDispatcher.logger.Errorf("udp socket set up failed, error: %v", err)
			return
		} else {
			udpDispatcher.logger.Infof("udp socket set up successfully")
			udpDispatcher.udpConnected[idx] = true
		}
	}

	err = udpDispatcher.listenUDP()
	if err != nil {
		udpDispatcher.logger.Errorf("udp server set up failed, error: %v", err)
		return
	}

	for idx := int64(0); idx < udpDispatcher.maxRoutineCount; idx++ {
		udpDispatcher.UDPRoutineLimit <- idx
		udpDispatcher.UDPReadRoutineLimit <- idx
		udpDispatcher.UDPReadBytesChan[idx] = make(chan []byte, 1)
	}
	return nil
}

func (udpDispatcher *UDPDispatcher) Start(qc *chan bool) (err error) {
	for !udpDispatcher.closing {
		select {
		case idx := <-udpDispatcher.UDPRoutineLimit:
			// fmt.Println("r", idx)
			go udpDispatcher.ServeUDPFromOut(idx)
		case idx := <-udpDispatcher.UDPReadRoutineLimit:
			// fmt.Println("s", idx)
			go udpDispatcher.ServeUDPReadFromOut(idx)
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
	var qwq = make(chan bool, 2)
	go func() {
		for i := udpDispatcher.maxRoutineCount; i > 0; i-- {
			select {
			case <-udpDispatcher.UDPRoutineLimit:
				aa++
				fmt.Println("UDPRoutineLimit", i, aa)
			}
			qwq <- true
		}
	}()
	go func() {
		for i := udpDispatcher.maxRoutineCount; i > 0; i-- {
			select {
			case <-udpDispatcher.UDPReadRoutineLimit:
				bb++
				fmt.Println("UDPReadRoutineLimit", i, bb)
			}
			qwq <- true
		}
	}()
	<-qwq
	<-qwq
	err = udpDispatcher.connUDP.Close()
	if err != nil {
		udpDispatcher.logger.Errorf("error occurred when close udp-server connection, error: %v", err)
	}
	for i := udpDispatcher.maxRoutineCount - 1; i >= 0; i-- {
		udpDispatcher.tryDisonnectFromRemoteDNSServer(i)
	}
	return
}

func (udpDispatcher *UDPDispatcher) ReleaseUDPRoutine(idx int64) {
	udpDispatcher.UDPRoutineLimit <- idx
}

func (udpDispatcher *UDPDispatcher) ReleaseUDPReadRoutine(idx int64) {
	udpDispatcher.UDPReadRoutineLimit <- idx
}

func (udpDispatcher *UDPDispatcher) ServeUDPReadFromOut(idx int64) {
	b := udpDispatcher.bytesPool.Get().([]byte)
	udpDispatcher.remoteUDPConn[idx].SetReadDeadline(time.Now().Add(1 * time.Second))
	for {
		if udpDispatcher.closing {
			udpDispatcher.ReleaseUDPReadRoutine(idx)
			return
		}
		_, err := udpDispatcher.remoteUDPConn[idx].Read(b)
		// udpDispatcher.udpDispatcherMutex.Lock()
		// defer udpDispatcher.udpDispatcherMutex.Unlock()
		if err != nil {
			if er, ok := err.(net.Error); !ok {
				udpDispatcher.logger.Errorf("read error: %v", err)
				continue
				//return
			} else {
				if udpDispatcher.closing {
					udpDispatcher.ReleaseUDPReadRoutine(idx)
					return
				}
				if er.Timeout() {
					udpDispatcher.remoteUDPConn[idx].SetReadDeadline(time.Now().Add(1 * time.Second))
					continue
				}
			}
		}
		// fast extract id from message
		udpDispatcher.UDPReadBytesChan[idx] <- b
		b = udpDispatcher.bytesPool.Get().([]byte)
		// return
	}
}

func (udpDispatcher *UDPDispatcher) ServeUDPFromOut(idx int64) {
	defer udpDispatcher.ReleaseUDPRoutine(idx)
	var tid uint16 = 0
	for !udpDispatcher.closing {
		udpDispatcher.serveUDPFromOut(idx, tid)
		tid++
	}
}

func (udpDispatcher *UDPDispatcher) serveUDPFromOut(idx int64, tid uint16) {
	buf := udpDispatcher.bytesPool.Get().([]byte)
	defer udpDispatcher.bytesPool.Put(buf)
	var tosendMessageBytes, remoteMessageBytes []byte
	udpDispatcher.connUDP.SetReadDeadline(time.Now().Add(1 * time.Second))
	for {

		for len(udpDispatcher.UDPReadBytesChan[idx]) > 0 {
			udpDispatcher.bytesPool.Put(<-udpDispatcher.UDPReadBytesChan[idx])
			fmt.Println("QAQ")
		}
		if udpDispatcher.closing {
			return
		}
		_, servingAddr, err := udpDispatcher.connUDP.ReadFromUDP(buf)
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
		_, err = message.Read(buf)
		if err != nil {
			udpDispatcher.logger.Errorf("failed when decoding udp msg, error: %v", err)
			return
		}

		udpDispatcher.logger.Infof("new message incoming: id, address: %v, %v", message.Header.ID, servingAddr)
		reply := msg.NewDNSMessageReply(message.Header.ID, message.Header.Flags, message.Question)
		message.Print()
		if message.Question[0].Type == qtype.A || message.Question[0].Type == qtype.AAAA {
			var (
				ipaddr net.IP
				ok     bool
			)
			if message.Question[0].Type == qtype.A {
				ipaddr, ok = hosts.HostsIPv4[string(message.Question[0].Name)]
				ipaddr = ipaddr.To4()
			} else {
				ipaddr, ok = hosts.HostsIPv6[string(message.Question[0].Name)]
			}
			if ok == true {
				replyans := msg.InitReply(message.Question[0])
				replyans.RDData = []byte(ipaddr)
				replyans.RDLength = uint16(len(replyans.RDData.([]byte)))
				//TODO verify
				reply.InsertAnswer(*replyans)
				tosendMessageBytes, err = reply.CompressToBytes()
				if err != nil {
					udpDispatcher.logger.Errorf("get redis cache error: %v", err)
					return
				}
				_, err = udpDispatcher.connUDP.WriteToUDP(tosendMessageBytes, servingAddr)
				if err != nil {
					udpDispatcher.logger.Errorf("write to client error: %v", err)
					return
				}
				udpDispatcher.logger.Infof("using hosts reply to address: %v, %v", message.Header.ID, servingAddr)
				return
			}

		}

		conn := mredis.RedisCacheClient.Pool.Get()
		defer conn.Close()
		if mredis.FindCache(reply, conn) {
			// reply.Print()
			tosendMessageBytes, err = reply.CompressToBytes()
			if err != nil {
				udpDispatcher.logger.Errorf("get redis cache error: %v", err)
				return
			}
			_, err = udpDispatcher.connUDP.WriteToUDP(tosendMessageBytes, servingAddr)
			if err != nil {
				udpDispatcher.logger.Errorf("write to client error: %v", err)
				return
			}

			udpDispatcher.logger.Infof("using redis cache reply to address: %v, %v", message.Header.ID, servingAddr)
			return
		} else {
			fid := message.Header.ID
			message.Header.ID = tid
			tosendMessageBytes, err = message.CompressToBytes()
			// b[0] = byte(idx >> 8)
			// b[1] = byte(idx & 0xff)
			if err != nil {
				udpDispatcher.logger.Errorf("convert error: %v", err)
				return
			}

			if _, err := udpDispatcher.remoteUDPConn[idx].Write(tosendMessageBytes); err != nil {
				udpDispatcher.logger.Errorf("write error: %v", err)
				return
			}

			var good = false
			for !good {
				select {
				case remoteMessageBytes = <-udpDispatcher.UDPReadBytesChan[idx]:
					if len(remoteMessageBytes) >= 12 &&
						((uint16(remoteMessageBytes[0])<<8)|uint16(remoteMessageBytes[1])) == tid {
						good = true
					}
				case <-time.After(time.Second * 1):
					//todo: reply...
					return
				}
			}
			fmt.Println("!!!")
			_, err = message.Read(remoteMessageBytes)
			udpDispatcher.bytesPool.Put(remoteMessageBytes)
			if err != nil {
				udpDispatcher.logger.Errorf("read error: %v", err)
				return
			}
			fmt.Println("!!!")
			// if idx != message.Header.ID {
			// 	udpDispatcher.logger.Errorf("not matching..., serving %v", servingAddr)
			// }
			// message.Print()
			message.Header.ID = fid
			message.Print()
			mredis.MessageToRedis(message, conn)

			tosendMessageBytes, err = message.CompressToBytes()
			defer udpDispatcher.bytesPool.Put(tosendMessageBytes)
			// b[0] = byte(fid >> 8)
			// b[1] = byte(fid & 0xff)
			if err != nil {
				udpDispatcher.logger.Errorf("convert error: %v", err)
				return
			}
			_, err = udpDispatcher.connUDP.WriteToUDP(tosendMessageBytes, servingAddr)
			if err != nil {
				udpDispatcher.logger.Errorf("write to client error: %v", err)
				return
			}

			udpDispatcher.logger.Infof("reply to address: %v, %v", message.Header.ID, servingAddr)
			return
		}
	}
}
