package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Build(path string) (image SirenImage, tag string, ok bool) {
	dir, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	outr, outw, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	errr, errw, err := os.Pipe()
	if err != nil {
		outw.Close()
		outr.Close()
		panic(err)
	}

	b := BuildContext{SirenImage{"", nil, false}, dir, outw, errw, "", ""}

	outWritten := make(chan bool)
	go func() {
		defer func(){outWritten <- true}()
		defer outr.Close()
		io.Copy(os.Stdout, outr)
	}()
	defer func(){<-outWritten}()
	defer outw.Close()

	errWritten := make(chan bool)
	go func() {
		defer func(){errWritten <- true}()
		defer errr.Close()
		io.Copy(os.Stderr, errr)
	}()
	defer func(){<-errWritten}()
	defer errw.Close()

	image, tag, err = build(&b)
	if err != nil {
		fmt.Fprintln(errw, err)
		return SirenImage{}, "", false
	}

	return image, tag, true
}

func build(context *BuildContext) (SirenImage, string, error) {
	sirenfile, err := ioutil.ReadFile(context.Directory + "/Sirenfile")
	if err != nil {
		return SirenImage{}, "", err
	}

	commands, err := ParseSirenfile(string(sirenfile))
	if err != nil {
		return SirenImage{}, "", err
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
				return SirenImage{}, "", err
			}
		}

		fmt.Fprintln(context.Stderr)
		fmt.Fprintln(context.Stderr, "# " + cmd[0] + " (" + strings.Join(cmd[1:], ") (") + ")")

		if err := context.Exec(cmd[0], cmd[1:]...); err != nil {
			return SirenImage{}, "", err
		}
	}

	if err := needImage(); err != nil {
		return SirenImage{}, "", err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Cleaning up the container...")

	if err := moveSystemdConfigToUsr(context.Image); err != nil {
		return SirenImage{}, "", err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Unmounting...")

	if err := context.Image.UnMount(); err != nil {
		return SirenImage{}, "", err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Reducing layer size...")

	OptimizeLayer(context.Image, func(status string){fmt.Fprintln(context.Stderr, status)})

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Freezing...")

	if err := context.Image.Freeze(); err != nil {
		return SirenImage{}, "", err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Mounting...")

	if err := context.Image.Mount(); err != nil {
		return SirenImage{}, "", err
	}

	fmt.Fprintln(context.Stderr)
	fmt.Fprintln(context.Stderr, "# Tagging...")

	tag := context.name
	if context.version != "" {
		tag = tag + "-" + context.version
	}

	UnTag(tag)
	if err := Tag(tag, context.Image); err != nil {
		return SirenImage{}, "", err
	}

	return context.Image, tag, nil
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
