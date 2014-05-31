package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codegangsta/cli"
)

const (
	GOPKG = repoType(iota)
	HTTP
	HTTPS
	GIT
	SSH
)

type repoType int

var (
	getCommand = cli.Command{
		Name:  "get",
		Usage: "'go get' a repo, and link it to your workspace",
		Description: `Clones a package into your GOPATH using the go tool chain, and 
   creates a link between it and your workspace, and if possible updates 
   the repos remote origin to use SSH.`,
		Action: getCommandAction,
		Flags: []cli.Flag{
			cli.BoolFlag{"update, u", "Update existing code"},
			cli.BoolFlag{"force, f", "Force updating and linking (Irreverseible)"},
		},
	}

	cloneCommand = cli.Command{
		Name:   "clone",
		Usage:  "Clone the repo into your GOPATH",
		Action: cloneCommandAction,
		Flags: []cli.Flag{
			cli.BoolFlag{"update, u", "Update existing code"},
		},
	}

	linkCommand = cli.Command{
		Name:   "link",
		Usage:  "Create a link from the GOPATH/project to WORKSPACE/project",
		Action: linkCommandAction,
		Flags: []cli.Flag{
			cli.BoolFlag{"update, u", "Update existing link"},
			cli.BoolFlag{"force, f", "Force updating and linking (Irreverseible)"},
		},
	}

	updateRemoteCommand = cli.Command{
		Name:  "update-remote",
		Usage: "Updates the git remote origin url to use SSH",
	}
)

func getCommandAction(c *cli.Context) {
	cloneCommandAction(c)
	linkCommandAction(c)

}

func cloneCommandAction(c *cli.Context) {
	pkgpath := projectFromURL(c.Args().First())
	log.Debug("Getting package: %v", pkgpath)
	if c.Bool("update") {
		log.Debug(" ----> running %v", concat("go", " ", "get", " -u ", pkgpath))
		err := pipeFromExec(os.Stdout, "go", "get", "-u", pkgpath)
		if err != nil {
			panic(err)
		}
	} else {
		log.Debug(" ----> running %v", concat("go", " ", "get", " ", pkgpath))
		err := pipeFromExec(os.Stdout, "go", "get", pkgpath)
		if err != nil {
			panic(err)
		}
	}
}

func linkCommandAction(c *cli.Context) {
	pkgpath := projectFromURL(c.Args().First())
	pkg := pkgFromPath(pkgpath)
	workspacepath := concat(WORKSPACE, "/", pkg)
	log.Debug("Linking package %v to %v", pkgpath, workspacepath)

	fullpkgpath := getAbsPath(concat(GOPATH, "/src/", pkgpath))
	fullworkspacepath := getAbsPath(workspacepath)

	// check if anything exists here
	if _, err := os.Stat(fullworkspacepath); !os.IsNotExist(err) {
		info, _ := os.Lstat(fullworkspacepath)
		if info.Mode()&os.ModeSymlink != 0 {
			if c.Bool("update") || c.Bool("force") {
				os.Remove(fullworkspacepath)
			} else {
				log.Warning("[WARNING]: Link already exists!")
				symlink, _ := filepath.EvalSymlinks(fullworkspacepath)
				log.Warning(" ----> %v -> %v", workspacepath, symlink)
				return
			}
		} else if c.Bool("force") {
			log.Warning(" ----> removing %v", workspacepath)
			os.Remove(fullworkspacepath)
		} else {
			log.Error(" ----> [ERROR]: File/Folder already exists at %v, if you wish to proceed use -f", workspacepath)
			exitStatus = FAIL
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
	log.Debug(" ----> Successfully linked!")
}

func updateRemoteCommandAction(c *cli.Context) {
	repo := c.Args().First()
	typ := getPathType(repo)
	if typ != SSH {
		repo = getSSHPath(repo)
	}
	log.Debug("Updating remote origin URL to: %v", repo)
	cmd := exec.Command("git", "remote", "set-url", "origin", repo)
	log.Debug(" ----> running: git remote set-url origin %v", repo)
	err := cmd.Run()
	if err != nil {
		log.Error("Couldnt update repo origin url!")
		exitStatus = FAIL
		return
	}
}
