package imagectl

type ImageCtl struct {
	mctl MachineCtl
	lictl LayeredImageCtl
}

func New() (ImageCtl, error) {
	ictl := ImageCtl{}
	err := error(nil)
	ictl.mctl, err = NewMachineCtl()
	if err != nil {
		return ImageCtl{}, err
	}
	ictl.lictl, err = NewLayeredImageCtl(ictl.mctl)
	if err != nil {
		return ImageCtl{}, err
	}
	return ictl, nil
}

func (ictl ImageCtl) GetImage(name string) (Image, error) {
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

func (ictl ImageCtl) CreateImage(name, baseName string) (Image, error) {
	i, err := ictl.lictl.CreateImage(name, baseName)
	if err == nil {
		return &i, nil
	} else {
		return nil, err
	}
}
