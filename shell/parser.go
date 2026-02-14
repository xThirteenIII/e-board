package shell

import "strings"

// parseCommand trims white spaces from the given string
// Returns a Command with each whitespace separated arg.
func parseCommand(s string) Command {
	s = strings.TrimSpace(s)

	// This will change in the future, for quotes handling
	return Command{Argv: strings.Fields(s)}
}
