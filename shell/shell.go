package shell

import (
	"fmt"
	"strings"
)

type CommandLine struct {
	Input       string
	Argv        []string
	ProgramName string
	Background  bool
}

// Eval evaluates the command line.
// First it calls parseLine, then runs the BuiltInProgram() fn.
func (cmd *CommandLine) Eval() error {
	err := cmd.ParseLine()
	if err != nil {
		return err
	}
	err = BuiltInCommands(*cmd)
	if err != nil {
		return err
	}

	return nil
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

	cmd.ProgramName = args[0]

	// program arguments, exclude program name
	cmd.Argv = args[1:]
	// Exclude & if in bg
	if cmd.Background {
		// If single ampercend, remove it from Argv
		if args[len(args)-1] == "&" {
			cmd.Argv = cmd.Argv[:len(cmd.Argv)-1]
		} else {
			// otherwise just remove & from the arg string
			cmd.Argv[len(cmd.Argv)-1] = strings.Replace(cmd.Argv[len(cmd.Argv)-1], "&", "", 1)
		}
	}
	fmt.Println(cmd.Argv)
	return nil
}
