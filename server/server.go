package dnssrv

import (
	"fmt"
	"net"
	"sync"

	msg "github.com/Myriad-Dreamin/go-dns/msg"
	log "github.com/sirupsen/logrus"
)

const MaxRoutineSize = 10

type Server struct {
	srvMutex sync.Mutex
	logger   *log.Entry

	conn      net.Conn
	connected bool

	quitChain chan int
}

func (srv *Server) SetLogger(mLogger *log.Logger) {
	srv.logger = mLogger.WithFields(log.Fields{
		"prog": "server",
	})
}

func (srv *Server) tryConnectToRemoteDNSServer(host string) (err error) {
	srv.conn, err = net.Dial("udp", host)

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
			return srv.conn.Close()
		}
	}
	return nil
}

func (srv *Server) ListenAndServe(port, host string) (err error) {
	if err = srv.tryConnectToRemoteDNSServer(host + ":53"); err != nil {
		return
	}

	srv.logger.Infof("udp socket set up successfully")
	srv.connected = true
	defer func() {
		err = srv.tryDisonnectFromRemoteDNSServer()
	}()

	requestNames := [][]byte{
		[]byte("www.baidu.com"),
		[]byte("www.163.com"),
	}
	requsetTypes := []uint16{1, 1}

	request := msg.Quest(
		requestNames,
		requsetTypes,
	)

	for len(request) != 0 {
		n, s := msg.NewDNSMessageContextRecursivelyQuery(1, request)
		fmt.Println(n, s)
		fmt.Println(s.ToBytes())
		request = request[n:]
	}

	return
}

func (srv *Server) CreateServeRoutine() {
	// go func() {
	// 	for {
	// 		select {
	// 		case job := <-queue:
	// 			job.Do(request)
	// 		case <-quit:
	// 			return
	// 		}
	//
	// 	}
	// }()
}
