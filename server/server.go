package dnssrv

import (
	"errors"
	"fmt"
	"net"
	"sync"

	msg "github.com/Myriad-Dreamin/go-dns/msg"
	mredis "github.com/Myriad-Dreamin/go-dns/redis"
	// "github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
)

const (
	UDPRange      = uint16(200)
	UDPBufferSize = 520
	TCPRange      = uint16(50)
	TCPBUfferSize = 65000
	// serverAddr    = "0.0.0.0:53"
	serverAddr = "127.0.0.1:53"
)

type Server struct {
	srvMutex sync.Mutex
	logger   *log.Entry

	conn       *net.UDPConn
	remoteConn net.Conn
	connected  bool

	UDPRoutineLimit chan uint16
	UDPBuffer       [UDPRange + 5][]byte

	UDPReadRoutineLimit chan uint16
	UDPReadBuffer       [UDPRange + 5][]byte
	UDPReadBytesChan    [UDPRange + 5]chan uint16

	TCPRoutineLimit chan uint16
	TCPBuffer       [TCPRange + 5][]byte
}

func (srv *Server) SetLogger(mLogger *log.Logger) {
	srv.logger = mLogger.WithFields(log.Fields{
		"prog": "server",
	})
}

func ParseUDPDNSIP6(host string) string {
	if _, err := net.ResolveUDPAddr("udp6", host); err == nil {
		return host
	} else {
		return host
	}
}

func ResolveUDPDNSIP(host string) (string, string) {
	if _, err := net.ResolveUDPAddr("udp4", host); err == nil {
		return "udp", host
	} else if _, err := net.ResolveIPAddr("ip4", host); err == nil {
		return "udp", host + ":53"
	} else if ip := net.ParseIP(host); ip != nil {
		return "udp6", "[" + ip.String() + "]:53"
	} else {
		return "udp6", ParseUDPDNSIP6(host)
	}
}

func (srv *Server) tryConnectToRemoteDNSServer(host string) (err error) {
	network, host := ResolveUDPDNSIP(host)
	srv.remoteConn, err = net.Dial(network, host)

	if err != nil {
		srv.logger.Errorf("error occurred when dial remote dns server: %v\n", err)
		return
	}

	return
}

func (srv *Server) tryDisonnectFromRemoteDNSServer() error {
	if srv.connected {
		srv.srvMutex.Lock()
		defer srv.srvMutex.Unlock()
		if srv.connected {
			srv.connected = false
			srv.logger.Infof("disconnected from remote DNS server")
			return srv.remoteConn.Close()
		}
	}
	return nil
}

func (srv *Server) ListenAndServe(host string) (err error) {

	if uint32(UDPRange)+uint32(TCPRange) > uint32(65536) {
		err = errors.New("limit size of link out of index")
		srv.logger.Errorln(err)
		return
	}

	if err = srv.tryConnectToRemoteDNSServer(host); err != nil {
		return
	}

	srv.logger.Infof("udp socket set up successfully")
	srv.connected = true
	defer func() {
		err = srv.tryDisonnectFromRemoteDNSServer()
	}()

	var udpAddr *net.UDPAddr
	udpAddr, err = net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return
	}
	srv.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}
	defer srv.conn.Close()

	if err != nil {
		srv.logger.Errorf("setup local server error: %v", err)
		return err
	}

	srv.UDPRoutineLimit = make(chan uint16, UDPRange)
	srv.TCPRoutineLimit = make(chan uint16, TCPRange)
	srv.UDPReadRoutineLimit = make(chan uint16, UDPRange)
	for idx := uint16(0); idx < UDPRange; idx++ {
		srv.UDPRoutineLimit <- idx
		srv.UDPReadRoutineLimit <- idx
		srv.UDPBuffer[idx] = make([]byte, UDPBufferSize)
		srv.UDPReadBuffer[idx] = make([]byte, UDPBufferSize)
		srv.UDPReadBytesChan[idx] = make(chan uint16, 1)
	}
	for idx := UDPRange; idx < TCPRange; idx++ {
		srv.TCPRoutineLimit <- idx
		srv.TCPBuffer[idx-UDPRange] = make([]byte, TCPBUfferSize)
	}
	for {
		select {
		case idx := <-srv.UDPRoutineLimit:
			go srv.ServeUDPFromOut(idx, srv.UDPBuffer[idx])
		case idx := <-srv.TCPRoutineLimit:
			go srv.ServeTCPFromOut(idx, srv.TCPBuffer[idx-UDPRange])
		case idx := <-srv.UDPReadRoutineLimit:
			go srv.ServeUDPReadFromOut(idx, srv.UDPReadBuffer[idx])
		}
	}
}

/*
0003 8180 0001000200010001 0763617074636861056774696d6703636f6d00001c0001
490a 8180 0001000200010001 0763617074636861056774696d6703636f6d00001c0001
c00c
                                       000500010000012c002007636170746368610567
‘0763617074636861056774696d6703636f6d00
'0763617074636861056774696d6703636f6d00000500010000012c002007636170746368610567


74696d6703636f6d05636c6f7564027463027171c01ac02f0005000100000201000b03703231047463646ec04ac05f00060001000000c5002a076e732d63646e31c04a097765626d6173746572c04a4fd6eede0000012c00000258000151800000012c000029100000000000001c000a00182bf4de06c37af48ac70b0f605d1efbf7ab21838eea49ec19

74696d6703636f6d05636c6f7564027463027171c01a0763617074636861056774696d6703636f
6d05636c6f756402746302717103636f6d000005000100000201000b03703231047463646ec04a
047463646e02717103636f6d0000060001000000c5002a076e732d63646e31c04a097765626d61
73746572c04a4fd6eede0000012c00000258000151800000012c000029100000000000001c000a
00182bf4de06c37af48ac70b0f605d1efbf7ab21838eea49ec19
*/

