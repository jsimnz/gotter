package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

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

func concatWithSpace(strs ...string) string {
	var final string
	for i, str := range strs {
		if i > 0 {
			final += " " + str
		} else {
			final += str
		}
	}
	return final
}

func getPathType(path string) repoType {
	if strings.HasPrefix(path, "http://") {
		return HTTP
	} else if strings.HasPrefix(path, "https://") {
		return HTTPS
	} else if strings.HasPrefix(path, "git://") {
		return GIT
	} else if strings.Contains(path, "@") {
		return SSH
	} else {
		return GOPKG
	}
}

func getSSHPath(path, user string) string {
	pkgpath := projectFromURL(path)
	pkgpath = user + "@" + pkgpath
	if strings.Count(pkgpath, "/") > 2 {
		pkgpath = pkgpath[:strings.LastIndex(pkgpath, "/")]
	}
	pkgpath = replaceAt(pkgpath, strings.Index(pkgpath, "/"), ":")
	pkgpath += ".git"

	return pkgpath
}

func parseGitOriginURL(line string) (string, error) {
	buf := bytes.NewBuffer([]byte(line))
	var giturl string
	_, err := fmt.Fscanf(buf, "origin %v (push)", &giturl)
	if err != nil {
		return "", err
	}
	return giturl, nil
}
