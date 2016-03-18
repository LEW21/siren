package imagectl

import (
	"fmt"
	"os"
	"strings"
)

type Command struct {
	Shortcuts   []string
	Name        string
	ReqArgs     []string
	OptArgs     []string
	Description string
	Executor    func(args []string) int
}

func (c Command) ArgsDescription() string {
	desc := strings.Join(c.ReqArgs, " ")
	if len(c.OptArgs) > 0 {
		desc = desc + " [" + strings.Join(c.OptArgs, "] [") + "]"
	}
	return desc
}

func (c Command) CheckArgs(args []string) bool {
	at_least := len(c.ReqArgs)
	at_most := len(c.ReqArgs) + len(c.OptArgs)

	last_arg := ""
	if len(c.OptArgs) > 0 {
		last_arg = c.OptArgs[len(c.OptArgs)-1]
	} else if len(c.ReqArgs) > 0 {
		last_arg = c.ReqArgs[len(c.ReqArgs)-1]
	}

	if len(args) < at_least {
		if at_least == 1 {
			fmt.Fprintf(os.Stderr, "%v: \"%v\" requires at least 1 argument.\n\n", os.Args[0], c.Name)
		} else {
			fmt.Fprintf(os.Stderr, "%v: \"%v\" requires at least %v arguments.\n\n", os.Args[0], c.Name, at_least)
		}
		fmt.Fprintf(os.Stderr, "Usage: %v %v %v\n\n", os.Args[0], c.Name, c.ArgsDescription())
		fmt.Fprint(os.Stderr, c.Description + "\n")
		return false
	}
	if len(args) > at_most && !strings.HasSuffix(last_arg, "...") {
		if at_most == 1 {
			fmt.Fprintf(os.Stderr, "%v: \"%v\" takes at most 1 argument.\n\n", os.Args[0], c.Name)
		} else {
			fmt.Fprintf(os.Stderr, "%v: \"%v\" takes at most %v arguments.\n\n", os.Args[0], c.Name, at_most)
		}
		fmt.Fprintf(os.Stderr, "Usage: %v %v %v\n\n", os.Args[0], c.Name, c.ArgsDescription())
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
			shortcut := ""
			if len(c.Shortcuts) > 0 {
				shortcut = c.Shortcuts[0] + ","
			}
			fmt.Printf("  %5v %-27v %v\n", shortcut, c.Name + " " + c.ArgsDescription(), c.Description)
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
			for _, shortcut := range cmd.Shortcuts {
				commandMap[shortcut] = cmd
			}
		}
	}

	cmd, ok := commandMap[args[0]]
	if !ok {
		PrintInvalidCommand(args[0])
		return 1
	}

	return cmd.Run(args[1:])
}
