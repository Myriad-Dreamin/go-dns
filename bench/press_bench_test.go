package bench

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Myriad-Dreamin/go-dns/msg"
)

type Client struct {
	WG       sync.WaitGroup
	srvMutex sync.Mutex

	conn       *net.UDPConn
	remoteConn net.Conn
	connected  bool
}

func (srv *Client) tryConnectToRemoteDNSServer(host string) (err error) {
	srv.remoteConn, err = net.Dial("udp", host)

	if err != nil {
		// fmt.Printf("error occurred when dial remote dns server: %v\n", err)
		return
	}
	return
}

func (srv *Client) tryDisonnectFromRemoteDNSServer() error {
	if srv.connected {
		srv.srvMutex.Lock()
		defer srv.srvMutex.Unlock()
		if srv.connected {
			srv.connected = false
			// fmt.Printf("disconnected from remote DNS server")
			return srv.remoteConn.Close()
		}
	}
	return nil
}

func (srv *Client) MTestLookUpA(tid uint16, req string) (ret string, err error) {
	//fmt.Println("+")
	defer func() {
		//fmt.Println("-")
		srv.WG.Done()
	}()
	requestNames := [][]byte{[]byte(req)}
	requsetTypes := []uint16{1}

	request := msg.Quest(
		requestNames,
		requsetTypes,
	)

	for len(request) != 0 {
		n, s := msg.NewDNSMessageRecursivelyQuery(tid, request)
		request = request[n:]

		// fmt.Println(n, s)
		b, err := s.ToBytes()
		if err != nil {
			//fmt.Printf("convert request message error: %v", err)
			return "", err
		}

		if _, err := srv.remoteConn.Write(b); err != nil {
			//fmt.Printf("write error: %v", err)
			return "", err
		}

		b = make([]byte, 1024)
		srv.remoteConn.SetDeadline(time.Now().Add(time.Millisecond * 5000))
		n, err = srv.remoteConn.Read(b)
		if err != nil {
			//fmt.Printf("read error: %v", err)
			return "", err
		}

		var rmsg = new(msg.DNSMessage)
		n, err = rmsg.Read(b)
		if err != nil {
			//fmt.Printf("convert read message error: %v", err)
			return "", err
		}
		// fmt.Println(n, err)
		// rmsg.Print()
	}
	return "", nil
}

func BenchmarkTestA(b *testing.B) {
	var c Client
	var host = "127.0.0.1"

	if err := c.tryConnectToRemoteDNSServer(host + ":53"); err != nil {
		b.Error(err)
		return
	}

	// fmt.Printf("udp socket set up successfully")
	c.connected = true
	defer func() {
		if err := c.tryDisonnectFromRemoteDNSServer(); err != nil {
			b.Error(err)
			return
		}
	}()
	b.ResetTimer()
	ff := uint16(1)
	b.N = 4000
	var bx = make(chan bool, 4000)
	for i := 0; i < b.N; i++ {
		c.WG.Add(1)
		go func() {
			ff++
			_, err := c.MTestLookUpA(ff, "www.baidu.com")
			if err != nil {
				bx <- false
			} else {
				bx <- true
			}
		}()
	}
	var erro, suco = 0, 0
	for i := 0; i < b.N; i++ {
		if <-bx {
			suco++
		} else {
			erro++
		}
	}
	fmt.Println(erro, suco)
	c.WG.Wait()
}
