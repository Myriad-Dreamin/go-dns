package dnssrv

import (
	"errors"
	"net"
	"sync"
	"time"

	// "github.com/garyburd/redigo/redis"
	mdnet "github.com/Myriad-Dreamin/go-dns/net"
	log "github.com/sirupsen/logrus"
)

const (
	UDPRange      = uint16(200)
	UDPBufferSize = 520
	TCPRange      = uint16(50)
	TCPBufferSize = 65000
	TCPTimeout    = 10 * time.Second
	// serverAddr    = "0.0.0.0:53"
	serverAddr = "127.0.0.1:53"
)

type Server struct {
	srvMutex sync.Mutex
	logger   *log.Entry

	// stateless udp managed by udp dispatcher
	// UDPDispatcher
	UDPDispatcher *UDPDispatcher

	// stateful tcp managed by tcp dispatcher
	// TCPDispatcher
	TCPDispatcher *TCPDispatcher

	Quit chan bool
}

func (srv *Server) SetLogger(mLogger *log.Logger) {
	srv.logger = mLogger.WithFields(log.Fields{
		"prog": "server",
	})
}

func (srv *Server) ListenAndServe(host string) (err error) {

	if uint32(UDPRange)+uint32(TCPRange) > uint32(65536) {
		err = errors.New("limit size of link out of index")
		srv.logger.Errorln(err)
		return
	}

	err = srv.setupUDPDispatcher()
	if err != nil {
		return
	}

	err = srv.setupTCPDispatcher()
	if err != nil {
		return
	}

	srv.Quit = make(chan bool, 2)

	err = srv.PrepareUDPDispatcher(host)
	if err != nil {
		return
	}

	err = srv.prepareTCPDispatcher(host)
	if err != nil {
		return
	}

	go srv.UDPDispatcher.Start(srv.Quit)
	go srv.TCPDispatcher.Start(srv.Quit)

	<-srv.Quit
	<-srv.Quit

	// close
	_ = srv.UDPDispatcher.Stop()
	_ = srv.TCPDispatcher.Stop()
	return
}

func (srv *Server) setupTCPDispatcher() error {
	srv.TCPDispatcher = NewTCPDispatcher(
		srv.logger,
		TCPBufferSize,
		UDPRange,
		UDPRange+TCPRange,
		TCPRange,
	)
	if srv.TCPDispatcher == nil {
		srv.logger.Errorf("set up tcp dispatcher failed")
		return errors.New("set up tcp dispatcher failed")
	}
	return nil
}

func (srv *Server) setupUDPDispatcher() error {
	srv.UDPDispatcher = NewUDPDispatcher(
		srv.logger,
		UDPBufferSize,
		0,
		UDPRange,
		UDPRange,
	)
	if srv.UDPDispatcher == nil {
		srv.logger.Errorf("set up udp dispatcher failed")
		return errors.New("set up udp dispatcher failed")
	}
	return nil
}

func (srv *Server) prepareTCPDispatcher(host string) (err error) {
	network, host := mdnet.ResolveDNSIP("tcp", host)
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr(network, host)
	if err != nil {
		srv.logger.Errorf(
			"error occurred when resolving remote dns ip: %v\n",
			err,
		)
		return err
	}
	return srv.TCPDispatcher.Prepare(network, addr)
}

func (srv *Server) PrepareUDPDispatcher(host string) (err error) {
	network, host := mdnet.ResolveDNSIP("udp", host)
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr(network, host)
	if err != nil {
		srv.logger.Errorf(
			"error occurred when resolving remote dns ip: %v\n",
			err,
		)
		return
	}
	return srv.UDPDispatcher.Prepare(network, addr)
}

// func (srv *Server) ReleaseTCPRoutine(tid uint16) {
// 	srv.TCPRoutineLimit <- tid
// }
//
// func (srv *Server) ReleaseTCPReadRoutine(tid uint16) {
// 	srv.TCPReadRoutineLimit <- tid
// }
//
// func (srv *Server) ReleaseTCPWriteRoutine(tid uint16) {
// 	srv.TCPWriteRoutineLimit <- tid
// }

