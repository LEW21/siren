package main

import (
	"os/exec"
	"github.com/LEW21/go-machine1"
)

type Image interface {
	Id() string
	Root() string
	RealPath(path string) string
	Command(name string, arg ...string) *exec.Cmd

	Exists() (bool, error) // Does machined know about our image?
	// NOTE: Layer may exist even if image does not exist.

	Alive() (bool, error)  // Is our image used as a machined's container?

	// machine1 dbus service properties
	ReadOnly() (bool, error)
}

func ImageRoot(i Image) string {
	return "/var/lib/machines/" + i.Id()
}

func ImageRealPath(i Image, path string) string {
	if len(path) > 0 && path[0] != '/' {
		panic("RealPath: The argument has to be an absolute path.")
	}
	return i.Root() + path
}

func ImageCommand(i Image, name string, arg ...string) *exec.Cmd {
	cacheDirs := []string{"/var/cache/pacman/pkg", "/var/cache/pip/http"}

	args := make([]string, 0, len(cacheDirs) * 2 + 3 + len(arg))

	for _, dir := range cacheDirs {
		args = append(args, "--bind", dir)
	}

	args = append(args, "-M", i.Id(), name)
	args = append(args, arg...)

	return exec.Command("systemd-nspawn", args...)
}

func ImageExists(i Image) (bool, error) {
	md, err := machine1.New()
	if err != nil {
		return false, err
	}

	_, err = md.GetImage(i.Id())
	if err == nil {
		return true, nil
	}
	if isDbusError(err, "org.freedesktop.machine1.NoSuchImage") {
		return false, nil
	}

	return false, err
}

func ImageAlive(i Image) (bool, error) {
	md, err := machine1.New()
	if err != nil {
		return false, err
	}

	_, err = md.GetMachine(i.Id())
	if err == nil {
		return true, nil
	}
	if isDbusError(err, "org.freedesktop.machine1.NoSuchMachine") {
		return false, nil
	}

	return false, err
}

func ImageReadOnly(i Image) (bool, error) {
	md, err := machine1.New()
	if err != nil {
		return false, err
	}

	mdi, err := md.GetImage(i.Id())
	if err != nil {
		return false, err
	}
	return mdi.ReadOnly()
}
