package dnssrv

import (
	"fmt"
	"net"
	"sync"

	msg "github.com/Myriad-Dreamin/go-dns/msg"
	log "github.com/sirupsen/logrus"
)

const (
	MaxRoutineSize = 10
	serverAddr     = "0.0.0.0:53"
)

type Server struct {
	srvMutex sync.Mutex
	logger   *log.Entry

	conn       *net.UDPConn
	remoteConn net.Conn
	connected  bool

	RoutineLimit chan bool
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

func (srv *Server) AccquireRoutine() {
	srv.RoutineLimit <- true
}

func (srv *Server) ReleaseRoutine() {
	<-srv.RoutineLimit
}

func (srv *Server) ListenAndServe(host string) (err error) {
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

	srv.RoutineLimit = make(chan bool, MaxRoutineSize)
	for {
		srv.AccquireRoutine()
		go srv.ServeFromOut()
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

func (srv *Server) ServeFromOut() {
	defer srv.ReleaseRoutine()
	b := make([]byte, 1024)
	_, servingAddr, err := srv.conn.ReadFromUDP(b)
	if err != nil {
		srv.logger.Errorf("failed read udp msg, error: " + err.Error())
		return
	}

	if _, err := srv.remoteConn.Write(b); err != nil {
		srv.logger.Errorf("write error: %v", err)
		return
	}

	b = make([]byte, 1024)
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
}
