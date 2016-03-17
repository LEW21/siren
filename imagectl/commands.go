package imagectl

import (
	"fmt"
	"os"
	"strconv"
	"github.com/fatih/color"
)

var Commands = []Command{CmdCreate, CmdTag, CmdSetReadOnly, CmdSetReady, CmdRemove, CmdList}

var CmdCreate = Command{[]string{"new"}, "create", []string{"NAME"}, []string{"BASE_NAME"}, "Create a new image", cmdCreate}
func cmdCreate(args []string) int {
	thisName := args[0]
	baseName := ""
	if len(args) >= 2 {
		baseName = args[1]
	}

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

var CmdRemove = Command{[]string{"rm"}, "remove", []string{"NAME"}, nil, "Remove an image", cmdRemove}
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

var CmdSetReadOnly = Command{[]string{"ro", "read-only"}, "set-read-only", []string{"NAME"}, []string{"BOOL"}, "Mark or unmark image read-only", cmdSetReadOnly}
func cmdSetReadOnly(args []string) int {
	thisName := args[0]
	svalue := "y"
	if len(args) >= 2 {
		svalue = args[1]
	}

	value, ok := BoolValues[svalue]
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

var CmdSetReady = Command{nil, "set-ready", []string{"NAME"}, []string{"BOOL"}, "Assemble or disassemble layered image", cmdSetReady}
func cmdSetReady(args []string) int {
	thisName := args[0]
	svalue := "y"
	if len(args) >= 2 {
		svalue = args[1]
	}

	value, ok := BoolValues[svalue]
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

var CmdTag = Command{nil, "tag", []string{"TAG", "NAME"}, nil, "Create an alias for the image", cmdTag}
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

// machinectl list-images / docker images
var CmdList = Command{[]string{"ls", "list-images", "images"}, "list", nil, nil, "Show available container and VM images", cmdList}
func cmdList(args []string) int {
	ictl, err := New()
	if err != nil {
		panic(err)
	}

	images, err := ictl.ListImages()
	if err != nil {
		panic(err)
	}

	columns := []Column{
		{"NAME",  func(i interface{})(string, color.Attribute){return i.(Image).Name(), 0}},
		{"TYPE",  func(i interface{})(string, color.Attribute){return i.(Image).Type(), 0}},
		{"RO",    func(i interface{})(string, color.Attribute){
			if i.(Image).ReadOnly() {
				return "RO", color.FgGreen
			} else {
				return "RW", color.FgBlue
			}
		}},
		{"READY", func(i interface{})(string, color.Attribute){
			if i.(Image).Ready() {
				return "yes", color.FgGreen
			} else {
				return "no", color.FgRed
			}
		}},
		{"ALIVE", func(i interface{})(string, color.Attribute){
			if i.(Image).Alive() {
				return "yes", color.FgGreen
			} else if !i.(Image).ReadOnly() { // RW images are usually supposed to be alive.
				return "no", color.FgRed
			} else {
				return "no", color.FgBlue
			}
		}},
	}
	data := make([]interface{}, len(images))
	for i := range data {
		data[i] = images[i]
	}
	printTable(os.Stdout, columns, data)

	fmt.Println()
	fmt.Println(strconv.Itoa(len(images)) + " images listed.")

	return 0
}
