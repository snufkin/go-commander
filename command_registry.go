package commander

import (
	"flag"
	"fmt"
	"path"
	"strings"

	"github.com/kardianos/osext"
)

// CommandRegistry will handle all CLI request
// and find the route to the proper Command
type CommandRegistry struct {
	Commands map[string]*CommandWrapper
	Helper   *CommandHelper
	Depth    int

	maximumCommandLength int
}

// Register is a function that adds your command into the registry
func (c *CommandRegistry) Register(f NewCommandFunc) {
	wrapper := f(c.executableName())
	name := wrapper.Help.Name
	c.Commands[name] = wrapper
	commandLength := len(fmt.Sprintf("%s %s", name, wrapper.Help.Arguments))
	if commandLength > c.maximumCommandLength {
		c.maximumCommandLength = commandLength
	}
}

// Execute finds the proper command, handle errors from the command and print Help
// if the given command it unknown or print the Command specific help
// if something went wrong or the user asked for it.
func (c *CommandRegistry) Execute() {
	name := flag.Arg(c.Depth)
	c.Helper = &CommandHelper{}
	c.Helper.Parse(flag.Args()[c.Depth:])
	if command, ok := c.Commands[name]; ok {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[E] %s\n\n", err)
				c.CommandHelp(name)
			}
		}()

		command.Handler.Execute(c.Helper)
	} else {
		c.Help()
	}
}

// Help lists all available commands to the user
func (c *CommandRegistry) Help() {
	if flag.Arg(c.Depth) == "help" && flag.Arg(c.Depth+1) != "" {
		c.CommandHelp(flag.Arg(c.Depth + 1))
		return
	}

	format := fmt.Sprintf("%%-%ds   %%s\n", c.maximumCommandLength)
	for name, command := range c.Commands {
		fmt.Printf(
			format,
			fmt.Sprintf("%s %s", name, command.Help.Arguments),
			command.Help.ShortDescription,
		)
	}
	fmt.Printf(
		format,
		"help [command]",
		"Display this help or a command specific help",
	)
}

// CommandHelp prints more detailed help for a specific Command
func (c *CommandRegistry) CommandHelp(name string) {
	if command, ok := c.Commands[name]; ok {
		extra := ""
		if c.Depth > 0 {
			extra = strings.Join(flag.Args()[0:c.Depth], " ")
		}
		if len(extra) > 0 {
			extra += " "
		}
		fmt.Printf("Usage: %s %s%s %s\n", c.executableName(), extra, name, command.Help.Arguments)

		if command.Help.LongDescription != "" {
			fmt.Println("")
			fmt.Println(command.Help.LongDescription)
		}

		if len(command.Help.Examples) > 0 {
			fmt.Printf("\nExamples:\n")
			for _, line := range command.Help.Examples {
				fmt.Printf("  %s %s%s %s\n", c.executableName(), extra, name, line)
			}
		}
	}
}

// Determine the name of the executable
func (c *CommandRegistry) executableName() string {
	filename, _ := osext.Executable()
	return path.Base(filename)
}

// NewCommandRegistry is a simple "constructor"-like function
// that initializes Commands map
func NewCommandRegistry() *CommandRegistry {
	flag.Parse()
	return &CommandRegistry{
		Commands: map[string]*CommandWrapper{},
	}
}
