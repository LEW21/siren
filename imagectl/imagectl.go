package imagectl

import (
	"sort"
)

type ImageCtl struct {
	mctl MachineCtl
	lictl LayeredImageCtl
}

func New() (*ImageCtl, error) {
	ictl := &ImageCtl{}
	err := error(nil)
	ictl.mctl, err = NewMachineCtl()
	if err != nil {
		return nil, err
	}
	ictl.lictl, err = NewLayeredImageCtl(ictl.mctl.md, ictl.GetImage)
	if err != nil {
		return nil, err
	}
	return ictl, nil
}

func (ictl *ImageCtl) ListImages() ([]Image, error) {
	image_map := make(map[string]Image)

	mctl_images, err := ictl.mctl.ListImages()
	if err != nil {
		return nil, err
	}
	for _, i := range mctl_images {
		i_copy := i
		image_map[i.Name()] = &i_copy
	}

	lictl_images, err := ictl.lictl.ListImages()
	if err != nil {
		return nil, err
	}
	for _, i := range lictl_images {
		i_copy := i
		image_map[i.Name()] = &i_copy
	}

	names := make([]string, 0, len(image_map))
	for k := range image_map {
		names = append(names, k)
	}
	sort.Strings(names)

	images := make([]Image, 0, len(names))
	for _, name := range names {
		images = append(images, image_map[name])
	}

	return images, nil
}

func (ictl *ImageCtl) GetImage(name string) (Image, error) {
	{
		i, err := ictl.lictl.GetImage(name)
		if err == nil {
			return &i, nil
		} else if err != ErrNoSuchImage {
			return nil, err
		}
	}

	{
		i, err := ictl.mctl.GetImage(name)
		if err == nil {
			return &i, nil
		} else {
			return nil, err
		}
	}
}

func (ictl *ImageCtl) CreateImage(name string, base Image) (Image, error) {
	i, err := ictl.lictl.CreateImage(name, base)
	if err == nil {
		return &i, nil
	} else {
		return nil, err
	}
}
