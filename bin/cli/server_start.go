package main

import (
	config "github.com/Myriad-Dreamin/go-dns/config"
	dnsrv "github.com/Myriad-Dreamin/go-dns/server"
	log "github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli"
)

type ServerStartCmd struct {
	parentCmd *ServerCmd
	logger    *log.Entry
}

func (srv *ServerStartCmd) RequestRootLogger() *log.Logger {
	return srv.parentCmd.RequestRootLogger()
}

func (cmd *ServerStartCmd) Before(c *urcli.Context) (err error) {
	cmd.logger = cmd.parentCmd.logger
	return nil
}

func (cmd *ServerStartCmd) Action(c *urcli.Context) error {
	var dnsServer = new(dnsrv.Server)
	dnsServer.SetLogger(cmd.RequestRootLogger())
	dnsServer.SetConfig(config.Config())

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
