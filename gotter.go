package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

const (
	PENDING = iota
	FAIL
	SUCCESS

	DEFAULT_CONF_LOCATION = "/etc/gotter/gotter.conf"
)

var (
	exitStatus int

	log = logging.MustGetLogger("gotter")

	WORKSPACE = os.Getenv("WORKSPACE")
	GOPATH    = os.Getenv("GOPATH")
)

func main() {
	app := cli.NewApp()
	app.Name = "gotter"
	app.Author = "John-Alan Simmons"
	app.Usage = "Utlity to unify and manage Go projects into a single workspace"
	app.Version = "0.0.6"

	app.Flags = []cli.Flag{
		cli.BoolFlag{"verbose, V", "Enable verbose logging"},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			logging.SetLevel(logging.DEBUG, "gotter")
		} else {
			logging.SetLevel(logging.WARNING, "gotter")
		}

		return nil
	}

	app.Commands = []cli.Command{getCommand, cloneCommand, linkCommand, updateRemoteCommand}

	initLogger()

	defer func() {
		if exitStatus == FAIL {
			log.Error("Status: FAILED")
		}
	}()

	app.Run(os.Args)
}

func initLogger() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackend("")
	if err != nil {
		panic(err)
	}
	logging.SetBackend(logBackend, syslogBackend)
}
