package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func isBuiltinCommand(s string) bool {
	switch s {
	case "pwd", "cd", "echo", "exit", "builtin", "type":
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
	case "type":
		if len(command.Argv) > 1 {
			for _, t := range command.Argv[1:] {
				if isBuiltinCommand(t) {
					fmt.Printf("%s is a shell builtin\n", t)
				} else {
					bin, err := exec.LookPath(t)
					if err != nil {
						return fmt.Errorf("type: %v", err)
					}
					fmt.Printf("%s is %s\n", t, bin)
				}
			}
		}
		return nil
	case "builtin":
		fmt.Printf("builtin commands:\n- pwd\n- cd\n- echo\n- exit\n" +
			"- builtin\n- type\n")
		return nil

	default:
		return nil
	}

}

// builtInCommands executes builtin command. Returns false if not a builtin command.
func (cu CommandUnit) executeExternal() error {
	// run built in commands
	progName := cu.Cmd.getProgramName()
	args := cu.Cmd.getArgs()

	// Using exec.Command() to support windows OS.
	// exec.Cmd is a sweet spot to handle also process group id, FDs and other attributes if needed.
	cmd := exec.Command(progName, args...)
	// No child process created at this point.

	// Set FDs
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// WARNING: bug, if a program e.g. 'cat' is called with no argument, that hangs, ctrl+c doesnt kill the process.

	// Run creates a child process, forking and executing.
	// It returns c.Wait() error
	err := cmd.Run()
	if errors.Is(err, &exec.ExitError{}) {
		return fmt.Errorf("error while completing program: %v\n", err)
	} else if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	CurrJob = CreateJob(cmd.Process.Pid)
	// Print after error check to avoid invalid memory pointer.
	fmt.Printf("shell pid: %d\n", os.Getpid())
	fmt.Printf("shell gid: %d\n", os.Getgid())
	fmt.Printf("child pid: %d\n", cmd.Process.Pid)
	fmt.Printf("child pid w/ sate: %d\n", cmd.ProcessState.Pid())
	fmt.Printf("child gid: %d\n", cmd.SysProcAttr.Pgid)

	return nil
}
