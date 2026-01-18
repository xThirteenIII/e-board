package main

import (
	"bufio"
	"fmt"
	"minishell/shell"
	"os"
)

func main() {

	for {

		/* READ FROM STDIN */
		fmt.Printf("miniSh> ")
		cmdLine := shell.CommandLine{}
		scanner := bufio.NewScanner(os.Stdin)

		// Set the split function to scan Lines
		// By dafault max 1 line should be scanned each command.
		// ScanLines should do that by default.
		// TODO:check if this is true
		scanner.Split(bufio.ScanLines)

		// Print out the words
		for scanner.Scan() {
			cmdLine.Input = scanner.Text()
			if err := cmdLine.Eval(); err != nil {
				// Println: Spaces are always added between operands and a newline is appended.
				fmt.Println("miniSh:", err)
			}
			break
		}
	}

}
