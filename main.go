package main

import (
	"bufio"
	"fmt"
	"minishell/shell"
	"os"
	"strings"
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
	scanner := bufio.NewScanner(os.Stdin)

	for {

		// Print beautiful and original shell name.
		fmt.Printf("miniSh> ")

		if !scanner.Scan() {
			// EOF or error
			// EOF: Scan() returns false and scanner.Err() == nil
			break
		}

		// Does this really need to be created each iteration?
		// Yes, don't want any leftovers from previous commands.
		cmdLine := shell.CommandLine{}

		/*
		 *  Wrapping the unbuffered os.Stdin with a buffered scanner gives a convenient Scan method
		 *  that advances the scanner to the next token; which is the next line in the default scanner.
		 *  Production safe. Used by Github CLI.
		 */

		// We want to read just a line for the command.
		// Thus an if is sufficient, we don't need a for loop.
		// If the scanner has read succesfully user input up until '\n' then:

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
