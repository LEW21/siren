package imagectl

import (
	"fmt"
	"os"
)

var Commands = []Command{CmdCreate, CmdTag, CmdSetReadOnly, CmdSetReady, CmdRemove}

var CmdCreate = Command{"create", []string{"NAME", "BASE_NAME"}, "Create a new image", cmdCreate}
func cmdCreate(args []string) int {
	thisName := args[0]
	baseName := args[1]

	ictl, err := New()
	if err != nil {
		panic(err)
	}

	i, err := ictl.CreateImage(thisName, baseName)
	if err != nil {
		switch err {
			case ErrBaseDoesNotExist:
				fmt.Fprintln(os.Stderr, "Base image does not exist.")
				return 1
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

	target, err := UnTag(thisName)
	switch err {
		case nil:
			fmt.Println("Tag removed.")
			thisName = target

		case ErrNotATag:
			err = nil
			break

		default:
			panic(err)
	}

	ictl, err := New()
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

var BoolValues = map[string]bool {
	"1": true,
	"y": true,
	"yes": true,
	"true": true,
	"Y": true,
	"YES": true,
	"TRUE": true,
	"0": false,
	"n": false,
	"no": false,
	"false": false,
	"N": false,
	"NO": false,
	"FALSE": false,
}

var CmdSetReadOnly = Command{"set-read-only", []string{"NAME", "BOOL"}, "Mark or unmark image read-only", cmdSetReadOnly}
func cmdSetReadOnly(args []string) int {
	thisName := args[0]
	value, ok := BoolValues[args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid boolean value.")
		return 1
	}

	ictl, err := New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.SetReadOnly(value); err != nil {
		return printChangeError(err)
	}

	if this.ReadOnly() {
		fmt.Println("Image is now read-only.")
	} else {
		fmt.Println("Image is now writable.")
	}
	return 0
}

var CmdSetReady = Command{"set-ready", []string{"NAME", "BOOL"}, "Assemble or disassemble layered image", cmdSetReady}
func cmdSetReady(args []string) int {
	thisName := args[0]
	value, ok := BoolValues[args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, "Invalid boolean value.")
		return 1
	}

	ictl, err := New()
	if err != nil {
		panic(err)
	}

	this, err := ictl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := this.SetReady(value); err != nil {
		return printChangeError(err)
	}

	if this.Ready() {
		fmt.Println("Image is now ready.")
	} else {
		fmt.Println("Image is now not ready.")
	}
	return 0
}

var CmdTag = Command{"tag", []string{"TAG", "NAME"}, "Create an alias for the image", cmdTag}
func cmdTag(args []string) int {
	tag := args[0]
	thisName := args[1]

	mctl, err := NewMachineCtl()
	if err != nil {
		panic(err)
	}

	this, err := mctl.GetImage(thisName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Image does not exist.")
		return 1
	}

	if err := Tag(tag, &this); err != nil {
		panic(err)
	}

	fmt.Println("Tag created.")
	return 0
}
