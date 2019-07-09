package dnssrv

import (
	"bytes"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type BytesPool struct {
	*sync.Pool
}

func NewBytesPool(maxSize int64) *BytesPool {
	return &BytesPool{
		Pool: &sync.Pool{
			New: func() interface{} { return make([]byte, maxSize) },
		},
	}
}

type BufferPool struct {
	*sync.Pool
}

func NewBufferPool() *BufferPool {
	return &BufferPool{
		Pool: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}
}

type sharedSpace struct {
	logger     *log.Entry
	bytesPool  *BytesPool
	bufferPool *BufferPool
	quit       chan bool
	dispatcher *TCPDispatcher
}

func (ss *sharedSpace) SetDispatcher(td *TCPDispatcher) {
	ss.dispatcher = td
}

type TCPDispatcher struct {
	*sharedSpace
	messageChan            chan *bytes.Buffer
	connTCP                *net.TCPListener
	tcpRemoteServerRoutine []*TCPRemoteServerRoutine
	tcpUserRoutine         []*TCPUserRoutine
	tidL                   uint16
	tidR                   uint16
	tcpRange               uint16
}

func NewTCPDispatcher(
	logger *log.Entry,
	maxSize int64,
	idRangeL, idRangeR, tcpRange uint16,
) (td *TCPDispatcher) {
	if tcpRange != idRangeR-idRangeL {
		return
	}
	td = &TCPDispatcher{
		sharedSpace: &sharedSpace{
			logger:     logger,
			bytesPool:  NewBytesPool(maxSize),
			bufferPool: NewBufferPool(),
			quit:       make(chan bool, tcpRange*2),
		},
		tidL:                   idRangeL,
		tidR:                   idRangeR,
		tcpRange:               tcpRange,
		messageChan:            make(chan *bytes.Buffer),
		tcpUserRoutine:         make([]*TCPUserRoutine, tcpRange, tcpRange),
		tcpRemoteServerRoutine: make([]*TCPRemoteServerRoutine, tcpRange, tcpRange),
	}
	td.SetDispatcher(td)
	return
}

func (tcpDispatcher *TCPDispatcher) listenTCP() (err error) {
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		tcpDispatcher.logger.Errorf("resolve local tcp server address error: %v", err)
		return
	}
	tcpDispatcher.connTCP, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		tcpDispatcher.logger.Errorf("setup local tcp server error: %v", err)
		return
	}
	return
}

func (d *TCPDispatcher) Prepare(network string, host *net.TCPAddr) error {
	for i := d.tidL; i < d.tidR; i++ {
		d.tcpRemoteServerRoutine[i-d.tidL] = NewTCPRemoteServerRoutine(
			d.sharedSpace,
			network,
			host,
		)
		d.tcpUserRoutine[i-d.tidL] = NewTCPUserRoutine(d.sharedSpace, i)
	}
	return d.listenTCP()
}

func (d *TCPDispatcher) Start(qc *chan bool) (err error) {
	osQuitSignalChan := make(chan os.Signal)
	signal.Notify(osQuitSignalChan, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT,
		syscall.SIGKILL, syscall.SIGILL, syscall.SIGTERM,
	)
	for i := uint16(0); i < d.tcpRange; i++ {
		go d.tcpRemoteServerRoutine[i].Run()
		go d.tcpUserRoutine[i].Run()
	}
	for {
		select {
		// case idx := <-srv.TCPRoutineLimit:
		// 	go srv.ServeTCPFromOut(idx, srv.TCPBuffer[idx-UDPRange])
		// case idx := <-srv.TCPWriteRoutineLimit:
		// 	go srv.ServeTCPWriteToOut(idx)
		// case idx := <-srv.TCPReadRoutineLimit:
		// 	go srv.ServeTCPReadFromOut(idx, srv.TCPReadBuffer[idx-UDPRange])
		}
	}
}

func (d *TCPDispatcher) AtExit() {
	if err := d.Stop(); err != nil {
		d.logger.Errorf("error occurred when stopping tcp dispatcher, error: %v", err)
	}
	d.logger.Infof("tcp server stop successfully")
}

func (d *TCPDispatcher) Stop() error {
	var reqs = 0
	for i := uint16(0); i < d.tcpRange; i++ {
		if d.tcpRemoteServerRoutine[i].RequestQuit() {
			reqs++
		}
		if d.tcpUserRoutine[i].RequestQuit() {
			reqs++
		}
	}
	for i := reqs; i > 0; i-- {
		<-d.quit
	}
	return d.connTCP.Close()
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
