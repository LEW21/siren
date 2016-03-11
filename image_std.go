package main

import (
	"errors"
	"os"
	"os/exec"
)

type StdImage struct {
	id string
}

// Image interface

func (i StdImage) Id() string {
	return i.id
}

func (i StdImage) Root() string {
	return ImageRoot(i)
}

func (i StdImage) RealPath(path string) string {
	return ImageRealPath(i, path)
}

func (i StdImage) Command(name string, arg ...string) *exec.Cmd {
	return ImageCommand(i, name, arg...)
}

func (i StdImage) Exists() (bool, error) {
	return ImageExists(i)
}

func (i StdImage) Alive() (bool, error) {
	return ImageAlive(i)
}

func (i StdImage) ReadOnly() (bool, error) {
	return ImageReadOnly(i)
}

// Loader

var ErrNotADirectory = errors.New("not a directory")

func LoadStdImage(name string) (StdImage, error) {
	tmp := StdImage{name}

	fi, err := os.Stat(tmp.Root())
	if err != nil {
		return StdImage{}, err
	}
	// TODO Add support for traversing symlinks
	if !fi.IsDir() {
		return StdImage{}, ErrNotADirectory
	}
	return StdImage{name}, nil
}
