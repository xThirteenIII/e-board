package shell

// Operator tells the shell how to handle each command given before one.
type Operator int

const (
	OpNone = iota + 1
	OpBackground
	OpPipe
)

func (op Operator) IsValid() bool {
	return op >= OpNone && op <= OpPipe
}

func (op Operator) String() string {
	switch op {
	case OpBackground:
		return "&"
	case OpPipe:
		return "|"
	case OpNone:
		return ""
	default:
		return ""
	}
}

func parseOperator(r rune) Operator {
	switch r {
	case '&':
		return OpBackground
	case '|':
		return OpPipe
	default:
		return OpNone
	}
}

func isOperator(r rune) bool {
	op := parseOperator(r)
	if op.IsValid() {
		return op == OpBackground || op == OpPipe
	}
	return false
}
