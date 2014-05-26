package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

const (
	PENDING = iota
	FAIL
	SUCCESS
)

var (
	log        = logging.MustGetLogger("gotter")
	exitStatus int

	WORKSPACE = "~/Workspace"
	GOPATH    = os.Getenv("GOPATH")
)

func main() {
	app := cli.NewApp()
	app.Name = "gotter"
	app.Author = "John-Alan Simmons"
	app.Usage = "Utlity to unify and manage Go projects into a single workspace"
	app.Version = "0.0.3"

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

	getCommand := cli.Command{
		Name:  "get",
		Usage: "'go get' a repo, and link it to your workspace",
		Description: `Clones a package into your GOPATH using the go tool chain, 
   creates a link between it and your workspace, and if possible updates the 
   repos remote origin to use SSH`,
		Action: getCommandAction,
	}

	cloneCommand := cli.Command{
		Name:   "clone",
		Usage:  "Clone the repo into your GOPATH",
		Action: cloneCommandAction,
	}

	linkCommand := cli.Command{
		Name:   "link",
		Usage:  "Create a link from the GOPATH/project to WORKSPACE/project",
		Action: linkCommandAction,
	}
	app.Commands = []cli.Command{getCommand, cloneCommand, linkCommand}

	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackend("")
	if err != nil {
		panic(err)
	}
	logging.SetBackend(logBackend, syslogBackend)

	defer func() {
		if exitStatus == FAIL {
			log.Error("Status: FAILED")
		}
	}()

	app.Run(os.Args)
}

func getCommandAction(c *cli.Context) {
	cloneCommandAction(c)
	linkCommandAction(c)
}

func cloneCommandAction(c *cli.Context) {
	pkgpath := projectFromURL(c.Args().First())
	log.Debug("Getting package: %v", pkgpath)
	log.Debug(" ----> running %v", concat("go", " ", "get", " ", pkgpath))
	err := pipeFromExec(os.Stdout, "go", "get", pkgpath)
	if err != nil {
		panic(err)
	}
}

func linkCommandAction(c *cli.Context) {
	pkgpath := projectFromURL(c.Args().First())
	pkg := pkgFromPath(pkgpath)
	workspacepath := concat(WORKSPACE, "/", pkg)
	log.Debug("Linking package %v to %v", pkgpath, workspacepath)

	fullpkgpath := getAbsPath(concat(GOPATH, "/src/", pkgpath))
	fullworkspacepath := getAbsPath(workspacepath)

	if info, err := os.Stat(fullworkspacepath); os.IsExist(err) {
		fmt.Println("File exists")
		if info.Mode()&os.ModeSymlink == 0 {
			log.Warning("[WARNING]: Link already exists!")
			symlink, _ := filepath.EvalSymlinks(fullworkspacepath)
			log.Warning(" ----> %v -> %v", fullworkspacepath, symlink)
			return
		}
	}
	cmd := exec.Command("ln", "-s", fullpkgpath, fullworkspacepath)
	log.Debug(" ----> running: %v", concat("ln", " -s", concat(" $GOPATH/src/", pkgpath), concat(" ", WORKSPACE, "/", pkg)))
	err := cmd.Run()
	if err != nil {
		log.Error("Failed to create link: %v", err)
		exitStatus = FAIL
		return
	}
}

func updateRemoteOrigin(gitpath string) {}

func projectFromURL(path string) string {
	if strings.HasSuffix(path, ".git") {
		path = path[:len(path)-4]
	}

	if !strings.HasPrefix(path, "http://") &&
		!strings.HasPrefix(path, "https://") &&
		!strings.HasPrefix(path, "git://") {
		path = "http://" + path
	}

	if index := strings.LastIndex(path, ":"); index != strings.Index(path, "://") && index > 0 {
		path = replaceAt(path, index, "/")
	}

	u, _ := url.Parse(path)

	return u.Host + u.Path
}

func pkgFromPath(path string) string {
	index := strings.LastIndex(path, "/")
	return path[index+1:]
}

func getAbsPath(path string) string {
	if strings.HasPrefix(path, "~") {
		usr, _ := user.Current()
		path = strings.Replace(path, "~", usr.HomeDir, 1)
	}

	abspath, _ := filepath.Abs(path)
	return abspath
}

func replaceAt(s string, at int, with string) string {
	return s[:at] + with + s[at+1:]
}

func pipeFromExec(dst io.Writer, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)
	err = c.Start()
	if err != nil {
		return err
	}
	err = c.Wait()
	if err != nil {
		return err
	}

	return nil
}

func concat(strs ...string) string {
	var final string
	for _, str := range strs {
		final += str
	}
	return final
}
