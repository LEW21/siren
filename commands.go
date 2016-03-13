package main

import (
	"fmt"
	"os"

	"github.com/LEW21/siren/imagectl"
)

var Commands = []Command{CmdBuild, CmdCreate, CmdRemove, CmdFreeze, CmdUnFreeze, CmdMount, CmdUnMount, CmdTag}

var CmdBuild = Command{"build", []string{"PATH"}, "Build an image from a Sirenfile", cmdBuild}
func cmdBuild(args []string) int {
	path := args[0]

	_, tag, ok := Build(path)
	if !ok {
		return 1
	}

	fmt.Println("Image created.")
	fmt.Println("Use 'siren create instance_name " + tag + "' to create a new, writable machine image using this image as a base.")
	return 0
}

var CmdCreate = Command{"create", []string{"NAME", "BASE_NAME"}, "Create a new image", cmdCreate}
func cmdCreate(args []string) int {
	thisName := args[0]
	baseName := args[1]

	ictl, err := imagectl.New()
	if err != nil {
		panic(err)
	}

	i, err := ictl.CreateImage(thisName, baseName)
	if err != nil {
		switch err {
			case imagectl.ErrBaseDoesNotExist:
				fmt.Fprintln(os.Stderr, "Base image does not exist.")
				return 1
			case imagectl.ErrBaseWritable:
				fmt.Fprintln(os.Stderr, "Base image is writable. Freeze it first.")
				return 1
			case imagectl.ErrImageExists:
				fmt.Fprintln(os.Stderr, "Image already exists. Refusing to overwrite.")
				return 1
			default:
				panic(err)
		}
	}

	if err := i.SetReady(true); err != nil {
		panic(err)
	}

	fmt.Println("Image created.")
	fmt.Println("Use machinectl start " + i.Name() + " to start the container.")
	return 0
}

func printWarnings(problems []error) {
	for _, warning := range problems {
		fmt.Fprintln(os.Stderr, "Warning: ", warning)
	}
}

func printChangeError(err error) int {
	switch err {
		case imagectl.ErrImageAlive:
			fmt.Fprintln(os.Stderr, "Image is currently running as a machine. Stop it first.")
			return 1
		default:
			panic(err)
	}
}

var CmdRemove = Command{"remove", []string{"NAME"}, "Remove an image or a tag", cmdRemove}
func cmdRemove(args []string) int {
	thisName := args[0]

	err := imagectl.UnTag(thisName)
	switch err {
		case nil:
			fmt.Println("Tag removed.")
			return 0

		case imagectl.ErrNotATag:
			err = nil
			break

		default:
			panic(err)
	}

	ictl, err := imagectl.New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
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

	ictl, err := imagectl.New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.SetReadOnly(true); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image frozen.")
	return 0
}

var CmdUnFreeze = Command{"unfreeze", []string{"NAME"}, "Mark image read-write", cmdUnFreeze}
func cmdUnFreeze(args []string) int {
	thisName := args[0]

	ictl, err := imagectl.New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.SetReadOnly(false); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image unfrozen.")
	return 0
}

var CmdMount = Command{"mount", []string{"NAME"}, "Mount the image in /var/lib/machines", cmdMount}
func cmdMount(args []string) int {
	thisName := args[0]

	ictl, err := imagectl.New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.SetReady(true); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image mounted.")
	return 0
}

var CmdUnMount = Command{"unmount", []string{"NAME"}, "Unmount the image from /var/lib/machines", cmdUnMount}
func cmdUnMount(args []string) int {
	thisName := args[0]

	ictl, err := imagectl.New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.SetReady(false); err != nil {
		return printChangeError(err)
	}

	fmt.Println("Image unmounted.")
	return 0
}

var CmdTag = Command{"tag", []string{"TAG", "NAME"}, "Create an alias for the image", cmdTag}
func cmdTag(args []string) int {
	tag := args[0]
	thisName := args[1]

	mctl, err := imagectl.NewMachineCtl()
	if err != nil {
		panic(err)
	}

	this, err := mctl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := imagectl.Tag(tag, &this); err != nil {
		panic(err)
	}

	fmt.Println("Tag created.")
	return 0
}
