package main

import (
	"fmt"
	"os"
	"strings"
)

type Command struct {
	Name        string
	ReqArgs     []string
	Description string
	Executor    func(args []string) int
}

func (c Command) CheckArgs(args []string) bool {
	if len(args) < len(c.ReqArgs) {
		if len(c.ReqArgs) == 1 {
			fmt.Fprintf(os.Stderr, "siren: \"%v\" requires 1 argument.\n\n", c.Name)
		} else {
			fmt.Fprintf(os.Stderr, "siren: \"%v\" requires %v arguments.\n\n", c.Name, len(c.ReqArgs))
		}
		fmt.Fprint(os.Stderr, "Usage:  siren " + c.Name + " " + strings.Join(c.ReqArgs, " ") + "\n\n")
		fmt.Fprint(os.Stderr, c.Description + "\n")
		return false
	}
	return true
}

func (c Command) Run(args []string) int {
	if !c.CheckArgs(args) {
		return 1
	}

	return c.Executor(args)
}

func PrintHelp(commands []Command) {
	fmt.Println("Usage: siren [OPTIONS] COMMAND [arg...]")
	fmt.Println("       siren [ -h | --help | -v | --version ]")
	fmt.Println()
	fmt.Println("A systemd-nspawn container builder.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  none yet")
	fmt.Println()
	fmt.Println("Commands:")

	for _, c := range commands {
		fmt.Printf("    %-9v %v\n", c.Name, c.Description)
	}
}

func PrintInvalidCommand(name string) {
	fmt.Fprintf(os.Stderr, "siren: \"%v\" is not a siren command.\n", name)
	fmt.Fprintln(os.Stderr, "See 'siren --help'.")
}