func (srv *Server) LookUpA(host, req string) (ret string, err error) {

	if err = srv.tryConnectToRemoteDNSServer(host); err != nil {
		return
	}

	srv.logger.Infof("udp socket set up successfully")
	srv.connected = true
	defer func() {
		err = srv.tryDisonnectFromRemoteDNSServer()
	}()

	requestNames := [][]byte{[]byte(req)}
	requsetTypes := []uint16{1}

	request := msg.Quest(
		requestNames,
		requsetTypes,
	)

	for len(request) != 0 {
		n, s := msg.NewDNSMessageRecursivelyQuery(1, request)
		request = request[n:]

		fmt.Println(n, s)
		b, err := s.ToBytes()
		if err != nil {
			srv.logger.Errorf("convert request message error: %v", err)
			return "", err
		}

		if _, err := srv.remoteConn.Write(b); err != nil {
			srv.logger.Errorf("write error: %v", err)
			return "", err
		}

		b = make([]byte, 1024)
		n, err = srv.remoteConn.Read(b)
		if err != nil {
			srv.logger.Errorf("read error: %v", err)
			return "", err
		}

		var rmsg = new(msg.DNSMessage)
		n, err = rmsg.Read(b)
		if err != nil {
			srv.logger.Errorf("convert read message error: %v", err)
			return "", err
		}
		fmt.Println(n, err)
		rmsg.Print()
	}
	return "", nil
}

func (srv *Server) ReleaseUDPRoutine(tid uint16) {
	srv.UDPRoutineLimit <- tid
}

func (srv *Server) ReleaseUDPReadRoutine(tid uint16) {
	srv.UDPReadRoutineLimit <- tid
}

func (srv *Server) ServeUDPReadFromOut(tid uint16, b []byte) {
	_, err := srv.remoteConn.Read(b)
	// srv.srvMutex.Lock()
	// defer srv.srvMutex.Unlock()
	if err != nil {
		srv.logger.Errorf("read error: %v", err)
		return
	}

	// fast extract id from message
	srv.UDPReadBytesChan[(uint16(b[0])<<8)+uint16(b[1])] <- tid
}

func (srv *Server) ServeUDPFromOut(tid uint16, b []byte) {
	defer srv.ReleaseUDPRoutine(tid)
	_, servingAddr, err := srv.conn.ReadFromUDP(b)

	var message msg.DNSMessage
	_, err = message.Read(b)

	// message.Print()

	if err != nil {
		srv.logger.Errorf("failed read udp msg, error: " + err.Error())
		return
	}
	srv.logger.Infof("new message incoming: id, address: %v, %v", message.Header.ID, servingAddr)

	conn := mredis.RedisCacheClient.Pool.Get()
	defer conn.Close()

	reply := msg.NewDNSMessageReply(message.Header.ID, message.Header.Flags, message.Question)
	if mredis.FindCache(reply, conn) {
		// reply.Print()
		b, err := reply.CompressToBytes()
		if err != nil {
			srv.logger.Errorf("get redis cache error: %v", err)
			return
		}
		_, err = srv.conn.WriteToUDP(b, servingAddr)
		if err != nil {
			srv.logger.Errorf("write to client error: %v", err)
			return
		}

		srv.logger.Infof("using redis cache reply to address: %v, %v", message.Header.ID, servingAddr)
	} else {
		fid := message.Header.ID
		message.Header.ID = tid
		b, err = message.CompressToBytes()
		// b[0] = byte(tid >> 8)
		// b[1] = byte(tid & 0xff)
		if err != nil {
			srv.logger.Errorf("convert error: %v", err)
			return
		}

		if _, err := srv.remoteConn.Write(b); err != nil {
			srv.logger.Errorf("write error: %v", err)
			return
		}
		rid := <-srv.UDPReadBytesChan[tid]
		defer srv.ReleaseUDPReadRoutine(rid)
		b = srv.UDPReadBuffer[rid]

		_, err = message.Read(b)
		// if tid != message.Header.ID {
		// 	srv.logger.Errorf("not matching..., serving %v", servingAddr)
		// }
		// message.Print()
		message.Header.ID = fid

		mredis.MessageToRedis(message, conn)

		b, err = message.CompressToBytes()

		// b[0] = byte(fid >> 8)
		// b[1] = byte(fid & 0xff)
		if err != nil {
			srv.logger.Errorf("convert error: %v", err)
			return
		}
		_, err = srv.conn.WriteToUDP(b, servingAddr)
		if err != nil {
			srv.logger.Errorf("write to client error: %v", err)
			return
		}

		srv.logger.Infof("reply to address: %v, %v", message.Header.ID, servingAddr)
	}
}

func (srv *Server) ReleaseTCPRoutine(tid uint16) {
	srv.TCPRoutineLimit <- tid
}

func (srv *Server) ServeTCPFromOut(tid uint16, b []byte) {
	defer srv.ReleaseUDPRoutine(tid)
	_, servingAddr, err := srv.conn.ReadFromUDP(b)
	if err != nil {
		srv.logger.Errorf("failed read udp msg, error: " + err.Error())
		return
	}
	srv.logger.Infof("new message incoming: address: %v", servingAddr)

	if _, err := srv.remoteConn.Write(b); err != nil {
		srv.logger.Errorf("write error: %v", err)
		return
	}

	_, err = srv.remoteConn.Read(b)
	if err != nil {
		srv.logger.Errorf("read error: %v", err)
		return
	}

	_, err = srv.conn.WriteToUDP(b, servingAddr)
	if err != nil {
		srv.logger.Errorf("write to client error: %v", err)
		return
	}

	srv.logger.Infof("reply to address: %v", servingAddr)
}
