package shell

// Job is any program interactively started by the shell.
// A job is a unit of work. It can have multiple processes and has its own ID.
type Job struct {
	PGID      int
	LeaderPID int
	Commands  []CommandUnit
	//status
}
