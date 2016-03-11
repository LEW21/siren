package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

func Build(context BuildContext) error {
	sirenfile, err := ioutil.ReadFile(context.Directory + "/Sirenfile")
	if err != nil {
		return err
	}

	commands, err := ParseSirenfile(string(sirenfile))
	if err != nil {
		return err
	}

	metaCommands := map[string]bool{"ID":true, "FROM":true}

	imageCreated := false
	needImage := func() error {
		if !imageCreated {
			fmt.Fprintln(context.Stderr)
			fmt.Fprintln(context.Stderr, "# Creating an image: " + context.Image.Id())

			if err := context.Image.Create(); err != nil {
				return err
			}

			imageCreated = true
		}

		return nil
	}

	for _, cmd := range commands {
		if !metaCommands[cmd[0]] {
			if err := needImage(); err != nil {
				return err
			}
		}

		fmt.Fprintln(context.Stderr)
		fmt.Fprintln(context.Stderr, "# " + cmd[0] + " (" + strings.Join(cmd[1:], ") (") + ")")

		if err := context.Exec(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}

	if err := needImage(); err != nil {
		return err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Cleaning up the container...")

	if err := moveSystemdConfigToUsr(context.Image); err != nil {
		fmt.Fprintln(context.Stderr, err.Error())
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Unmounting...")

	if err := context.Image.UnMount(); err != nil {
		fmt.Fprintln(context.Stderr, err.Error())
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Reducing layer size...")

	OptimizeLayer(context.Image, func(status string){fmt.Fprintln(context.Stderr, status)})

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Freezing...")

	if err := context.Image.Freeze(); err != nil {
		return err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Mounting...")

	if err := context.Image.Mount(); err != nil {
		return err
	}

	return nil
}

func moveSystemdConfigToUsr(i Image) error {
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
