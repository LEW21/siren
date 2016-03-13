package imagectl

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
			fmt.Fprintf(os.Stderr, "%v: \"%v\" requires 1 argument.\n\n", os.Args[0], c.Name)
		} else {
			fmt.Fprintf(os.Stderr, "%v: \"%v\" requires %v arguments.\n\n", os.Args[0], c.Name, len(c.ReqArgs))
		}
		fmt.Fprintf(os.Stderr, "Usage: %v %v %v\n\n", os.Args[0], c.Name, strings.Join(c.ReqArgs, " "))
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

type CommandGroup struct {
	Name string
	Commands []Command
}

func PrintHelp(desc string, commandGroups []CommandGroup) {
	fmt.Printf("Usage: %v COMMAND [arg...]\n", os.Args[0])
	fmt.Printf("       %v [ -h | --help | -v | --version ]\n", os.Args[0])
	fmt.Println()
	fmt.Println(desc)

	for _, group := range commandGroups {
		fmt.Println()
		fmt.Println(group.Name + " Commands:")
		for _, c := range group.Commands {
			fmt.Printf("  %-27v %v\n", c.Name + " " + strings.Join(c.ReqArgs, " "), c.Description)
		}
	}
}

func PrintInvalidCommand(name string) {
	fmt.Fprintf(os.Stderr, "%v: \"%v\" is not a valid command.\n", os.Args[0], name)
	fmt.Fprintf(os.Stderr, "See '%v --help'.\n", os.Args[0])
}

func RunCommand(args []string, commandGroups []CommandGroup) int {
	var commandMap = map[string]Command{}

	for _, group := range commandGroups {
		for _, cmd := range group.Commands {
			commandMap[cmd.Name] = cmd
		}
	}

	cmd, ok := commandMap[args[0]]
	if !ok {
		PrintInvalidCommand(args[0])
		return 1
	}

	return cmd.Run(args[1:])
}
