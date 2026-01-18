package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	for {

		/* READ FROM STDIN */
		fmt.Printf("miniSh> ")
		scanner := bufio.NewScanner(os.Stdin)

		// Set the split function to scan Lines
		// By dafault max 1 line should be scanned each command.
		// ScanLines should do that by default.
		// TODO:check if this is true
		scanner.Split(bufio.ScanLines)

		// Print out the words
		for scanner.Scan() {
			if scanner.Text() == "exit" {
				return
			}
			fmt.Println(scanner.Text())
			/* EVALUATE COMMAND */
			break
		}
	}

}
