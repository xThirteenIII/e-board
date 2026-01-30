package shell

import (
	"fmt"
	"os"
)

func BuiltInCommands(cmd CommandLine) error {
	// run built in commands
	switch cmd.ProgramName {
	case "pwd":
		// NB: no newline needed at the end of it
		// TODO: this should be done later as: builtinProgram(args[0]) error
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error while getting directory: %w", err)
		}
		fmt.Println(dir)
		return nil

	case "exit":
		// TODO: this should send a signal to close the shell
		fmt.Println("dont have SIGTERM yet :P")
		os.Exit(0)
		return nil
	// If not a builtin commands, print the info and return an error
	default:
		return fmt.Errorf("%s: command not found", cmd.ProgramName)
	}
}
