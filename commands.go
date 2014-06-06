package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	getCommandDesc = `Clones a package into your GOPATH using the go tool chain, 
   creates a link between it and your workspace, and if possible updates 
   the repos remote origin to use SSH.`

	getCommand = cli.Command{
		Name:        "get",
		Usage:       "'go get' a repo, and link it to your workspace",
		Description: getCommandDesc,
		Action: func(c *cli.Context) {
			err := getCommandAction(c)
			if err != nil {
				exitStatus = FAIL
				return
			}
		},
		Flags: []cli.Flag{
			cli.BoolFlag{"update, u", "Update existing code"},
			cli.BoolFlag{"download-only, d", "Only download the code, don't install with the go toolchain (go build, go install)"},
			cli.BoolFlag{"force, f", "Force updating and linking (Irreverseible)"},
			cli.BoolFlag{"no-ssh", "Do not update the remote origin to use SSH"},
			cli.StringFlag{"ssh-user", "<user>", "Set the user for the SSH url (Default: git)"},
		},
	}

	cloneCommand = cli.Command{
		Name:  "clone",
		Usage: "Clone the repo into your GOPATH",
		Action: func(c *cli.Context) {
			err := cloneCommandAction(c)
			if err != nil {
				exitStatus = FAIL
				return
			}
		},
		Flags: []cli.Flag{
			cli.BoolFlag{"update, u", "Update existing code"},
			cli.BoolFlag{"download-only, d", "Only download the code, don't install with the go toolchain (go build, go install)"},
		},
	}

	linkCommand = cli.Command{
		Name:  "link",
		Usage: "Create a link from the GOPATH/project to WORKSPACE/project",
		Action: func(c *cli.Context) {
			err := linkCommandAction(c)
			if err != nil {
				exitStatus = FAIL
				return
			}
		},
		Flags: []cli.Flag{
			cli.BoolFlag{"update, u", "Update existing link"},
			cli.BoolFlag{"force, f", "Force updating and linking (Irreverseible)"},
		},
	}

	updateRemoteCommand = cli.Command{
		Name:  "update-remote",
		Usage: "Updates the git remote origin url to use SSH",
		Action: func(c *cli.Context) {
			err := updateRemoteCommandAction(c)
			if err != nil {
				exitStatus = FAIL
				return
			}
		},
		Flags: []cli.Flag{
			cli.StringFlag{"ssh-user", "git", "Set the user for the SSH url (Default: git)"},
		},
	}
)

func getCommandAction(c *cli.Context) error {
	err := cloneCommandAction(c)
	if err != nil {
		return err
	}
	err = linkCommandAction(c)
	if err != nil {
		return err
	}
	if !c.Bool("no-ssh") {
		err = updateRemoteCommandAction(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func cloneCommandAction(c *cli.Context) error {
	pkgpath := projectFromURL(c.Args().First())
	log.Info("Getting package: %v", pkgpath)
	args := []string{"get"}
	if c.Bool("update") {
		args = append(args, "-u")
		//log.Debug(" ----> running %v", concat("go", " ", "get", " -u ", pkgpath))
		//err = pipeFromExec(os.Stdout, "go", "get", "-u", pkgpath)
	}
	if c.Bool("download-only") {
		args = append(args, "-d")
		//log.Debug(" ----> running %v", concat("go", " ", "get", " ", pkgpath))
		//err = pipeFromExec(os.Stdout, "go", "get", pkgpath)
	}
	args = append(args, pkgpath)

	log.Debug(" ----> running go %v", concatWithSpace(args...))
	err := pipeFromExec(os.Stdout, "go", args...)
	if err != nil {
		log.Error("Couldn't get package: %v", err)
		return err
	}

	log.Debug(" ----> Successfully got package!")
	return nil
}

func linkCommandAction(c *cli.Context) error {
	pkgpath := projectFromURL(c.Args().First())
	pkg := pkgFromPath(pkgpath)
	workspacepath := concat(WORKSPACE, "/", pkg)
	log.Info("Linking package %v to %v", pkgpath, workspacepath)

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
				return errors.New("Link already exists")
			}
		} else if c.Bool("force") {
			log.Warning(" ----> removing %v", workspacepath)
			os.Remove(fullworkspacepath)
		} else {
			log.Error(" ----> [ERROR]: File/Folder already exists at %v, if you wish to proceed use -f", workspacepath)
			return errors.New("File/Folder already exists")
		}
	}

	cmd := exec.Command("ln", "-s", fullpkgpath, fullworkspacepath)
	log.Debug(" ----> running: %v", concat("ln", " -s", concat(" $GOPATH/src/", pkgpath), concat(" ", WORKSPACE, "/", pkg)))
	err := cmd.Run()
	if err != nil {
		log.Error("Failed to create link: %v", err)
		return err
	}
	log.Debug(" ----> Successfully linked!")
	return nil
}

func updateRemoteCommandAction(c *cli.Context) error {
	path := c.Args().First()
	pkgpath := projectFromURL(path)
	fullpkgpath := getAbsPath(concat(GOPATH, "/src/", pkgpath))
	log.Info("Update remote origin URL for repo: %v", pkgpath)

	os.Chdir(fullpkgpath)
	cmd := exec.Command("git", "remote", "-v")
	log.Debug(" ----> running: git remote -v")

	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		log.Error("Couldn't update remote url: %v", err)
		return err
	}

	endpoints := strings.Split(buf.String(), "\n")
	var repo string
	for _, line := range endpoints {
		url, err := parseGitOriginURL(line)
		if err == nil {
			repo = url
		}
	}
	if repo == "" {
		log.Error("Couldn't parse git remote origin url")
		return errors.New("Couldn't parse git remote origin url")
	}

	user := c.String("ssh-user")
	newRepoURL := getSSHPath(repo, user)
	os.Chdir(fullpkgpath)
	cmd = exec.Command("git", "remote", "set-url", "origin", newRepoURL)
	log.Debug(" ----> running: git remote set-url origin %v", newRepoURL)
	err = cmd.Run()
	if err != nil {
		log.Error("Failed to update remote origin: %v", err)
		return err
	}
	log.Debug(" ----> Successfully updated remote origin")
	return nil
}
