package imagectl

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"github.com/coreos/go-systemd/unit"
	systemd "github.com/coreos/go-systemd/dbus"
)

func setupMount(sd *systemd.Conn, what, where, fstype, options string) error {
	// mount -t $type -o$options $what $where

	sdname := unit.UnitNamePathEscape(where)
	sdfile := "[Mount]\nWhat=" + what + "\nWhere=" + where + "\nType=" + fstype + "\nOptions=" + options + "\n"
	sdauto := "[Automount]\nWhere=" + where + "\n\n[Install]\nWantedBy=local-fs.target\n"

	if err := ioutil.WriteFile("/etc/systemd/system/" + sdname + ".mount", []byte(sdfile), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile("/etc/systemd/system/" + sdname + ".automount", []byte(sdauto), 0644); err != nil {
		return err
	}

	if err := sd.Reload(); err != nil {
		return err
	}

	possible, _, err := sd.EnableUnitFiles([]string{sdname + ".automount"}, false, false)
	if !possible {
		return errors.New("Internal error: Auto-generated " + sdname + ".automount does not have an [Install] section.")
	}
	if err != nil {
		return err
	}

	_, err = sd.StartUnit(sdname + ".automount", "replace", nil)
	if err != nil {
		return err
	}

	return nil
}

func destroyMount(sd *systemd.Conn, where string) error {
	sdname := unit.UnitNamePathEscape(where)

	_, err := sd.StopUnit(sdname + ".automount", "replace", nil)
	if err != nil {
		return err
	}

	_, err = sd.DisableUnitFiles([]string{sdname + ".automount"}, false)
	if err != nil {
		return err
	}

	if err := os.Remove("/etc/systemd/system/" + sdname + ".mount"); err != nil {
		return err
	}
	if err := os.Remove("/etc/systemd/system/" + sdname + ".automount"); err != nil {
		return err
	}

	if err := sd.Reload(); err != nil {
		return err
	}

	return nil
}

func setupMountOverlay(sd *systemd.Conn, whatRo []string, whatRW, where string) error {
	if err := os.MkdirAll("/var/lib/image-layers/.work", 0700); err != nil {
		return err
	}

	// I guess overlayfs was designed in Australia.
	whatRoReversed := make([]string, 0, len(whatRo))
	for i := len(whatRo)-1; i >= 0; i-- {
		whatRoReversed = append(whatRoReversed, whatRo[i])
	}
	whatRo = whatRoReversed

	if whatRW == "" {
		switch len(whatRo) {
			case 0:
				return setupMount(sd, "tmpfs", where, "tmpfs", "ro")
			case 1:
				return setupMount(sd, whatRo[0], where, "none", "bind,ro") // Note: ro probably doesn't work
			default:
				return setupMount(sd, "overlay", where, "overlay", "lowerdir=" + strings.Join(whatRo, ":"))
		}
	} else {
		switch len(whatRo) {
			case 0:
				return setupMount(sd, whatRW, where, "none", "bind")
			default:
				return setupMount(sd, "overlay", where, "overlay", "lowerdir=" + strings.Join(whatRo, ":") + ",upperdir=" + whatRW + ",workdir=/var/lib/image-layers/.work")
		}
	}
}

func doUnitOperation(op func(name, mode string, ch chan<- string) (int, error), name, mode string) error {
	done := make(chan string)
	_, err := op(name, mode, done)
	if err != nil {
		return err
	}
	result := <-done
	if result != "done" {
		return errors.New(result)
	}

	return nil
}

func mount(sd *systemd.Conn, where string) error {
	sdname := unit.UnitNamePathEscape(where)
	return doUnitOperation(sd.StartUnit, sdname + ".mount", "replace")
}

func remount(sd *systemd.Conn, where string) error {
	sdname := unit.UnitNamePathEscape(where)
	return doUnitOperation(sd.RestartUnit, sdname + ".mount", "replace")
}

func umount(sd *systemd.Conn, where string) error {
	sdname := unit.UnitNamePathEscape(where)
	return doUnitOperation(sd.StopUnit, sdname + ".mount", "replace")
}
