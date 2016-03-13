package main

import (
	"fmt"
	"os"

	"github.com/LEW21/siren/imagectl"
)

func main() {
	args := os.Args[1:]

	allCommands := []imagectl.CommandGroup{
		{"Image", imagectl.Commands},
	}

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		imagectl.PrintHelp("Image manager for systemd-machined.", allCommands)
		return
	}

	if args[0] == "-v" || args[0] == "--version" {
		fmt.Println("imagectl version 0")
		return
	}

	os.Exit(imagectl.RunCommand(args, allCommands))
}
