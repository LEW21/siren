package imagectl

import (
	"os"
	"os/exec"
	"io/ioutil"
	systemd "github.com/coreos/go-systemd/dbus"
)

type LayeredImageCtl struct {
	mctl MachineCtl
}

func NewLayeredImageCtl(mctl MachineCtl) (LayeredImageCtl, error) {
	return LayeredImageCtl{mctl}, nil
}

func (lictl LayeredImageCtl) GetImage(name string) (LayeredImage, error) {
	if target, err := ReadTag(name); err == nil {
		name = target
	}

	i := LayeredImage{name, nil, false, lictl.mctl, false, false}
	err := i.Update()
	return i, err
}

func (lictl LayeredImageCtl) ListImages() ([]LayeredImage, error) {
	f, err := os.Open("/var/lib/image-layers")
	if err != nil {
		return nil, err
	}
	dirs, err := f.Readdir(-1)
	f.Close()

	images := make([]LayeredImage, 0, len(dirs))
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		if i, err := lictl.GetImage(dir.Name()); err == nil {
			images = append(images, i)
		}
	}
	return images, nil
}

func (lictl LayeredImageCtl) CreateImage(name, baseName string) (LayeredImage, error) {
	base := (*MdImage)(nil)
	if baseName != "" {
		b, err := lictl.mctl.GetImage(baseName)
		if err != nil {
			if err == ErrNoSuchImage {
				err = ErrBaseDoesNotExist
			}
			return LayeredImage{}, err
		}
		if !b.ReadOnly() {
			return LayeredImage{}, ErrBaseWritable
		}
		base = &b
	}

	if _, err := lictl.mctl.GetImage(name); err == nil {
		return LayeredImage{}, ErrImageExists
	}

	i := LayeredImage{name, base, false, lictl.mctl, false, false}
	err := i.create()
	return i, err
}

type LayeredImage struct {
	id string
	base *MdImage
	frozen bool

	mctl MachineCtl
	ready, alive bool
}

// Image interface

func (i *LayeredImage) Name() string {
	return i.id
}

func (i *LayeredImage) Path() string {
	return "/var/lib/machines/" + i.Name()
}

func (i *LayeredImage) Type() string {
	return "layered"
}

func (i *LayeredImage) ReadOnly() bool {
	return i.frozen
}

func (i *LayeredImage) Ready() bool {
	return i.ready
}

func (i *LayeredImage) Alive() bool {
	return i.alive
}

func (i *LayeredImage) RealPath(path string) string {
	return ImageRealPath(i, path)
}

func (i *LayeredImage) Command(name string, arg ...string) *exec.Cmd {
	return ImageCommand(i, name, arg...)
}

// Soft actions

func (i *LayeredImage) Update() error {
	if _, err := ioutil.ReadFile(i.LayerPath("/id")); err != nil {
		return ErrNoSuchImage
	}

	i.base = nil
	baseName, _ := ioutil.ReadFile(i.LayerPath("/base"))
	if baseName != nil {
		b, err := i.mctl.GetImage(string(baseName))
		if err != nil {
			return err
		}
		i.base = &b
	}

	frozen, _ := ioutil.ReadFile(i.LayerPath("/frozen"))
	i.frozen = string(frozen) != "n"

	image, err := i.mctl.GetImage(i.Name())
	if err == ErrNoSuchImage {
		i.ready = false
		i.alive = false
	} else {
		i.ready = true
		i.alive = image.Alive()
	}

	return nil
}

// Hard actions

func (i *LayeredImage) Remove() error {
	if err := i.SetReady(false); err != nil {
		return err
	}

	return os.RemoveAll(i.LayerRoot())
}

func (i *LayeredImage) SetReadOnly(readOnly bool) error {
	if i.ReadOnly() == readOnly {
		return nil
	}

	ready := i.Ready()

	if ready {
		if err := i.SetReady(false); err != nil {
			return err
		}
	}

	i.frozen = readOnly
	i.saveMetadata()

	if ready {
		if err := i.SetReady(true); err != nil {
			return err
		}
	}

	return nil
}

func (i *LayeredImage) SetReady(ready bool) error {
	if i.Ready() == ready {
		return nil
	}

	if i.Alive() {
		return ErrImageAlive
	}

	sd, err := systemd.New()
	if err != nil {
		return err
	}

	if ready {
		err = i.setReady_Mount(sd)
	} else {
		err = i.setReady_UnMount(sd)
	}

	if err != nil {
		return err
	}

	i.ready = ready
	return nil
}

func (i *LayeredImage) Optimize(statusCb func(string), errorCb func(error)) {
	if i.Alive() {
		errorCb(ErrImageAlive)
		return
	}

	optimizeLayeredImage(i, statusCb, errorCb)
}

// Implementation

func (i LayeredImage) LayerRoot() string {
	return "/var/lib/image-layers/" + i.Name()
}

func (i LayeredImage) LayerFSRoot() string {
	return i.LayerRoot() + "/rootfs"
}

func (i LayeredImage) LayerPath(path string) string {
	if len(path) > 0 && path[0] != '/' {
		panic("PackagePath: The argument has to be an absolute path.")
	}
	return i.LayerRoot() + path
}

func (i *LayeredImage) create() error {
	if err := os.MkdirAll(i.LayerRoot(), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(i.LayerFSRoot(), 0755); err != nil {
		return err
	}

	if err := i.saveMetadata(); err != nil {
		return err
	}
	return i.SetReady(true)
}

func (i *LayeredImage) saveMetadata() error {
	if err := ioutil.WriteFile(i.LayerPath("/id"), []byte(i.Name()), 0644); err != nil {
		return err
	}

	if i.base != nil {
		if err := ioutil.WriteFile(i.LayerPath("/base"), []byte(i.base.Name()), 0644); err != nil {
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

// SetReady

func (i LayeredImage) setReady_Mount(sd *systemd.Conn) error {
	if err := i.setReady_setupMountPoints(sd); err != nil {
		return err
	}

	if err := i.setReady_mount(sd); err != nil {
		return err
	}

	return nil
}

func (i LayeredImage) setReady_UnMount(sd *systemd.Conn) error {
	if err := i.setReady_umount(sd); err != nil {
		return err
	}

	if err := i.setReady_destroyMountPoints(sd); err != nil {
		return err
	}

	return nil
}

func (i LayeredImage) setReady_setupMountPoints(sd *systemd.Conn) error {
	if err := os.MkdirAll(i.Path(), 0700); err != nil {
		return err
	}

	roLayers := []string{}
	rwLayer := i.LayerFSRoot()

	if i.base != nil {
		roLayers = []string{i.base.Path()}
	}

	if i.frozen {
		roLayers = append(roLayers, rwLayer)
		rwLayer = ""
	}

	if err := setupMountOverlay(sd, roLayers, rwLayer, i.Path()); err != nil {
		return err
	}

	return nil
}

func (i LayeredImage) setReady_destroyMountPoints(sd *systemd.Conn) error {
	if err := destroyMount(sd, i.Path()); err != nil {
		return err
	}

	if err := os.Remove(i.Path()); err != nil {
		return err
	}
	return nil
}

func (i LayeredImage) setReady_mount(sd *systemd.Conn) error {
	return mount(sd, i.Path())
}

func (i LayeredImage) setReady_remount(sd *systemd.Conn) error {
	return remount(sd, i.Path())
}

func (i LayeredImage) setReady_umount(sd *systemd.Conn) error {
	return umount(sd, i.Path())
}
