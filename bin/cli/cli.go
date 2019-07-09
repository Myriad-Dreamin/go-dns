package main

import (
	"fmt"
	"io"
	"os"

	scrlog "github.com/Myriad-Dreamin/screenrus"
	log "github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli"
)

const (
	SrvName   = "go-dns"
	Usage     = "local dns server"
	UsageText = "local dns server"
	Version   = "0.1.0"
)

type ServerX struct {
	handler *urcli.App

	// submodules
	serve *ServerCmd

	logToScreen   bool
	logfiledir    string
	logfile       *os.File
	loggerFactory *log.Logger
	logger        *log.Entry
}

var screenLog, _ = scrlog.NewScreenLogPlugin(nil)

func (srv *ServerX) SetRootLogger(rd io.Writer) {
	srv.loggerFactory = log.New()
	srv.loggerFactory.Out = rd
	if srv.logToScreen {
		srv.loggerFactory.AddHook(screenLog)
	}

	srv.logger = srv.loggerFactory.WithFields(log.Fields{
		"prog": "cmd",
	})
}

func (srv *ServerX) RequestRootLogger() *log.Logger {
	return srv.loggerFactory
}

func (srv *ServerX) SetLog(loggerFactory *log.Logger) {
	srv.loggerFactory = loggerFactory
	if srv.logToScreen {
		srv.loggerFactory.AddHook(screenLog)
	}
	srv.logger = srv.loggerFactory.WithFields(log.Fields{
		"prog": "cmd",
	})
}

func (srv *ServerX) Before(c *urcli.Context) (err error) {
	srv.logfile, err = os.OpenFile(srv.logfiledir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 666)
	if err != nil {
		srv.logfile = nil
		return err
	}
	srv.SetRootLogger(srv.logfile)
	return nil
}

func (srv *ServerX) After(c *urcli.Context) error {
	return nil
}
func (srv *ServerX) SetInfo() {
	srv.handler.Name = SrvName
	srv.handler.Usage = Usage
	srv.handler.UsageText = UsageText
	srv.handler.Version = Version
}

func (srv *ServerX) Init() {
	srv.SetInfo()

	srv.handler.Before = srv.Before
	srv.handler.Action = nil
	srv.handler.After = srv.After

	srv.handler.Flags = []urcli.Flag{
		urcli.StringFlag{
			Name:        "logdir, ld",
			Value:       "prog.log",
			Usage:       "logger address",
			Destination: &srv.logfiledir,
		},
		urcli.BoolFlag{
			Name:        "logtoscr, lscr",
			Usage:       "logger to screen",
			Destination: &srv.logToScreen,
		},
	}
	urcli.HelpFlag = urcli.BoolFlag{
		Name:  "help, h",
		Usage: "show manual",
	}

	srv.serve = NewServerCmd(srv)
	srv.handler.Commands = []urcli.Command{
		*srv.serve.cmd,
	}

}

func (srv *ServerX) CommandNotFound(c *urcli.Context, cmdString string) {
	fmt.Println("command not found,", cmdString)
}

func (srv *ServerX) Stop() {
	if srv.logfile != nil {
		srv.logfile.Close()
	}
}

func NewServerX() *ServerX {
	return &ServerX{
		handler: urcli.NewApp(),
	}
}

func (srv *ServerX) Run() {
	if err := srv.handler.Run(os.Args); err != nil {
		if srv.logger == nil {
			srv.SetRootLogger(os.Stdout)
		}
		srv.logger.Fatal(err)
	}
}

func (srv *ServerX) CliExit(status int) {
	fmt.Println("prog exit with", status)
	srv.Stop()
	os.Exit(status)
}

func main() {
	var srv = NewServerX()
	urcli.OsExiter = srv.CliExit
	srv.Init()
	srv.Run()
	srv.Stop()
}
