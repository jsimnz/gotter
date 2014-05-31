package main

import (
	"errors"
	"os"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

const (
	PENDING = iota
	FAIL
	SUCCESS

	// TODO - DEFAULT_CONF_LOCATION = "/etc/gotter/gotter.conf"
)

var (
	exitStatus int

	log = logging.MustGetLogger("gotter")

	WORKSPACE = os.Getenv("WORKSPACE")
	GOPATH    = os.Getenv("GOPATH")
)

func main() {
	initLogger()

	app := cli.NewApp()
	app.Name = "gotter"
	app.Author = "John-Alan Simmons <simmons.johnalan@gmail.com>"
	app.Usage = "Utlity to unify and manage Go projects into a single workspace"
	app.Version = "0.1.0-rc1"

	// Overwrite default 'version' shorthand 'v' flag
	cli.VersionFlag.Name = "version"
	app.Flags = []cli.Flag{
		cli.BoolFlag{"verbose, v", "Enable verbose logging"},
		cli.BoolFlag{"extra-verbose, vv", "Enable more verbose logging"},
	}
	app.Before = func(c *cli.Context) error {
		cmd := c.Args().First()
		if cmd == "" || cmd == "help" || cmd == "h" {
			return nil
		}

		if c.Bool("extra-verbose") {
			logging.SetLevel(logging.DEBUG, "gotter")
		} else if c.Bool("verbose") {
			logging.SetLevel(logging.INFO, "gotter")
		} else {
			logging.SetLevel(logging.WARNING, "gotter")
		}

		// Make sure environment variable is set
		if WORKSPACE == "" {
			log.Error("[ERROR]: WORKSPACE enviromental variable not set!")
			return errors.New("WORKSPACE enviromental variable not set!")
		}
		if GOPATH == "" {
			log.Error("[ERROR]: GOPATH enviromental variable not set!")
			return errors.New("GOPATH enviromental variable not set!")
		}

		return nil
	}

	app.Commands = []cli.Command{getCommand, cloneCommand, linkCommand, updateRemoteCommand}

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
