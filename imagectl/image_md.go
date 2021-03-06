package imagectl

import (
	"os"
	"os/exec"
	"github.com/godbus/dbus"
	"github.com/LEW21/siren/imagectl/machine1"
)

type MachineCtl struct {
	md *machine1.Conn
}

func NewMachineCtl() (MachineCtl, error) {
	md, err := machine1.New()
	return MachineCtl{md}, err
}

func (mctl MachineCtl) GetImage(name string) (MdImage, error) {
	target, err := ReadTag(name)
	if err == nil {
		name = target
	}

	mdImage, err := mctl.md.GetImage(name)
	if isDbusError(err, "org.freedesktop.machine1.NoSuchImage") {
		return MdImage{}, ErrNoSuchImage
	}
	if err != nil {
		return MdImage{}, err
	}

	i := MdImage{"", "", "", false, false, false, mctl.md, mdImage}
	err = i.Update()
	return i, err
}

func (mctl MachineCtl) listImages_Dbus() ([]MdImage, error) {
	mlist, err := mctl.md.ListImages()
	if err != nil {
		return nil, err
	}
	list := make([]MdImage, 0, len(mlist))
	for _, mdImageInfo := range mlist {
		i := MdImage{mdImageInfo.Name, "/var/lib/machines/" + mdImageInfo.Name, mdImageInfo.Type, mdImageInfo.ReadOnly, true, true, mctl.md, mdImageInfo.Object}
		i.alive, _ = IsImageAlive(mctl.md, i.Name())
		list = append(list, i)
	}
	return list, nil
}

func (mctl MachineCtl) ListImages() ([]MdImage, error) {
	f, err := os.Open("/var/lib/machines")
	if err != nil {
		return nil, err
	}
	dirs, err := f.Readdir(-1)
	f.Close()

	images := make([]MdImage, 0, len(dirs))
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		if i, err := mctl.GetImage(dir.Name()); err == nil {
			images = append(images, i)
		}
	}
	return images, nil
}

func (mctl MachineCtl) CreateImage(name string, base Image) (MdImage, error) {
	if base != nil && !base.ReadOnly() {
		return MdImage{}, ErrBaseWritable
	}

	if _, err := mctl.GetImage(name); err == nil {
		return MdImage{}, ErrImageExists
	}

	if base != nil {
		call := mctl.md.Object.Call("org.freedesktop.machine1.Manager.CloneImage", 0, base.Name(), name, false)
		if call.Err != nil {
			return MdImage{}, call.Err
		}

		return mctl.GetImage(name)
	}

	if err := os.Mkdir("/var/lib/machines/" + name, 0700); err != nil {
		return MdImage{}, err
	}
	return mctl.GetImage(name)
}

type MdImage struct {
	name, path, type_ string
	readOnly, ready, alive bool

	md *machine1.Conn
	mdImage dbus.BusObject
}

// Properties

func (i *MdImage) Name() string {
	return i.name
}

func (i *MdImage) Path() string {
	return i.path
}

func (i *MdImage) Type() string {
	return i.type_
}

func (i *MdImage) ReadOnly() bool {
	return i.readOnly
}

func (i *MdImage) Ready() bool {
	return i.ready
}

func (i *MdImage) Alive() bool {
	return i.alive
}

// Utilities

func (i *MdImage) RealPath(path string) string {
	return ImageRealPath(i, path)
}

func (i *MdImage) Command(name string, arg ...string) *exec.Cmd {
	return ImageCommand(i, name, arg...)
}

// Soft actions

func (i *MdImage) Update() error {
	call := i.mdImage.Call("org.freedesktop.DBus.Properties.GetAll", 0, "")
	if call.Err != nil {
		i.ready = false
		return ErrNoSuchImage
	}
	properties := call.Body[0].(map[string]dbus.Variant)

	i.name     = properties["Name"].Value().(string)
	i.path     = properties["Path"].Value().(string)
	i.type_    = properties["Type"].Value().(string)
	i.readOnly = properties["ReadOnly"].Value().(bool)
	i.ready    = true

	var err error
	i.alive, err = IsImageAlive(i.md, i.Name())
	return err
}

// Hard actions

func (i *MdImage) Remove() error {
	if i.Alive() {
		return ErrImageAlive
	}

	call := i.mdImage.Call("org.freedesktop.machine1.Image.Remove", 0)
	if call.Err != nil {
		i.ready = false
		return call.Err
	}
	return nil
}

func (i *MdImage) SetReadOnly(readOnly bool) error {
	if i.ReadOnly() == readOnly {
		return nil
	}

	if i.Alive() {
		return ErrImageAlive
	}

	call := i.mdImage.Call("org.freedesktop.machine1.Image.MarkReadOnly", 0, readOnly)
	if call.Err != nil {
		i.ready = false
		return call.Err
	}

	i.readOnly = readOnly
	return nil
}

func (i *MdImage) SetReady(ready bool) error {
	if i.Ready() == ready {
		return nil
	}

	if i.Alive() {
		return ErrImageAlive
	}

	if !i.ready {
		return ErrImpossible
	}

	// Unreadying machinectl images is not supported
	// (and not necessary)
	return nil
}

func (i *MdImage) Optimize(statusCb func(string), errorCb func(error)) {
	if i.Alive() {
		errorCb(ErrImageAlive)
		return
	}
}
