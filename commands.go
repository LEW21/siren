package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LEW21/siren/imagectl"
)

var Commands = []imagectl.Command{CmdBuild, CmdPull}

var CmdBuild = imagectl.Command{nil, "build", []string{"DIR_PATH"}, []string{"TAG"}, "Build an image from a Sirenfile", cmdBuild}
func cmdBuild(args []string) int {
	path := args[0]
	tag := ""
	if len(args) >= 2 {
		tag = args[1]
	}

	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	_, tag, ok := Build(path, tag, os.Stderr)
	if !ok {
		return 1
	}

	fmt.Println()
	fmt.Println("Image created.")
	fmt.Println("Use 'siren create instance_name " + tag + "' to create a new, writable machine image using this image as a base.")
	return 0
}

var CmdPull = imagectl.Command{nil, "pull", []string{"URI"}, []string{"TAG"}, "Pull and build an image from a git repostory", cmdPull}
func cmdPull(args []string) int {
	uri := args[0]
	tag := ""
	if len(args) >= 2 {
		tag = args[1]
	}

	_, tag, ok := Pull(uri, tag, os.Stderr)
	if !ok {
		return 1
	}

	fmt.Println()
	fmt.Println("Image created.")
	fmt.Println("Use 'siren create instance_name " + tag + "' to create a new, writable machine image using this image as a base.")
	return 0
}
