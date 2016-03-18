package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"
	"strconv"

	"github.com/LEW21/siren/imagectl"
)

func ReadMetadata(commands_in [][]string) (id, tag, name, version, base string, commands [][]string, err error) {
	commands = commands_in

	var idCmd, fromCmd []string

	done := false
	for len(commands) > 0 && !done {
		switch commands[0][0] {
			case "ID":
				idCmd = commands[0]
				commands = commands[1:]
			case "FROM":
				fromCmd = commands[0]
				commands = commands[1:]
			default:
				done = true
		}
	}

	if idCmd == nil {
		err = errors.New("No ID command.")
		return
	}

	if len(idCmd) >= 2 {
		name = idCmd[1]
	} else {
		err = errors.New("ID requires at least one argument.")
		return
	}

	if len(idCmd) >= 3 {
		version = idCmd[2]
	}

	// Systemd allows only 3 special characters in machine names: ".", "-", "_".
	// We need one of them - and leave the other two to the users.
	// And we can't take "." as it is commonly used in version numbers.
	tag = name
	if version != "" {
		tag = tag + "-" + version
	}

	// UnixNano has 64 bytes. 16 values are stored in 4 bytes.
	// This means we always use 64/4 = 16-character identifiers.
	id = tag + "-" + strconv.FormatInt(time.Now().UnixNano(), 16)

	if fromCmd != nil {
		if len(fromCmd) >= 2 {
			base = fromCmd[1]
		} else {
			err = errors.New("FROM requires at least one argument.")
			return
		}
	}

	return
}

func Build(directory, tag string, writer io.Writer) (image imagectl.Image, ret_tag string, ok bool) {
	EnsureSirenDirExists()

	defer func(){
		if r := recover(); r != nil {
			ok = false
		}
	}()

	var sirenfile []byte
	func(){
		task := NewTask(writer, "Reading Sirenfile"); defer task.Finish()
		var err error
		sirenfile, err = ioutil.ReadFile(directory + "/Sirenfile")
		task.Require(err)
	}()

	var commands [][]string
	func(){
		task := NewTask(writer, "Parsing Sirenfile"); defer task.Finish()
		var err error
		commands, err = ParseSirenfile(string(sirenfile))
		task.Require(err)
	}()

	var id, base string
	//ret tag
	func(){
		task := NewTask(writer, "Reading metadata"); defer task.Finish()
		var tag2 string
		var err error
		id, tag2, _, _, base, commands, err = ReadMetadata(commands)
		if tag == "" {
			tag = tag2
		}
		task.Require(err)
	}()

	//ret image
	func(){
		task := NewTask(writer, "Creating an image: " + id); defer task.Finish()

		ictl, err := imagectl.New()
		task.Require(err)
		image, err = ictl.CreateImage(id, base)
		task.Require(err)
	}()

	func(){
		task := NewTask(writer, "Building the image"); defer task.Finish()
		b := BuildContext{task, image, directory}
		for _, cmd := range commands {
			b.SubtaskExec(cmd)
		}
	}()

	NewTask(writer, "Cleaning up the container").RequireAndFinish(moveSystemdConfigToUsr(image))
	NewTask(writer, "Unmounting").RequireAndFinish(image.SetReady(false))

	func(){
		task := NewTask(writer, "Reducing layer size"); defer task.Finish()
		image.Optimize(func (status string){fmt.Fprintln(task, status)}, func(err error){fmt.Fprintln(task, err)})
	}()

	NewTask(writer, "Freezing").RequireAndFinish(image.SetReadOnly(true))
	NewTask(writer, "Mounting").RequireAndFinish(image.SetReady(true))

	if tag != "-" {
		func(){
			task := NewTask(writer, "Tagging"); defer task.Finish()
			imagectl.UnTag(tag)
			task.Require(imagectl.Tag(tag, image))
		}()
	}

	return image, tag, true
}

func moveSystemdConfigToUsr(i imagectl.Image) error {
	cmd := i.Command("mkdir", "-p", "/etc/systemd/system", "/etc/systemd/user", "/etc/systemd/network")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New(err.Error() + ": " + string(out))
	}

	cmd = i.Command("cp", "-R", "/etc/systemd/system", "/etc/systemd/user", "/etc/systemd/network", "/usr/lib/systemd")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New(err.Error() + ": " + string(out))
	}

	cmd = i.Command("rm", "-R", "/etc/systemd/system", "/etc/systemd/user", "/etc/systemd/network")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New(err.Error() + ": " + string(out))
	}

	return nil
}
