/*
 Copyright 2015 CoreOS Inc.
 Copyright 2016 Janusz Lewandowski

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 */

// Integration with the systemd machined API.  See http://www.freedesktop.org/wiki/Software/systemd/machined/
package machine1

import (
	"os"
	"strconv"

	"github.com/godbus/dbus"
)

const (
	dbusInterface = "org.freedesktop.machine1.Manager"
	dbusPath      = "/org/freedesktop/machine1"

	imageInterface = "org.freedesktop.machine1.Image"
)

// Conn is a connection to systemds dbus endpoint.
type Conn struct {
	conn   *dbus.Conn
	object dbus.BusObject
}

type Image struct {
	Object dbus.BusObject
}

type Machine struct {
	object dbus.BusObject
}

// New() establishes a connection to the system bus and authenticates.
func New() (*Conn, error) {
	c := new(Conn)

	if err := c.initConnection(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Conn) initConnection() error {
	var err error
	c.conn, err = dbus.SystemBusPrivate()
	if err != nil {
		return err
	}

	// Only use EXTERNAL method, and hardcode the uid (not username)
	// to avoid a username lookup (which requires a dynamically linked
	// libc)
	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = c.conn.Auth(methods)
	if err != nil {
		c.conn.Close()
		return err
	}

	err = c.conn.Hello()
	if err != nil {
		c.conn.Close()
		return err
	}

	c.object = c.conn.Object("org.freedesktop.machine1", dbus.ObjectPath(dbusPath))

	return nil
}

// RegisterMachine registers the container with the systemd-machined
func (c *Conn) RegisterMachine(name string, id []byte, service string, class string, pid int, root_directory string) error {
	return c.object.Call(dbusInterface+".RegisterMachine", 0, name, id, service, class, uint32(pid), root_directory).Err
}

func (c *Conn) GetImage(name string) (dbus.BusObject, error) {
	call := c.object.Call(dbusInterface+".GetImage", 0, name)
	if call.Err != nil {
		return nil, call.Err
	}

	return c.conn.Object("org.freedesktop.machine1", call.Body[0].(dbus.ObjectPath)), nil
}

func (c *Conn) GetMachine(name string) (dbus.BusObject, error) {
	call := c.object.Call(dbusInterface+".GetMachine", 0, name)
	if call.Err != nil {
		return nil, call.Err
	}

	return c.conn.Object("org.freedesktop.machine1", call.Body[0].(dbus.ObjectPath)), nil
}

type ImageInfo struct {
	Name, Type string
	ReadOnly bool
	CreationTimestamp, ModificationTimestamp, Usage uint64
	Object dbus.BusObject
}

func (c *Conn) ListImages() ([]ImageInfo, error) {
	call := c.object.Call(dbusInterface+".ListImages", 0)
	if call.Err != nil {
		return nil, call.Err
	}

	vlist := call.Body[0].([][]interface{})
	list := make([]ImageInfo, 0, len(vlist))

	for _, v := range vlist {
		list = append(list, ImageInfo{
			v[0].(string), v[1].(string),
			v[2].(bool),
			v[3].(uint64), v[4].(uint64), v[5].(uint64),
			c.conn.Object("org.freedesktop.machine1", v[6].(dbus.ObjectPath)),
		})
	}

	return list, nil
}
