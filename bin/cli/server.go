package main

import (
	"fmt"

	urcli "github.com/urfave/cli"
)

type ServerCmd struct {
	srv *ServerX
	cmd *urcli.Command

	port int
	host string
}

func (serve *ServerCmd) Before(c *urcli.Context) (err error) {
	fmt.Println("serve Before")
	return nil
}

func (serve *ServerCmd) After(c *urcli.Context) (err error) {
	fmt.Println("serve After")
	return nil
}

func (serve *ServerCmd) MakeCommands() urcli.Commands {
	return []urcli.Command{
		NewServerStartCmd(serve),
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
			urcli.IntFlag{
				Name:        "port, p",
				Value:       23335,
				Usage:       "listening port",
				Destination: &serve.port,
			},
			urcli.StringFlag{
				Name:        "port, p",
				Value:       "114.114.114.114",
				Usage:       "parent dns address",
				Destination: &serve.host,
			},
		},
	}
	return serve
}
