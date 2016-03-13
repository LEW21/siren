package main

import (
	"fmt"

	"github.com/LEW21/siren/imagectl"
)

var Commands = []imagectl.Command{CmdBuild}

var CmdBuild = imagectl.Command{nil, "build", []string{"PATH"}, nil, "Build an image from a Sirenfile", cmdBuild}
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
