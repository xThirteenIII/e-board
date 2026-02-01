package shell

import "strings"

func parseCommand(s string) Command {
	s = strings.TrimSpace(s)

	// This will change in the future, for quotes handling
	return Command{Argv: strings.Fields(s)}
}
