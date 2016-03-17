package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LEW21/siren/imagectl"
)

var Commands = []imagectl.Command{CmdBuild}

var CmdBuild = imagectl.Command{nil, "build", []string{"PATH"}, nil, "Build an image from a Sirenfile", cmdBuild}
func cmdBuild(args []string) int {
	path := args[0]

	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	_, tag, ok := Build(path, os.Stderr)
	if !ok {
		return 1
	}

	fmt.Println()
	fmt.Println("Image created.")
	fmt.Println("Use 'siren create instance_name " + tag + "' to create a new, writable machine image using this image as a base.")
	return 0
}
