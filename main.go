package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		PrintHelp(Commands)
		return
	}

	if args[0] == "-v" || args[0] == "--version" {
		fmt.Println("Siren version 0")
		return
	}

	var CommandMap = map[string]Command{}

	for _, cmd := range Commands {
		CommandMap[cmd.Name] = cmd
	}

	cmd, ok := CommandMap[args[0]]
	if !ok {
		PrintInvalidCommand(args[0])
		return
	}

	os.Exit(cmd.Run(args[1:]))
}
