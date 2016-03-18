package main

import (
	"errors"
	"io"
	"os"
	"net/url"
	"strings"

	"github.com/coreos/go-systemd/unit"
	"github.com/LEW21/siren/imagectl"
)

func Pull(ictl imagectl.ImageCtl, uri, tag string, writer io.Writer) (image imagectl.Image, ret_tag string, ok bool) {
	EnsureSirenDirExists()

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

		if u.Scheme == "git" || strings.HasPrefix(u.Scheme, "git+") {
			u.Scheme = strings.TrimPrefix(u.Scheme, "git+")
		} else {
			task.Require(errors.New("Unsupported scheme: " + u.Scheme))
		}
	}()

	fragment := u.Fragment
	u.Fragment = ""

	uri = u.String()
	repoRoot := "/var/lib/siren/" + unit.UnitNameEscape(uri)
	sourceRoot := repoRoot
	if fragment != "" {
		sourceRoot = repoRoot + "/" + fragment
	}

	fi, err := os.Stat(repoRoot)
	if err != nil {
		func(){
			task := NewTask(writer, "Cloning"); defer task.Finish()
			task.Require(task.RunCommand("git", "clone", uri, repoRoot))
		}()
	} else {
		func(){
			task := NewTask(writer, "Updating"); defer task.Finish()
			task.Assert(fi.IsDir(), errors.New(repoRoot + " is not a directory."))
			task.Require(task.RunCommand("git", "-C", repoRoot, "pull"))
		}()
	}

	return Build(ictl, sourceRoot, tag, writer)
}
