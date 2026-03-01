package shell

import (
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

	// For now, I don't care about background jobs for builtin programs.
	// I'll just execute them and that's it.

	// if cu.OpAfter == OpBackground {
	// }

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
	progName := cu.Cmd.getProgramName()
	args := cu.Cmd.getArgs()

	miniSh := GetMiniShell()

	// Using exec.Command() to support windows OS.
	// exec.Cmd is a sweet spot to handle also process group id, FDs and other attributes if needed.
	cmd := exec.Command(progName, args...)
	// No child process created at this point.

	// By default, the child has the same process group as the parent.
	// If we want to change it we must use the setpgid(pid, pgid) function.
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

	//WARNING:  BUG when using nvim or other editors or commands that need a true tty to perform.
	//          Killing the group kills also the shell, might be that the process is assigned
	//          shell group process id automatically. Gotta investigate!

	// Handle race conditions with a mutex.
	// This is to avoid that the SIGCHLD handler tries to delete a job from the
	// background jobs slice, without the parent having added the job yet.
	// E.g.: ls &
	// child command ls terminates and SIGCHLD handler calls deletejob()
	// before the append to background jobs happens.
	// We might run the code a billion times without a problem, but then trigger a race
	// the next one.
	// This would be true here if it the handler in the switch case in the main.go file was
	// responsible for deleting jobs from the slice. But it's actually the PrintJobDone method
	// at the end of the main for{} loop. I actually don't know for now whether it's correct or not,
	// but it works.

	// Lock bgJob
	miniSh.mu.Lock()
	defer miniSh.mu.Unlock()

	// Start starts the specified command but does not wait for it to complete.
	//
	// If Start returns successfully, the c.Process field will be set.
	//
	// After a successful call to Start the [Cmd.Wait] method must be called in
	// order to release associated system resources.
	// If command has to run in background
	// Start the command process
	// Start does fork + execve
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("%s: command not found", progName)
	}
	pgidChild, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("couldn't get child process group ID")
	}
	// TODO: check sleep command behaviour, which actually listens to input while sleeping and puts the
	// command in queue

	// If the job is foreground
	if cu.OpAfter != OpBackground {

		miniSh.AddForegroundJob(Job{
			Pgid:     pgidChild,
			Status:   "Running",
			Finished: false,
		})
		// Parent waits for job to terminate
		// NOTE: ALL SYS resources are freed by Wait
		err = cmd.Wait()
		// If the job has to run in the background
	} else {
		// Add to background job table
		// TODO: handle job statuses better
		miniShell.bgJobs = append(miniShell.bgJobs, Job{
			Pgid:     pgidChild,
			Status:   "Running",
			Command:  cu,
			Finished: false,
		})
		fmt.Printf("[%d] %d\n", len(miniShell.bgJobs), miniShell.bgJobs[len(miniSh.bgJobs)-1].Pgid)
	}

	return nil
}