// func (srv *Server) ServeTCPReadFromOut(tid uint16, b []byte) {
// 	_, err := srv.remoteTCPConn[tid-UDPRange].Read(b)
// 	if err != nil {
// 		srv.logger.Errorf("read error: %v", err)
// 		return
// 	}
//
// 	// fast extract id from message
// 	f := (uint16(b[0]) << 8) + uint16(b[1])
// 	if f < UDPRange {
// 		srv.UDPReadBytesChan[f] <- tid
// 	} else {
// 		srv.TCPReadBytesChan[f-UDPRange] <- tid
// 	}
// }

// func (srv *Server) ServeTCPWriteToOut(tid uint16) {
// 	select {
// 	case b := <-srv.TCPWriteBytesChan:
// 		_, err := srv.remoteTCPConn[tid-UDPRange].Write(b)
// 		if err != nil {
// 			srv.logger.Errorf("tcp write error: %v", err)
// 		}
// 	}
// }

// func (srv *Server) ServeTCPFromOut(tid uint16, b []byte) {
// 	defer srv.ReleaseTCPRoutine(tid)
//
// 	tcpConn, err := srv.connTCP.AcceptTCP()
// 	if err != nil {
// 		srv.logger.Errorf("failed when accepting tcp connection, error: %v", err)
// 		return
// 	}
// 	tcpConn.SetDeadline(time.Now().Add(TCPTimeout))
// 	defer tcpConn.Close()
//
// 	_, err = tcpConn.Read(b)
// 	if err != nil && err != io.EOF {
// 		srv.logger.Errorf("failed when reading tcp flow, error: %v", err)
// 		return
// 	}
//
// 	var message msg.DNSMessage
// 	_, err = message.Read(b)
// 	if err != nil {
// 		srv.logger.Errorf("failed when decoding tcp flow, error: %v", err)
// 		return
// 	}
//
// 	srv.logger.Infof("new message(tcp) incoming: id, address: %v, %v", message.Header.ID, tcpConn.RemoteAddr())
//
// 	conn := mredis.RedisCacheClient.Pool.Get()
// 	defer conn.Close()
//
// 	reply := msg.NewDNSMessageReply(message.Header.ID, message.Header.Flags, message.Question)
// 	if mredis.FindCache(reply, conn) {
// 		// reply.Print()
// 		b, err := reply.CompressToBytes()
// 		if err != nil {
// 			srv.logger.Errorf("get redis cache error: %v", err)
// 			return
// 		}
// 		_, err = tcpConn.Write(b)
// 		if err != nil {
// 			srv.logger.Errorf("write to client error: %v", err)
// 			return
// 		}
//
// 		srv.logger.Infof("using redis cache reply to address: %v, %v", message.Header.ID, tcpConn.RemoteAddr())
// 	} else {
// 		fid := message.Header.ID
// 		message.Header.ID = tid
// 		b, err = message.CompressToBytes()
// 		// b[0] = byte(tid >> 8)
// 		// b[1] = byte(tid & 0xff)
// 		if err != nil {
// 			srv.logger.Errorf("convert error: %v", err)
// 			return
// 		}
//
// 		srv.TCPWriteBytesChan <- b
// 		rid := <-srv.TCPReadBytesChan[tid-UDPRange]
// 		defer srv.ReleaseTCPReadRoutine(rid)
// 		b = srv.TCPReadBuffer[rid-UDPRange]
//
// 		_, err = message.Read(b)
// 		if err != nil {
// 			srv.logger.Errorf("read error: %v", err)
// 			return
// 		}
// 		// if tid != message.Header.ID {
// 		// 	srv.logger.Errorf("not matching..., serving %v", servingAddr)
// 		// }
// 		// message.Print()
// 		message.Header.ID = fid
//
// 		mredis.MessageToRedis(message, conn)
//
// 		b, err = message.CompressToBytes()
//
// 		// b[0] = byte(fid >> 8)
// 		// b[1] = byte(fid & 0xff)
// 		if err != nil {
// 			srv.logger.Errorf("convert error: %v", err)
// 			return
// 		}
// 		_, err = tcpConn.Write(b)
// 		if err != nil {
// 			srv.logger.Errorf("write to client error: %v", err)
// 			return
// 		}
//
// 		srv.logger.Infof("reply to address: %v, %v", message.Header.ID, tcpConn.RemoteAddr())
// 	}
// }
