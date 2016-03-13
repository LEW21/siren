package imagectl

import (
	"os/exec"
)

type StdImage struct {
	name, path, type_ string
	readOnly, ready, alive bool
}

// Properties

func (i *StdImage) Name() string {
	return i.name
}

func (i *StdImage) Path() string {
	return i.path
}

func (i *StdImage) Type() string {
	return i.type_
}

func (i *StdImage) ReadOnly() bool {
	return i.readOnly
}

func (i *StdImage) Ready() bool {
	return i.ready
}

func (i *StdImage) Alive() bool {
	return i.alive
}

// Utilities

func (i *StdImage) RealPath(path string) string {
	return ImageRealPath(i, path)
}

func (i *StdImage) Command(name string, arg ...string) *exec.Cmd {
	return ImageCommand(i, name, arg...)
}

// Soft actions

func (i *StdImage) Update() error {
	return nil
}

// Hard actions

func (i *StdImage) Remove() error {
	if i.Alive() {
		return ErrImageAlive
	}

	return nil
}

func (i *StdImage) SetReadOnly(readOnly bool) error {
	if i.ReadOnly() == readOnly {
		return nil
	}

	if i.Alive() {
		return ErrImageAlive
	}

	// ...
	i.readOnly = readOnly
	return nil
}

func (i *StdImage) SetReady(ready bool) error {
	if i.Ready() == ready {
		return nil
	}

	if i.Alive() {
		return ErrImageAlive
	}

	// ...
	i.ready = ready
	return nil
}

func (i *StdImage) Optimize(statusCb func(string), errorCb func(error)) {
	if i.Alive() {
		errorCb(ErrImageAlive)
		return
	}
}
