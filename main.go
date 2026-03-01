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

// NOTE: cat penso trasferisca poi a stdin, per questo rimane stuck. SIGCHLD viene mandato appena
// dopo cat viene eseguito. I comandi in input durante lo "stuck time" sono bufferati e poi eseguiti,
// quando si killa cat, che viene killato solo con sudo kill -KILL

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
	lineCh := make(chan string, 1) // Buffered channel for string input
	defer close(lineCh)
	stopCh := make(chan struct{}) // Stop signal

	/*
		At any point in time, there can be at most one pending signal of a particular type.
		If a process has a pending signal of type k, then any subsequent signals of type
		k sent to that process are not queued; they are simply discarded.
		This is why sigCh is buffered with capacity 1.
	*/
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh)

	// Redirect all incoming signals to sigCh channel.
	// Init minishell struct

	// Global variable to handle minishell.
	// TODO: check if there's better ways to do it.
	shell.InitMiniShell()
	miniSh := shell.GetMiniShell()

	// Handle Ctrl+C
	/*
		Typing Ctrl+C at the keyboard causes the main loop to send a SIGINT signal to
		every process in the foreground process group. In the default case, the result is to
		terminate the foreground job. Similarly, typing Ctrl+Z causes the shell loop to send a
		SIGTSTP signal to every process in the foreground process group. In the default
		case, the result is to stop (suspend) the foreground job.
	*/
	// This runs cuncurrently with main loop
	// Keep signal handling as simple as possible
	// Print beautiful and original shell name.
	//NOTE: if this signal handling is put AFTER the for{} main loop, it does not process children!
	// This is caused by Go runtime scheduler that does not guarantee immediate scheduling.
	// So even if yes, concurrency is delivered, scheduling order is not. Race might happen (and it happens)
	go func() {

		// Signal handling
		// NOTE: pending queue is only buffered at 1 signal per type, that's default UNIX behaviour.
		//		 This means kernel sends at most 1 SIGNAL per type.
		//		 The parent has a pending bitmask like [SIGINT = 0, SIGCHLD = 1, SIG...]
		//		 If a bit is set, a signal of the corresponding type is pending for delivery.
		//	     If more signals of that type happen to be sent, they're discarded.
		//		 E.g.: ls & sleep 1  & sleep 1
		//		 ls terminates execution: Kernel sets SIGCHLD = 1, pending SIGCHLD
		//		 sleep terminates after 1s: SIGCHLD already set, coalesced ("accorpato in Italian"), no extra signal
		//		 sleep2 terminates after 1s: same as above
		//		 NOTE THAT this behaviour happens when we're talking about procesess finishing at the same time
		//		 that is <1microsecond, but that is often the case with normal commands
		for sig := range sigCh {
			switch sig {
			// SIGINT handler
			case syscall.SIGINT:
				// NOTE: the use of fmt.Printf/ln does not guarantee synchronization with main loop.
				// Prompt could be printed before!
				fgJob := miniSh.GetForegroundJob()
				// Spread SIGINT to every process belonging to pgid group
				if fgJob.Status != "" {
					err := syscall.Kill(-fgJob.Pgid, syscall.SIGINT)
					if err != nil {
						fmt.Printf("\nerror killing processes in pgid:  %d: %v", fgJob.Pgid, err)
					}
					miniSh.RemoveForegroundJob()
				}
				fmt.Printf("\n")
				// Stop from scanning
				stopCh <- struct{}{}
				// NOTE: after a process is terminated by SIGINT, a SIGCHLD signal is sent by the kernel
				// Printf is said not to be async-signal-safe, in CSAPP 8.5.5, but we don't care atm.
			// SIGCHLD handler
			case syscall.SIGCHLD:
				// The while true is necessary, since we want to reap every zombie process,
				// because remember that signals cannot be used to count the occurrence of events in other processes!!!
				// This means if two or more processes end in the same span of time (~microseconds), only 1 SIGCHLD is received
				/*
					sleep 1 & sleep 1 & sleep 1 &  # 3 zombies
					SIGCHLD → handler
					Wait4 → reap sleep1
					handler ends → sleep2/sleep3 will be zombies forever until miniSh is exited
				*/
				//fmt.Println("sto a ricevere sigchld")

				for {
					if len(miniSh.GetBackgroundJobs()) == 0 {
						break
					}
					miniSh.ReapZombies()
				}
			}
		}
	}()

	// Scanner  go routine
	go func() {
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}
	}()

	for {
		fmt.Printf("miniSh> ")
		select {
		case line := <-lineCh:

			// We want to read just a line for the command.
			// Thus an if is sufficient, we don't need a for loop.
			// This blocks until EOF (\n)
			// Block until line EOF, \n, or channel is closed
			// If the scanner has read succesfully user input up until '\n' then:
			// Does this really need to be created each iteration?
			// Yes, don't want any leftovers from previous commands.
			cmdLine := shell.CommandLine{}

			// Save Input as string
			// Text returns the current token, here the user command, from the input.
			cmdLine.UserInput = line

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
			// Print done jobs
			// Listen for stop signal
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		case <-stopCh:
			fmt.Println("che succ?")
			continue
		}
		miniSh.PrintDoneJobs()
	}
}
