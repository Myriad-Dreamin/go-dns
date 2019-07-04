package dnssrv

import (
	"errors"
	"fmt"
	"net"
	"sync"

	msg "github.com/Myriad-Dreamin/go-dns/msg"
	log "github.com/sirupsen/logrus"
)

const (
	UDPRange      = uint16(6500)
	UDPBufferSize = 520
	TCPRange      = uint16(50)
	TCPBUfferSize = 50000
	serverAddr    = "0.0.0.0:53"
)

type Server struct {
	srvMutex sync.Mutex
	logger   *log.Entry

	conn       *net.UDPConn
	remoteConn net.Conn
	connected  bool

	UDPRoutineLimit chan uint16
	UDPBuffer       [UDPRange][]byte
	TCPRoutineLimit chan uint16
	TCPBuffer       [TCPRange][]byte
}

func (srv *Server) SetLogger(mLogger *log.Logger) {
	srv.logger = mLogger.WithFields(log.Fields{
		"prog": "server",
	})
}

func (srv *Server) tryConnectToRemoteDNSServer(host string) (err error) {
	srv.remoteConn, err = net.Dial("udp", host)

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

	if err = srv.tryConnectToRemoteDNSServer(host + ":53"); err != nil {
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
	for idx := uint16(0); idx < UDPRange; idx++ {
		srv.UDPRoutineLimit <- idx
		srv.UDPBuffer[idx] = make([]byte, UDPBufferSize)
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
		}
	}
}

func (srv *Server) LookUpA(host, req string) (ret string, err error) {
	if err = srv.tryConnectToRemoteDNSServer(host + ":53"); err != nil {
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
		n, s := msg.NewDNSMessageContextRecursivelyQuery(1, request)
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

func (srv *Server) ServeUDPFromOut(tid uint16, b []byte) {
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
