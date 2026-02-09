package shell

// Job is any program interactively started by the shell.
// A job is a unit of work. It can have multiple processes and has its own ID.
type Job struct {
	PGID      int // list of process IDs still running
	LeaderPID int
	Commands  []CommandUnit
}

func CreateJob(pid int) *Job {
	return &Job{
		PGID: pid,
	}
}

var CurrJob *Job
