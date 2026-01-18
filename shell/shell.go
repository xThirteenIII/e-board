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
		return fmt.Errorf("%w", err)
	}
	if cmd.Background {
		fmt.Println("vado in bg siuum")
	} else {
		fmt.Println("sono in fg")
	}
	return nil
}

// SplitArgs splits the command line into space-separated words.
// It returns an error if the command is not built-in or the executable is not found.
// It sets the Background flag to True if the command terminates with a "&".
func (cmd *CommandLine) SplitArgs() error {
	if strings.HasSuffix(cmd.Input, "&") {
		cmd.Background = true
	}
	args := strings.Fields(cmd.Input)

	// if input is empty, just return
	if len(args) == 0 {
		return nil
	}

	if args[0] == "ciao" {
		// NB: no newline needed at the end of it
		// TODO: this should be done later as: builtinProgram(args[0]) error
		return fmt.Errorf("%s: command not found", args[0])
	}

	if args[0] == "exit" {
		// TODO: this should send a signal to close the shell
		fmt.Println("dont have SIGTERM yet :P")
		return nil
	}

	fmt.Println(args)
	return nil
}
