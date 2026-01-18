package shell

import (
	"fmt"
	"strings"
)

type CommandLine struct {
	Input      string
	Argv       []string
	Background bool
}

// Eval evaluates the command line.
// First it calls parseLine, then runs the BuiltInProgram() fn.
func (cmd *CommandLine) Eval() error {
	return cmd.ParseLine()
}

// ParseLine parses the CommandLine Input, by space-separator.
// Checks if the command terminates with '&'.
// Sets the according Background flag to True if so, otherwise sets it to 0.
// It sets also the Argv slice to pass to excve function.
func (cmd *CommandLine) ParseLine() error {
	err := cmd.SplitArgs()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (cmd *CommandLine) SplitArgs() error {
	args := strings.Fields(cmd.Input)

	// if input is empty, just return
	if len(args) == 0 {
		return nil
	}

	if args[0] == "ciao" {
		// NB: no newline needed at the end of it
		return fmt.Errorf("%s: command not found", args[0])
	}
	fmt.Println(args)
	return nil
}
