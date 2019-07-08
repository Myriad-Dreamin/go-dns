package dnssrv

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	// "github.com/garyburd/redigo/redis"
	mdnet "github.com/Myriad-Dreamin/go-dns/net"
	log "github.com/sirupsen/logrus"
)

const (
	UDPRange      = uint16(1000)
	UDPBufferSize = 520
	TCPRange      = uint16(5)
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
	SetUpUDP      chan bool
	QuitUDP       chan bool

	// stateful tcp managed by tcp dispatcher
	// TCPDispatcher
	TCPDispatcher *TCPDispatcher
	QuitTCP       chan bool
	SetUpTCP      chan bool

	quit chan bool
}

func (srv *Server) SetLogger(mLogger *log.Logger) {
	srv.logger = mLogger.WithFields(log.Fields{
		"prog": "server",
	})
}

type handler struct {
	logger *log.Entry
	funcs  []func()
}

func (h *handler) register(atexit func()) {
	h.funcs = append(h.funcs, atexit)
}

func (h *handler) atExit() {
	osQuitSignalChan := make(chan os.Signal)
	signal.Notify(osQuitSignalChan, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT,
		syscall.SIGKILL, syscall.SIGILL, syscall.SIGTERM,
	)
	for {
		select {
		case osc := <-osQuitSignalChan:
			h.logger.Infoln("handlering", osc)
			for _, f := range h.funcs {
				f()
			}
			return
		}
	}
}

func (srv *Server) ListenAndServe(host string) (err error) {

	if uint32(UDPRange)+uint32(TCPRange) > uint32(65536) {
		err = errors.New("limit size of link out of index")
		srv.logger.Errorln(err)
		return
	}

	srv.QuitTCP = make(chan bool, 1)
	srv.QuitUDP = make(chan bool, 1)
	srv.SetUpTCP = make(chan bool, 1)
	srv.SetUpUDP = make(chan bool, 1)
	srv.quit = make(chan bool, 1)
	go func() {

		err = srv.setupUDPDispatcher()
		if err != nil {
			srv.SetUpUDP <- false
			return
		}

		err = srv.PrepareUDPDispatcher(host)
		if err != nil {
			srv.SetUpUDP <- false
			return
		}

		srv.logger.Infof("all is ready for start udp server at %v", host)
		srv.SetUpUDP <- true
		srv.UDPDispatcher.Start(&srv.QuitUDP)
	}()

	go func() {
		err = srv.setupTCPDispatcher()
		if err != nil {
			srv.SetUpTCP <- false
			return
		}

		err = srv.prepareTCPDispatcher(host)
		if err != nil {
			srv.SetUpTCP <- false
			return
		}

		srv.logger.Infof("all is ready for start tcp server at %v", host)
		srv.SetUpTCP <- true
		go srv.TCPDispatcher.Start(&srv.QuitTCP)
	}()
	var mh = handler{srv.logger, nil}
	go mh.atExit()

	wait := 2
	for wait > 0 {
		select {
		case qwq := <-srv.SetUpTCP:
			if qwq {
				mh.register(srv.TCPDispatcher.AtExit)
			}
			wait--
		case qwq := <-srv.SetUpUDP:
			if qwq {
				mh.register(srv.UDPDispatcher.AtExit)
			}
			wait--
		}
	}
	mh.register(func() { srv.quit <- true })
	// close
	select {
	case <-srv.quit:
		return
	}
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
