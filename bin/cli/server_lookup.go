package main

import (
	"fmt"

	dnsrv "github.com/Myriad-Dreamin/go-dns/server"
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
	var dnsServer = new(dnsrv.Server)
	dnsServer.SetLogger(cmd.RequestRootLogger())

	if x, err := dnsServer.LookUpA(cmd.parentCmd.host, cmd.a); err != nil {
		fmt.Println(x)
		return err
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
		// Flags: []urcli.Flag
		// 	urcli.StringFlag{
		// 		Name:        "a",
		// 		Value:       "",
		// 		Usage:       "address",
		// 		Destination: &serve.a,
		// 	},
		// },
	}
}
