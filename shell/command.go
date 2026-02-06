package shell

import "fmt"

// CommandLine is used to handle full user input and accessing each command.
// Each command-line is seen as:
// * cmd1 operator cmd2 operator cmd3 ...
type CommandLine struct {
	commands  []CommandUnit // List of '&' separated commands
	UserInput string        // Unmodified user input
}

// A CommandUnit is made of a Command and the operator (&, ;, |) after the command.
type CommandUnit struct {
	Cmd     Command
	OpAfter Operator
}

// A command is an executable unit, parsed form CommandLine.
// It is composed by the argv vector: argv[0] being programName and argv[1:] being the arguments, if any.
// Each command is a process. (builtin or execv)
type Command struct {
	Argv []string
}

func (c Command) getProgramName() string {
	if len(c.Argv) > 0 {
		return c.Argv[0]
	}
	return ""
}

func (c Command) getArgs() []string {
	if len(c.Argv) > 1 {
		return c.Argv[1:]
	}
	return nil
}

// Eval evaluates the command line.
// First it calls parseLine, then runs the BuiltInProgram() fn.
func (cl *CommandLine) Eval() error {
	// Ignore empty lines.
	if cl.UserInput == "" {
		return nil
	}

	// Get and assign each command unit. I.e. cmd + operator.
	// Evaluate operators.
	cl.getCommandUnits()
	err := cl.checkCommandSyntax()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// Run each command in the commands unit.
	for _, cu := range cl.commands {

		// Given that there is at least one argument
		if len(cu.Cmd.Argv) > 0 {

			// Execute builtin program
			if isBuiltinCommand(cu.Cmd.Argv[0]) {
				err := cu.executeBuiltIn()
				if err != nil {
					fmt.Println(err)
				}
				// Or run external program (OS)
			} else {
				err := cu.executeExternal()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}

// getCommandUnits reads the line from left to right.
func (cl *CommandLine) getCommandUnits() {

	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		start int
		end   int
		op    Operator
	}
	spans := make([]span, 0, 32)

	// Find the command unit start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	start := -1 // valid span start if >= 0

	// What is happening in this seemingly messed up cycle?
	// start is -1.
	// First iteration start is set to end (=0).
	// When we find an operator, append the current span[start, end, operator].
	// start is set to negative again.
	// Next iteration, start = end, i.e. the index after the operator, the start of our next span.
	for end, rune := range cl.UserInput {
		if isOperator(rune) {
			if start >= 0 {
				spans = append(spans, span{start, end, parseOperator(rune)})
				// Set start to a negative value.
				// Note: using -1 here consistently and reproducibly
				// slows down this code by a several percent on amd64.
				// Invert start. i.e. start = 4 = 0b0100; ^start = 0b1011 = -5
				start = ^start
			}
		} else {
			if start < 0 {
				start = end
			}
		}
	}

	// Last field might end at EOF.
	// Last field has always OpNone
	// Last char whitespace is handled already in main package. User input is trimmed.
	if start >= 0 {
		spans = append(spans, span{start, len(cl.UserInput), OpNone})
	}

	// Create strings from recorded field indices.
	cl.commands = make([]CommandUnit, len(spans))
	for i, span := range spans {
		// WARNING: does this modify UserInput?
		cl.commands[i] = CommandUnit{Cmd: parseCommand(cl.UserInput[span.start:span.end]), OpAfter: span.op}
	}
}

// getCommandUnits reads the line from left to right.
func (cl *CommandLine) Commands() []CommandUnit {

	return cl.commands
}

// getCommandUnits reads the line from left to right.
func (cl CommandLine) checkCommandSyntax() error {

	// If a syntax error is found return an error.
	for _, cu := range cl.commands {
		if len(cu.Cmd.Argv) == 0 && cu.OpAfter == OpBackground {
			return fmt.Errorf("syntax error near unexpected token '&'")
		}
	}
	//TODO: handle singleton &

	return nil
}
