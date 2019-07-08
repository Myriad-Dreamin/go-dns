package dnssrv

import (
	"bytes"
	"net"
	"sync"

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

func (d *TCPDispatcher) Start(qc chan bool) (err error) {
	qc <- true
	return
}

func (d *TCPDispatcher) Stop() error {
	for i := uint16(0); i < d.tcpRange; i++ {
		d.tcpRemoteServerRoutine[i].QuitRequest <- true
		d.tcpUserRoutine[i].QuitRequest <- true
	}
	for i := int32(d.tcpRange) * 2; i > 0; i-- {
		<-d.quit
	}
	return d.connTCP.Close()
}
