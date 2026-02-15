package main

import (
	"bufio"
	"fmt"
	"minishell/shell"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

	// Main routine loop.
	// Each iteration:
	// * prints the prompt
	// * scanner.Scan() calls os.Stdin.Read() and blocks the loop
	// * it waits for user input and terminating \n character
	// * Scan() returns true (token found)
	// * reads the command and evaluates it
	// * eval executes the command
	// * if an error when reading the command occures, the program exits

	// Create a new pointer to a Scanner struct.
	/*
	 *  Wrapping the unbuffered os.Stdin with a buffered scanner gives a convenient Scan method
	 *  that advances the scanner to the next token; which is the next line in the default scanner.
	 *  Production safe. Used by Github CLI.
	 */
	scanner := bufio.NewScanner(os.Stdin)

	/*
		At any point in time, there can be at most one pending signal of a particular type.
		If a process has a pending signal of type k, then any subsequent signals of type
		k sent to that process are not queued; they are simply discarded.
		This is why sigCh is buffered with capacity 1.
	*/
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh)
	// Init minishell struct
	// Init fgJobs and bgJobs to slices of capacity 32. At most 32 jobs

	// Global variable to handle minishell.
	// TODO: check if there's better ways to do it.
	shell.InitMiniShell()
	miniSh := shell.GetMiniShell()
	for {

		// Handle Ctrl+C
		/*
			Typing Ctrl+C at the keyboard causes the main loop to send a SIGINT signal to
			every process in the foreground process group. In the default case, the result is to
			terminate the foreground job. Similarly, typing Ctrl+Z causes the shell loop to send a
			SIGTSTP signal to every process in the foreground process group. In the default
			case, the result is to stop (suspend) the foreground job.
		*/
		go func() {
			for sig := range sigCh {
				switch sig {
				case syscall.SIGINT:
					// If there's foreground jobs, send a signal to every fg group
					fgJob := miniSh.GetForegroundJob()
					err := syscall.Kill(-fgJob.Pgid, syscall.SIGKILL)
					if err != nil {
						fmt.Printf("\nerror killing processes in pgid:  %d: %v", fgJob.Pgid, err)
					}

					_, err = os.Stdout.WriteString("\nminiSh>")
					if err != nil {

						// Don't wait for scanner.Scan, continue to next iteration
						continue
					}
				default:
				}
			}
		}()

		// Print beautiful and original shell name.
		fmt.Printf("miniSh> ")

		// We want to read just a line for the command.
		// Thus an if is sufficient, we don't need a for loop.
		// This blocks until EOF (\n)
		if !scanner.Scan() {
			// EOF or error
			// EOF: Scan() returns false and scanner.Err() == nil
			break
		}

		// If the scanner has read succesfully user input up until '\n' then:
		// Does this really need to be created each iteration?
		// Yes, don't want any leftovers from previous commands.
		cmdLine := shell.CommandLine{}

		// Save Input as string
		// Text returns the current token, here the user command, from the input.
		cmdLine.UserInput = scanner.Text()

		// Remove leading and trailing whitespaces.
		// Needed to handle "command & " cases.
		cmdLine.UserInput = strings.TrimSpace(cmdLine.UserInput)

		/* EVALUATE COMMAND */
		// Check for errors while Evaluating the command, then print it out.
		// These errors do not terminate the shell.
		if err := cmdLine.Eval(); err != nil {
			fmt.Println("miniSh:", err)
		}

		// Check for errors during Scan. End of file is expected and not reported by Scan as an error.
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

	}
}
