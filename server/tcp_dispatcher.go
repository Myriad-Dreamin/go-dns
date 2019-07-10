package dnssrv

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type BytesPool struct {
	*sync.Pool
}

func NewBytesPool(maxSize int64) *BytesPool {
	return &BytesPool{
		Pool: &sync.Pool{
			New: func() interface{} {
				xx := make([]byte, maxSize)
				// fmt.Println("get bytes", len(xx), cap(xx))
				return xx
			},
		},
	}
}

type BufferPool struct {
	*sync.Pool
}

func NewBufferPool(bf *BytesPool) *BufferPool {
	return &BufferPool{
		Pool: &sync.Pool{
			New: func() interface{} {
				xx := bf.Get().([]byte)
				// fmt.Println("set bytes", len(xx), cap(xx))
				s := bytes.NewBuffer(xx)
				// fmt.Println("get buffer", s.Len(), s.Cap())
				s.Reset()
				return s
			},
		},
	}
}

type sharedSpace struct {
	logger     *log.Entry
	bytesPool  *BytesPool
	bufferPool *BufferPool
	quit       chan bool
	dispatcher *TCPDispatcher
	tcpTimeout time.Duration
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
	tcpTimeout time.Duration,
) (td *TCPDispatcher) {
	if tcpRange != idRangeR-idRangeL {
		return
	}
	var bp = NewBytesPool(maxSize)
	td = &TCPDispatcher{
		sharedSpace: &sharedSpace{
			logger:     logger,
			bytesPool:  bp,
			bufferPool: NewBufferPool(bp),
			quit:       make(chan bool, tcpRange*2),
			tcpTimeout: tcpTimeout,
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

func (tcpDispatcher *TCPDispatcher) listenTCP(serverAddr string) (err error) {
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
	tcpDispatcher.logger.Infof("setup local tcp server success, at: %v", tcpAddr)
	return
}

func (d *TCPDispatcher) Prepare(localserverAddr, network string, host *net.TCPAddr) (err error) {
	for i := d.tidL; i < d.tidR; i++ {
		d.tcpRemoteServerRoutine[i] = NewTCPRemoteServerRoutine(
			d.sharedSpace,
			network,
			host,
		)

		if d.tcpRemoteServerRoutine[i] == nil {
			err = fmt.Errorf("new tcp remote server routine failed at %v", i)
			d.logger.Errorln(err)
			return
		}

		// fmt.Println(i)
		d.tcpUserRoutine[i] = NewTCPUserRoutine(d.sharedSpace, i)
		if d.tcpRemoteServerRoutine[i] == nil {
			err = fmt.Errorf("new tcp user routine failed at %v", i)
			d.logger.Errorln(err)
			return
		}
	}
	return d.listenTCP(localserverAddr)
}

func MinI(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (d *TCPDispatcher) Start(qc *chan bool) (err error) {
	// osQuitSignalChan := make(chan os.Signal)
	// signal.Notify(osQuitSignalChan, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT,
	// 	syscall.SIGKILL, syscall.SIGILL, syscall.SIGTERM,
	// )
	for i := d.tidL; i < d.tidR; i++ {
		// fmt.Println(i)
		go d.tcpRemoteServerRoutine[i].Run()
		go d.tcpUserRoutine[i].Run()
	}
	for {
		select {
		case buf := <-d.messageChan:
			mymi := 1113181943
			for idx := uint16(0); idx < d.tcpRange; idx++ {
				mlen := len(d.tcpRemoteServerRoutine[idx].MessageChan)
				if mlen == 0 {
					d.tcpRemoteServerRoutine[idx].MessageChan <- buf
					mymi = -1
					break
				}
				mymi = MinI(mymi, mlen)
			}
			if mymi == -1 {
				continue
			}
			for idx := uint16(0); idx < d.tcpRange; idx++ {
				mlen := len(d.tcpRemoteServerRoutine[idx].MessageChan)
				if mlen <= mymi {
					d.tcpRemoteServerRoutine[idx].MessageChan <- buf
					break
				}
			}
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
	for i := d.tidL; i < d.tidR; i++ {
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
