package main

import (
	"fmt"

	dnsrv "github.com/Myriad-Dreamin/go-dns/server"
	log "github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli"
)

const (
	MAX_PORT_NUM = 65535
)

type ServerStartCmd struct {
	parentCmd *ServerCmd
	logger    *log.Entry

	seed    string
	outfile string
	datadir string
	wltname string
	show    bool
}

func (srv *ServerStartCmd) RequestRootLogger() *log.Logger {
	return srv.parentCmd.RequestRootLogger()
}

func (cmd *ServerStartCmd) Before(c *urcli.Context) (err error) {
	cmd.logger = cmd.parentCmd.logger
	return nil
}

func convertPortFromInt(port int) (string, error) {
	if port > MAX_PORT_NUM {
		return "", fmt.Errorf("input port exceed max port number")
	}
	return ":" + string(port), nil
}

func (cmd *ServerStartCmd) Action(c *urcli.Context) error {
	var dnsServer = new(dnsrv.Server)
	dnsServer.SetLogger(cmd.RequestRootLogger())

	if err := dnsServer.ListenAndServe(cmd.parentCmd.host); err != nil {
		return err
	}
	return nil
}

func NewServerStartCmd(srv *ServerCmd) urcli.Command {
	var srvStart = &ServerStartCmd{parentCmd: srv}
	return urcli.Command{
		Name:        "start",
		ShortName:   "start",
		Usage:       "server api",
		UsageText:   "start a new server",
		Category:    "server",
		Action:      srvStart.Action,
		Before:      srvStart.Before,
		After:       nil,
		Subcommands: nil,
	}
}
