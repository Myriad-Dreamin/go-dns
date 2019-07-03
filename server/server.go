package dnssrv

import (
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	logger *log.Entry

	srvMutex  sync.Mutex
	conn      net.Conn
	connected bool
}

func (srv *Server) SetLogger(mLogger *log.Logger) {
	srv.logger = mLogger.WithFields(logrus.Fields{
		"prog": "server",
	})
}

func (srv *Server) tryConnectToRemoteDNSServer(host string) (err error) {
	srv.conn, err = net.Dial("udp", host)

	if err != nil {
		srv.logger.Printf("error occurred when dial remote dns server: %v\n", err)
		return
	}
	return
}

func (srv *Server) tryDisonnectFromRemoteDNSServer() error {
	if srv.connected {
		srv.srvMutex.Lock()
		defer srv.srvMutex.Unlock()
		if srv.connected {
			return srv.conn.Close()
		}
	}
	return nil
}

func (srv *Server) ListenAndServe(port, host string) (err error) {
	if err := srv.tryConnectToRemoteDNSServer(host + ":53"); err != nil {
		return err
	}
	srv.connected = true
	defer func() {
		err = srv.tryDisonnectFromRemoteDNSServer()
	}()

	srv.logger.Infof("tcp socket set up successfully")

	return nil
}
