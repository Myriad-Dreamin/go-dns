package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli"
)

type ServerCmd struct {
	srv    *ServerX
	logger *log.Entry

	cmd *urcli.Command

	host string
}

func (srv *ServerCmd) RequestRootLogger() *log.Logger {
	return srv.srv.RequestRootLogger()
}

func (serve *ServerCmd) Before(c *urcli.Context) (err error) {
	fmt.Println("serve Before")
	serve.logger = serve.srv.logger
	return nil
}

func (serve *ServerCmd) After(c *urcli.Context) (err error) {
	fmt.Println("serve After")
	return nil
}

func (serve *ServerCmd) MakeCommands() urcli.Commands {
	return []urcli.Command{
		NewServerStartCmd(serve),
		NewServerLookUpACmd(serve),
	}
}

func NewServerCmd(dnsSrv *ServerX) *ServerCmd {
	var serve = &ServerCmd{srv: dnsSrv}
	serve.cmd = &urcli.Command{
		Name:        "server",
		ShortName:   "srv",
		Usage:       "server api",
		UsageText:   "start or configure server",
		ArgsUsage:   "wait for edition",
		Category:    "server",
		Before:      serve.Before,
		Action:      nil,
		Subcommands: serve.MakeCommands(),
		Flags: []urcli.Flag{
			urcli.StringFlag{
				Name:        "host",
				Value:       "114.114.114.114",
				Usage:       "parent dns address",
				Destination: &serve.host,
			},
		},
	}
	return serve
}
