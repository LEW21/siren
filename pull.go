package main

import (
	"errors"
	"io"
	"os"
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

	var subpath string
	u.Path, subpath = SplitSubPath(u.Path)

	uri = u.String()
	repoRoot := "/var/lib/siren/" + unit.UnitNameEscape(uri)
	sourceRoot := repoRoot
	if subpath != "" {
		sourceRoot = repoRoot + "/" + subpath
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
