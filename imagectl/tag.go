package imagectl

import (
	"errors"
	"os"
)

func Tag(tag string, image Image) error {
	if err := os.Symlink(image.Name(), "/var/lib/machines/" + tag); err != nil {
		return err
	}

	return nil
}

var ErrNotATag = errors.New("not a tag")

func ReadTag(tag string) (target string, err error) {
	target, err = os.Readlink("/var/lib/machines/" + tag);
	if err != nil {
		return "", ErrNotATag
	}
	return
}

func UnTag(tag string) (target string, err error) {
	target, err = ReadTag(tag)
	if err != nil {
		return
	}

	err = os.Remove("/var/lib/machines/" + tag)
	return
}
