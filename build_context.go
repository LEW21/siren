package main

import (
	"fmt"
	"time"
	"io"
	"io/ioutil"
	"os/exec"
	"errors"
	"strings"
	"strconv"
)

type BuildContext struct {
	Image SirenImage

	Directory string

	Stdout io.Writer
	Stderr io.Writer

	name string
	version string
}

func (b BuildContext) RealPath(path string) string {
	if path[0] == '/' {
		panic("BuildContext.RealPath: The argument has to be a relative path.")
	}
	return b.Directory + "/" + path
}

func (b *BuildContext) Id(name, version string) error {
	// UnixNano has 64 bytes. 16 values are stored in 4 bytes.
	// This means we always use 64/4 = 16-character identifiers.
	b.Image.SetId(name, version, strconv.FormatInt(time.Now().UnixNano(), 16))
	b.name = name
	b.version = version
	return nil
}

func (b *BuildContext) From(baseName string) error {
	base, err := LoadStdImage(baseName)
	if err != nil {
		return err
	}
	b.Image.base = base
	return nil
}

func (b *BuildContext) runCmd(cmd *exec.Cmd) error {
	fmt.Fprintln(b.Stderr, "## " + cmd.Args[0] + " (" + strings.Join(cmd.Args[1:], ") (") + ")")

	cmd.Stdout = b.Stdout
	cmd.Stderr = b.Stderr

	return cmd.Run()
}

func (b *BuildContext) runCommand(name string, arg ...string) error {
	return b.runCmd(exec.Command(name, arg...))
}

func (b *BuildContext) Run(name string, arg ...string) error {
	return b.runCmd(b.Image.Command(name, arg...))
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
	return b.runCommand("cp", args...)
}

func (b *BuildContext) Untar(arg ...string) error {
	dst := arg[len(arg)-1]
	src := arg[:len(arg)-1]

	for _, tarfile := range src {
		err := b.runCommand("tar", "-xf", b.RealPath(tarfile), "-C", b.Image.RealPath(dst))
		if err != nil {
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
		fmt.Fprintln(b.Stderr, "Warning: " + name + " file not found.")
	}
	return b.Run("systemctl", "enable", name)
}

var ErrNotEnoughArguments = errors.New("not enough arguments")

// Don't error out when we get too many arguments - we can use them to extend the commands in the future.

func (b *BuildContext) Exec(command string, arg ...string) error {
	switch (command) {
		case "ID":
			switch len(arg) {
				case 0:
					return ErrNotEnoughArguments
				case 1:
					return b.Id(arg[0], "")
			}
			return b.Id(arg[0], arg[1])

		case "FROM":
			return b.From(arg[0])

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
