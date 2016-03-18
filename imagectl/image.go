package imagectl

import (
	"errors"
	"os/exec"
	"github.com/LEW21/siren/imagectl/machine1"
)

type Image interface {
	// org.freedesktop.machine1.Image properties (subset)
	Name() string
	Path() string
	Type() string
	ReadOnly() bool

	// Custom properties
	Ready() bool // Is ready to use? (In case of layered images: is it mounted?)
	Alive() bool // Is our image used as a machined's container?

	// Soft actions
	Update() error // Reload all properties

	// Hard actions
	Remove() error
	SetReadOnly(bool) error
	SetReady(bool) error
	Optimize(statusCb func(string), errorCb func(error))

	// Utilities
	RealPath(path string) string
	Command(name string, arg ...string) *exec.Cmd
}

var ErrNoSuchImage = errors.New("org.freedesktop.machine1.NoSuchImage") // for GetImage() and Update
var ErrImageAlive = errors.New("image is alive") // for hard actions
var ErrImpossible = errors.New("impossible") // for SetReady(true)
var ErrImageExists = errors.New("image already exists") // for CreateImage()
var ErrBaseWritable = errors.New("base image is writable") // for CreateImage()
var ErrBaseDoesNotExist = errors.New("base image does not exist") // for CreateImage()

func ImageRealPath(i Image, path string) string {
	if len(path) > 0 && path[0] != '/' {
		panic("RealPath: The argument has to be an absolute path.")
	}
	return i.Path() + path
}

func ImageCommand(i Image, name string, arg ...string) *exec.Cmd {
	cacheDirs := []string{"/var/cache/pacman/pkg", "/var/cache/pip/http"}

	args := make([]string, 0, len(cacheDirs) * 2 + 3 + len(arg))

	for _, dir := range cacheDirs {
		args = append(args, "--bind", dir)
	}

	args = append(args, "-M", i.Name(), name)
	args = append(args, arg...)

	return exec.Command("systemd-nspawn", args...)
}

func IsImageAlive(md *machine1.Conn, name string) (bool, error) {
	_, err := md.GetMachine(name)
	if isDbusError(err, "org.freedesktop.machine1.NoSuchMachine") {
		return false, nil
	}
	return true, err
}
