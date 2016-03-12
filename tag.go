package main

import (
	"errors"
	"os"
	systemd "github.com/coreos/go-systemd/dbus"
)

func Tag(tag string, image Image) error {
	if err := os.Symlink(image.Id(), "/var/lib/siren/" + tag); err != nil {
		return err
	}

	// TODO Fix https://github.com/systemd/systemd/issues/2001 and use symlink.

	if err := os.Mkdir("/var/lib/machines/" + tag, 0700); err != nil {
		return err
	}

	sd, err := systemd.New()
	if err != nil {
		return err
	}

	if err := setupMountOverlay(sd, nil, image.Root(), "/var/lib/machines/" + tag); err != nil {
		return err
	}

	if err := mount(sd, "/var/lib/machines/" + tag); err != nil {
		return err
	}

	return nil
}

var ErrNotATag = errors.New("not a tag")

func UnTag(tag string) error {
	if _, err := os.Readlink("/var/lib/siren/" + tag); err != nil {
		return ErrNotATag
	}

	sd, err := systemd.New()
	if err != nil {
		return err
	}

	if err := umount(sd, "/var/lib/machines/" + tag); err != nil {
		return err
	}

	if err := destroyMount(sd, "/var/lib/machines/" + tag); err != nil {
		return err
	}

	if err := os.Remove("/var/lib/machines/" + tag); err != nil {
		return err
	}

	if err := os.Remove("/var/lib/siren/" + tag); err != nil {
		return err
	}

	return nil
}
