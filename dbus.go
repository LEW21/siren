package main

import (
	"github.com/godbus/dbus"
)

func isDbusError(err error, name string) bool {
	switch e := err.(type) {
		case dbus.Error:
			if e.Name == name {
				return true
			}
	}

	return false
}
