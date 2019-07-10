package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Myriad-Dreamin/go-dns/msg"
	mdnet "github.com/Myriad-Dreamin/go-dns/net"
	log "github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli"
)

type ServerLookUpACmd struct {
	parentCmd *ServerCmd
	logger    *log.Entry

	seed    string
	outfile string
	datadir string
	wltname string
	show    bool
	a       string
	t       string
}

func (srv *ServerLookUpACmd) RequestRootLogger() *log.Logger {
	return srv.parentCmd.RequestRootLogger()
}

func (cmd *ServerLookUpACmd) Before(c *urcli.Context) (err error) {
	cmd.logger = cmd.parentCmd.logger
	cmd.a = c.Args().Get(0)
	return nil
}

func (cmd *ServerLookUpACmd) Action(c *urcli.Context) error {
	switch cmd.t {
	case "tcp":
		if x, err := cmd.TCPLookUpA(cmd.parentCmd.host, cmd.a); err != nil {
			fmt.Println(x)
			return err
		}
	case "udp":
		if x, err := cmd.UDPLookUpA(cmd.parentCmd.host, cmd.a); err != nil {
			fmt.Println(x)
			return err
		}
	default:
		return errors.New("unknown network type")
	}
	return nil
}

func NewServerLookUpACmd(srv *ServerCmd) urcli.Command {
	var srvLookUpA = &ServerLookUpACmd{parentCmd: srv}
	return urcli.Command{
		Name:        "lookupA",
		ShortName:   "lpa",
		Usage:       "server api",
		UsageText:   "lookup the name",
		Category:    "server",
		Action:      srvLookUpA.Action,
		Before:      srvLookUpA.Before,
		After:       nil,
		Subcommands: nil,
		Flags: []urcli.Flag{
			urcli.StringFlag{
				Name:        "stype",
				Value:       "udp",
				Usage:       "remote dns type",
				Destination: &srvLookUpA.t,
			},
		},
	}
}

func (cmd *ServerLookUpACmd) UDPLookUpA(host, req string) (ret string, err error) {
	var conn *net.UDPConn

	network, host := mdnet.ResolveDNSIP("udp", host)
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr(network, host)

	if err = func() (err error) {
		if err != nil {
			cmd.logger.Errorf("error occurred when resolving remote dns server ip: %v\n", err)
		}
		conn, err = net.DialUDP(network, nil, addr)
		if err != nil {
			cmd.logger.Errorf("error occurred when dial remote udp DNS Server: %v\n", err)
		}
		return
	}(); err != nil {
		return
	}

	cmd.logger.Infof("udp socket set up successfully")
	defer func() {
		cmd.logger.Infof("disconnected from remote udp DNS Server")
		err = conn.Close()
	}()

	requestNames := [][]byte{[]byte(req)}
	requsetTypes := []uint16{1}

	request := msg.Quest(
		requestNames,
		requsetTypes,
	)

	for len(request) != 0 {
		n, s := msg.NewDNSMessageRecursivelyQuery(1, request)
		request = request[n:]

		fmt.Println(n, s)
		b, err := s.ToBytes()
		b = append(b, []byte{1, 2, 3, 0, 6}...)
		if err != nil {
			cmd.logger.Errorf("convert request message error: %v", err)
			return "", err
		}

		if _, err := conn.Write(b); err != nil {
			cmd.logger.Errorf("write error: %v", err)
			return "", err
		}

		b = make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, err = conn.Read(b)
		if err != nil {
			cmd.logger.Errorf("read error: %v", err)
			return "", err
		}

		var rmsg = new(msg.DNSMessage)
		n, err = rmsg.Read(b)
		if err != nil {
			cmd.logger.Errorf("convert read message error: %v", err)
			return "", err
		}
		fmt.Println(n, err)
		rmsg.Print()
	}
	return "", nil
}

func (cmd *ServerLookUpACmd) TCPLookUpA(host, req string) (ret string, err error) {
	var conn *net.TCPConn

	if err = func() (err error) {
		network, host := mdnet.ResolveDNSIP("tcp", host)
		fmt.Println(network, host)
		ad, err := net.ResolveTCPAddr("tcp", host)
		if err != nil {
			cmd.logger.Errorf("error occurred when dial remote dns server: %v\n", err)
			return
		}
		conn, err = net.DialTCP(network, nil, ad)
		if err != nil {
			cmd.logger.Errorf("error occurred when dial remote dns server: %v\n", err)
			return
		}
		fmt.Println(conn.LocalAddr(), conn.RemoteAddr())
		return
	}(); err != nil {
		return
	}

	cmd.logger.Infof("tcp socket set up successfully")
	defer func() {
		cmd.logger.Infof("disConnected from remote DNS server")
		err = conn.Close()
	}()

	requestNames := [][]byte{[]byte(req)}
	requsetTypes := []uint16{1}

	request := msg.Quest(
		requestNames,
		requsetTypes,
	)

	for len(request) != 0 {
		n, s := msg.NewDNSMessageRecursivelyQuery(1, request)
		request = request[n:]

		fmt.Println(n, s)
		b, err := s.CompressToBytes()
		if err != nil {
			cmd.logger.Errorf("convert request message error: %v", err)
			return "", err
		}
		cmd.logger.Infof("Writing")
		var lb = make([]byte, 2)
		binary.BigEndian.PutUint16(lb, uint16(len(b)))
		fmt.Println(lb, b)
		if _, err := conn.Write(lb); err != nil {
			cmd.logger.Errorf("write error: %v", err)
			return "", err
		}
		if _, err := conn.Write(b); err != nil {
			cmd.logger.Errorf("write error: %v", err)
			return "", err
		}
		cmd.logger.Infof("reading")
		time.Sleep(time.Second)
		b = make([]byte, 65535)
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err = conn.Read(b)
		if err != nil && err != io.EOF {
			cmd.logger.Errorf("read error: %v", err)
			return "", err
		}
		var x uint16
		binary.Read(bytes.NewBuffer(b), binary.BigEndian, &x)
		b = b[2 : 2+x]
		fmt.Println(x, b)

		var rmsg = new(msg.DNSMessage)
		n, err = rmsg.Read(b)
		if err != nil {
			cmd.logger.Errorf("convert read message error: %v", err)
			return "", err
		}
		fmt.Println(n, err)
		rmsg.Print()
	}
	return "", nil
}
