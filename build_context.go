package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/unit"
	"github.com/LEW21/siren/imagectl"
)

type BuildContext struct {
	Task *Task
	Image imagectl.Image
	Directory string
}

func (b BuildContext) RealPath(path string) string {
	if path[0] == '/' {
		panic("BuildContext.RealPath: The argument has to be a relative path.")
	}
	return b.Directory + "/" + path
}

func (b *BuildContext) Run(name string, arg ...string) error {
	return b.Task.RunCmd(b.Image.Command(name, arg...))
}

func (b *BuildContext) Copy(arg ...string) error {
	dst := arg[len(arg)-1]
	src := arg[:len(arg)-1]

	// TODO? add support for wildcards

	args := []string{"-R"}
	for _, srcdir := range src {
		args = append(args, b.RealPath(srcdir))
	}

	args = append(args, b.Image.RealPath(dst))
	return b.Task.RunCommand("cp", args...)
}

func (b *BuildContext) download(uri *url.URL) (res *url.URL, err error) {
	res = &url.URL{}

	res.Fragment = uri.Fragment
	uri.Fragment = ""

	res.Path = "/var/lib/siren/" + unit.UnitNameEscape(uri.String())

	{
		_, err := os.Stat(res.Path)
		if err == nil {
			return res, nil
		}
	}

	err = b.Task.RunCommand("wget", uri.String(), "-O", res.Path, "--progress=dot:mega")

	return res, err
}

func (b *BuildContext) Untar(arg ...string) error {
	dst := arg[len(arg)-1]
	src := arg[:len(arg)-1]

	for _, tarfile := range src {
		u, err := url.Parse(tarfile)
		if err != nil {
			return err
		}

		if u.Scheme == "http" || u.Scheme == "https" {
			var err error
			u, err = b.download(u)
			if err != nil {
				return err
			}
		} else {
			u.Path = b.RealPath(u.Path)
		}

		args := []string{"-xf", u.Path, "-C", b.Image.RealPath(dst)}

		if u.Fragment != "" {
			strip_components := 1 + strings.Count(u.Fragment, "/")
			args = append(args, u.Fragment, "--strip-components", strconv.Itoa(strip_components))
		}

		if err := b.Task.RunCommand("tar", args...); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildContext) Set(name, value string) error {
	return ioutil.WriteFile(b.Image.RealPath(name), []byte(value), 0644)
}

func unitName(name string) string {
	if strings.HasSuffix(name, ".service") || strings.HasSuffix(name, ".socket") {
		return name
	}
	return name + ".service"
}

func (b *BuildContext) addUnit(name string) error {
	return b.Copy(name, "/usr/lib/systemd/system/")
}

func (b *BuildContext) AddUnit(name string) error {
	return b.addUnit(unitName(name))
}

func (b *BuildContext) Enable(name string) error {
	name = unitName(name)
	if err := b.addUnit(name); err != nil {
		fmt.Fprintln(b.Task, "Warning: " + name + " file not found.")
	}
	return b.Run("systemctl", "enable", name)
}

var ErrNotEnoughArguments = errors.New("not enough arguments")

// Don't error out when we get too many arguments - we can use them to extend the commands in the future.

func (b *BuildContext) Exec(cmd []string) error {
	command := cmd[0]
	arg := cmd[1:]

	switch (command) {
		case "RUN":
			return b.Run(arg[0], arg[1:]...)

		case "COPY":
			return b.Copy(arg...)

		case "UNTAR":
			return b.Untar(arg...)

		case "SET":
			return b.Set(arg[0], arg[1])

		case "ADD_UNIT":
			return b.AddUnit(arg[0])

		case "ENABLE":
			return b.Enable(arg[0])

		default:
			return errors.New("Unknown command: " + command)
	}
}

func (b *BuildContext) SubtaskExec(cmd []string) {
	subtask := NewTask(b.Task, cmd[0] + " (" + strings.Join(cmd[1:], ") (") + ")"); defer subtask.Finish()
	maintask := b.Task
	b.Task = subtask; defer func(){b.Task = maintask}()

	b.Task.Require(b.Exec(cmd))
}
