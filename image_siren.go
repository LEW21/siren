package main

import (
	"errors"
	"os"
	"os/exec"
	"io/ioutil"
	systemd "github.com/coreos/go-systemd/dbus"
)

type SirenImage struct {
	id string
	base Image
	frozen bool
}

// Image interface

func (i SirenImage) Id() string {
	return i.id
}

func (i SirenImage) Root() string {
	return ImageRoot(i)
}

func (i SirenImage) RealPath(path string) string {
	return ImageRealPath(i, path)
}

func (i SirenImage) Command(name string, arg ...string) *exec.Cmd {
	return ImageCommand(i, name, arg...)
}

func (i SirenImage) Exists() (bool, error) {
	return ImageExists(i)
}

func (i SirenImage) Alive() (bool, error) {
	return ImageAlive(i)
}

func (i SirenImage) ReadOnly() (bool, error) {
	return ImageReadOnly(i)
}

// Loader

func LoadSirenImage(name string) (SirenImage, []error, error) {
	tmp := SirenImage{name, nil, false}

	id, err := ioutil.ReadFile(tmp.LayerPath("/id"))
	if err != nil {
		return SirenImage{}, nil, err
	}

	problems := make([]error, 0, 2)

	base := Image(nil)
	baseName, _ := ioutil.ReadFile(tmp.LayerPath("/base"))
	if baseName != nil {
		base, err = LoadStdImage(string(baseName))
		if err != nil {
			problems = append(problems, err)
			base = nil
			err = nil
		}
	}

	frozen, err := ioutil.ReadFile(tmp.LayerPath("/frozen"))
	if err != nil {
		problems = append(problems, err)
		frozen = nil
		err = nil
	}
	isFrozen := string(frozen) != "n"

	return SirenImage{string(id), base, isFrozen}, problems, nil
}

// Custom methods

func (i SirenImage) LayerRoot() string {
	return "/var/lib/siren/" + i.Id()
}

func (i SirenImage) LayerFSRoot() string {
	return i.LayerRoot() + "/rootfs"
}

func (i SirenImage) LayerPath(path string) string {
	if len(path) > 0 && path[0] != '/' {
		panic("PackagePath: The argument has to be an absolute path.")
	}
	return i.LayerRoot() + path
}

func (i SirenImage) IsFrozen() bool {
	return i.frozen
}

func (i *SirenImage) SetFrozen(frozen bool) error {
	mounted, err := i.Exists()
	if err != nil {
		return err
	}

	if mounted {
		if err := i.UnMount(); err != nil {
			return err
		}
	}

	i.frozen = frozen
	i.SaveMetadata()

	if mounted {
		if err := i.Mount(); err != nil {
			return err
		}
	}

	return nil
}

func (i *SirenImage) Freeze() error {
	return i.SetFrozen(true)
}

func (i *SirenImage) UnFreeze() error {
	return i.SetFrozen(false)
}

func (i *SirenImage) Mount() error {
	alive, err := i.Alive()
	if err != nil {
		return err
	}
	if alive {
		return ErrImageAlive
	}

	sd, err := systemd.New()
	if err != nil {
		return err
	}

	if err := i.setupMountPoints(sd); err != nil {
		return err
	}

	if err := i.mount(sd); err != nil {
		return err
	}

	return nil
}

var ErrImageAlive = errors.New("image is alive")

func (i *SirenImage) UnMount() error {
	alive, err := i.Alive()
	if err != nil {
		return err
	}
	if alive {
		return ErrImageAlive
	}

	sd, err := systemd.New()
	if err != nil {
		return err
	}

	if err := i.umount(sd); err != nil {
		return err
	}

	if err := i.destroyMountPoints(sd); err != nil {
		return err
	}

	return nil
}

var ErrBaseWritable = errors.New("base image is writable")
var ErrImageExists = errors.New("image already exists")

func (i *SirenImage) Create() error {
	if i.base != nil {
		readOnly, err := i.base.ReadOnly()
		if err == nil && !readOnly {
			return ErrBaseWritable
		}
	}

	if _, err := LoadStdImage(i.Id()); err == nil {
		return ErrImageExists
	}

	if err := i.SaveMetadata(); err != nil {
		return err
	}

	if err := i.Mount(); err != nil {
		return err
	}

	return nil
}

func (i *SirenImage) Remove() error {
	mounted, err := i.Exists()
	if err != nil {
		return err
	}

	if mounted {
		if err := i.UnMount(); err != nil {
			return err
		}
	}

	return os.RemoveAll(i.LayerRoot())
}

func (i SirenImage) SaveMetadata() error {
	if err := os.MkdirAll(i.LayerRoot(), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(i.LayerFSRoot(), 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(i.LayerPath("/id"), []byte(i.Id()), 0644); err != nil {
		return err
	}

	if i.base != nil {
		if err := ioutil.WriteFile(i.LayerPath("/base"), []byte(i.base.Id()), 0644); err != nil {
			return err
		}
	}

	frozen := []byte{'n'}
	if i.frozen {
		frozen = []byte{'y'}
	}

	if err := ioutil.WriteFile(i.LayerPath("/frozen"), frozen, 0644); err != nil {
		return err
	}

	return nil
}

func (i SirenImage) setupMountPoints(sd *systemd.Conn) error {
	if err := os.MkdirAll(i.Root(), 0700); err != nil {
		return err
	}

	roLayers := []string{}
	rwLayer := i.LayerFSRoot()

	if i.base != nil {
		roLayers = []string{i.base.Root()}
	}

	if i.frozen {
		roLayers = append(roLayers, rwLayer)
		rwLayer = ""
	}

	if err := setupMountOverlay(sd, roLayers, rwLayer, i.Root()); err != nil {
		return err
	}

	return nil
}

func (i SirenImage) destroyMountPoints(sd *systemd.Conn) error {
	if err := destroyMount(sd, i.Root()); err != nil {
		return err
	}

	if err := os.Remove(i.Root()); err != nil {
		return err
	}
	return nil
}

func (i SirenImage) mount(sd *systemd.Conn) error {
	return mount(sd, i.Root())
}

func (i SirenImage) remount(sd *systemd.Conn) error {
	return remount(sd, i.Root())
}

func (i SirenImage) umount(sd *systemd.Conn) error {
	return umount(sd, i.Root())
}
