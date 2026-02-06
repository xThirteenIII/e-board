package shell

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func isBuiltinCommand(s string) bool {
	switch s {
	case "pwd", "cd", "echo", "exit":
		return true
	default:
		return false
	}
}

func (cu CommandUnit) executeBuiltIn() error {

	command := cu.Cmd
	if len(command.Argv) == 0 {
		return fmt.Errorf("shouldn't get argv length 0")
	}

	if cu.OpAfter == OpBackground {
		fmt.Println("running in bg")
	}

	// Check program name
	switch command.Argv[0] {
	case "pwd":
		// NB: no newline needed at the end of it
		// TODO: this should be done later as: builtinProgram(args[0]) error
		// Fun fact, pwd does not care about args. 'pwd whatever you want' will still
		// print working directory.
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("pwd: %w", err)
		}
		fmt.Println(dir)
		return nil

	case "exit":
		// TODO: this should send a signal to close the shell
		// If more than 1 arg, return an error.
		if len(command.Argv) > 2 {
			return fmt.Errorf("exit: too many arguments")
		}

		if len(command.Argv) == 1 {
			// Default exit code is 0.
			os.Exit(0)
		}

		// convert exit code from string to int.
		// return an error if is not a number
		code, err := strconv.Atoi(command.Argv[1])
		if err != nil {
			return fmt.Errorf("exit: %s: numeric argument required", command.Argv[1])
		}

		os.Exit(code)

		return nil
	case "cd":

		// If more than 1 arg, return an error.
		if len(command.Argv) > 2 {
			return fmt.Errorf("cd: too many arguments")
		}

		// If arg is empty, cd to $HOME
		if len(command.Argv) == 1 {
			home := os.Getenv("HOME")
			err := os.Chdir(home)
			if err != nil {
				return fmt.Errorf("cd: %v", err)
			}
			return nil
		}
		err := os.Chdir(command.Argv[1])
		if err != nil {
			return fmt.Errorf("cd: %v", err)
		}
		return nil
	case "echo":
		if len(command.Argv) > 1 {
			fmt.Println(strings.Join(command.Argv[1:], " "))
		}

		return nil

		// NOT A BUILTIN COMMAND
	default:
		return nil
	}

}

// builtInCommands executes builtin command. Returns false if not a builtin command.
func (cu CommandUnit) executeExternal() error {
	// run built in commands
	progName := cu.Cmd.getProgramName()
	args := cu.Cmd.getArgs()

	fmt.Printf("executing %s:", progName)
	// Combination of fork and exec, careful to be thread safe.
	pid, err := syscall.ForkExec(
		progName,
		args,
		&syscall.ProcAttr{
			Files: []uintptr{os.Stdin.Fd(), os.Stderr.Fd(), os.Stdout.Fd()},
			// Set process group id
			Sys: &syscall.SysProcAttr{
				Setpgid: true,
			},
			Env: os.Environ(),
		})
	/*
		// ProcAttr holds attributes that will be applied to a new process started
		// by [StartProcess].
		type ProcAttr struct {
			Dir   string    // Current working directory.
			Env   []string  // Environment.
			Files []uintptr // File descriptors.
			Sys   *SysProcAttr
		}
	*/
	if err != nil {
		fmt.Printf("errno %d:", err)
		fmt.Printf("%s: command not found\n", progName)
	}

	fmt.Println("pid:", pid)

	// WARNING: funny bug, if i edit this file in my minishell using text editor and saving the result, i will get
	//			segmentation fault and invalid memory pointer

	return nil
}
