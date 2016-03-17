package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"net/url"

	"github.com/coreos/go-systemd/unit"
	"github.com/LEW21/siren/imagectl"
)

func Pull(uri, tag string, writer io.Writer) (image imagectl.Image, ret_tag string, ok bool) {
	defer func(){
		if r := recover(); r != nil {
			ok = false
		}
	}()

	var u *url.URL
	func(){
		task := NewTask(writer, "Parsing URI"); defer task.Finish()
		var err error
		u, err = url.Parse(uri)
		task.Require(err)
	}()

	dir := ""
	if strings.Contains(u.Path, ":") {
		splitted := strings.SplitN(u.Path, ":", 2)
		u.Path = splitted[0]
		dir = splitted[1]
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
	}

	uri = u.String()
	repoRoot := "/var/lib/siren/" + unit.UnitNameEscape(uri)
	sourceRoot := repoRoot
	if dir != "" {
		sourceRoot = repoRoot + "/" + dir
	}

	fi, err := os.Stat(repoRoot)
	if err != nil {
		func(){
			task := NewTask(writer, "Cloning"); defer task.Finish()
			task.Require(task.RunCommand("git", "clone", u.String(), repoRoot))
		}()
	} else {
		func(){
			task := NewTask(writer, "Updating"); defer task.Finish()
			task.Assert(fi.IsDir(), errors.New(repoRoot + " is not a directory."))
			task.Require(task.RunCommand("git", "-C", repoRoot, "pull"))
		}()
	}

	return Build(sourceRoot, tag, writer)
}
