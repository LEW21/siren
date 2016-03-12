package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var Commands = []Command{CmdBuild, CmdCreate, CmdRemove, CmdFreeze, CmdUnFreeze, CmdMount, CmdUnMount}

var CmdBuild = Command{"build", []string{"PATH"}, "Build an image from a Sirenfile", cmdBuild}
func cmdBuild(args []string) int {
	thisName := args[0]

	dir, err := filepath.Abs(thisName)
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

	b := BuildContext{SirenImage{"", nil, false}, dir, outw, errw}

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

	err = Build(&b)
	if err != nil {
		fmt.Fprintln(errw, err)
		return 1
	}

	fmt.Println("Image created.")
	fmt.Println("Use 'siren create instance_name " + b.Image.Id() + "' to create a new, writable machine image using this image as a base.")
	return 0
}

var CmdCreate = Command{"create", []string{"NAME", "BASE_NAME"}, "Create a new image", cmdCreate}
func cmdCreate(args []string) int {
	thisName := args[0]
	baseName := args[1]

	base := Image(nil)
	if baseName != "" {
		var err error
		base, err = LoadStdImage(baseName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Base image does not exist.")
			return 1
		}
	}
	this := SirenImage{thisName, base, false}

	if err := this.Create(); err != nil {
		switch err {
			case ErrBaseWritable:
				fmt.Fprintln(os.Stderr, "Base image is writable. Freeze it first.")
				return 1
			case ErrImageExists:
				fmt.Fprintln(os.Stderr, "Image already exists. Refusing to overwrite.")
				return 1
			default:
				panic(err)
		}
	}

	fmt.Println("Image created.")
	fmt.Println("Use machinectl start " + thisName + " to start the container.")
	return 0
}

func printWarnings(problems []error) {
	for _, warning := range problems {
		fmt.Fprintln(os.Stderr, "Warning: ", warning)
	}
}

func printChangeError(err error) int {
	switch err {
		case ErrImageAlive:
			fmt.Fprintln(os.Stderr, "Image is currently running as a machine. Stop it first.")
			return 1
		default:
			panic(err)
	}
}

var CmdRemove = Command{"remove", []string{"NAME"}, "Remove an image", cmdRemove}
func cmdRemove(args []string) int {
	thisName := args[0]

	this, _, err := LoadSirenImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.Remove(); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image removed.")
	return 0
}

var CmdFreeze = Command{"freeze", []string{"NAME"}, "Mark image read-only", cmdFreeze}
func cmdFreeze(args []string) int {
	thisName := args[0]

	this, problems, err := LoadSirenImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	printWarnings(problems)

	if err := this.Freeze(); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image frozen.")
	return 0
}

var CmdUnFreeze = Command{"unfreeze", []string{"NAME"}, "Mark image read-write", cmdUnFreeze}
func cmdUnFreeze(args []string) int {
	thisName := args[0]

	this, problems, err := LoadSirenImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	printWarnings(problems)

	if err := this.UnFreeze(); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image unfrozen.")
	return 0
}

var CmdMount = Command{"mount", []string{"NAME"}, "Mount the image in /var/lib/machines", cmdMount}
func cmdMount(args []string) int {
	thisName := args[0]

	this, problems, err := LoadSirenImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	printWarnings(problems)

	if err := this.Mount(); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image mounted.")
	return 0
}

var CmdUnMount = Command{"unmount", []string{"NAME"}, "Unmount the image from /var/lib/machines", cmdUnMount}
func cmdUnMount(args []string) int {
	thisName := args[0]

	this, problems, err := LoadSirenImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	printWarnings(problems)

	if err := this.UnMount(); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image unmounted.")
	return 0
}
