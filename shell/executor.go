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

	miniSh := GetMiniShell()

	// Using exec.Command() to support windows OS.
	// exec.Cmd is a sweet spot to handle also process group id, FDs and other attributes if needed.
	cmd := exec.Command(progName, args...)
	// No child process created at this point.

	// By default, the child has the same process group as the parent.
	// If we want to change it we must use the setpgid(pid, pgid) function.
	// Setpgid sets the process group of `pid` to `pgid`.
	// setpgid(0, pgid), sets the PID of the current process in use to `pgid`.
	//		e.g. current PID = 15333, current PGID = 15333.
	//		setpgid(0, 15444) -> current PID = 15333, its PGID = 15444
	// setpgid(pid, 0), the pid of the process specified by	`pid` is used also for `pgid`.
	//		e.g.  current PGID = 15333.
	//		setpgid(15545, 0) -> group ID of process 15545 becomes current PGID = 15333
	// setpgid(0, 0), sets the current process ID as also the the group ID.
	//		e.g. current PID = 15333, current PGID = 15333.
	//		setpgid(0, 0) -> current PID = 15333, current PGID = 15333
	// We want to set a new process group id, different from the parent (shell pgid),
	// to avoid killing the shell process with exception SIGNALS.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// This simply tells that we want to set the group pid
		Setpgid: true,
		// Let's use the current process id (command child process) as the group id
		Pgid: 0,
		// Foreground is buggy. For some reason it doesn't handle file descriptors well, and
		// even using ctty stops the current job and the minishell exits.
	}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// WARNING: bug: when some programs like 'cat', who handle no args with passing to reading user input,
	// ctrl+c does not kill the process.

	// Start starts the specified command but does not wait for it to complete.
	//
	// If Start returns successfully, the c.Process field will be set.
	//
	// After a successful call to Start the [Cmd.Wait] method must be called in
	// order to release associated system resources.
	// If command has to run in background
	// Start the command process
	err := cmd.Start()
	pgidChild, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("couldn't get child group ID")
	}

	// TODO: handle multiple commands that belong to a single job

	// If the job is foreground
	if cu.OpAfter != OpBackground {

		miniSh.AddForegroundJob(Job{
			Pgid:     pgidChild,
			Status:   "Running",
			Commands: nil,
		})
		// Parent wait for job to terminate
		err = cmd.Wait()
		if errors.Is(err, &exec.ExitError{}) {
			return fmt.Errorf("error while completing program: %v\n", err)
		} else if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		// If the job has to run in the background
	} else {
		// Add to job table on
		// TODO: handle job statuses better
		miniShell.bgJobs = append(miniShell.bgJobs, Job{
			Pgid:     pgidChild,
			Status:   "Running",
			Commands: []CommandUnit{cu},
		})
		// Don't wait for job to finish
	}

	return nil
}
