package main

import (
	urcli "github.com/urfave/cli"
)

type ServerStartCmd struct {
	parentCmd *ServerCmd
	seed      string
	outfile   string
	datadir   string
	wltname   string
	show      bool
}

func (srv *ServerStartCmd) Action(c *urcli.Context) error {

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
		After:       nil,
		Subcommands: nil,
	}
}
