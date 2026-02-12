package shell

// Job is any program interactively started by the shell.
// A job is a unit of work. It can have multiple processes and has its own ID.
// Added to jobs it's only background jobs, since foreground commands wait() for the
/*
	Unix shells use the abstraction of a job to represent the processes that are created
	as a result of evaluating a single command line. At any point in time, there is at
	most one foreground job and zero or more background jobs. For example, typing
	linux> ls | sort
	creates a foreground job consisting of two processes connected by a Unix pipe: one
	running the ls program, the other running the sort program. The shell creates
	a separate process group for each job. Typically, the process group ID is taken
	from one of the parent processes in the job
*/
type Job struct {
	PGID      int // list of process group ID
	LeaderPID int
	Commands  []CommandUnit
}

func CreateJob(pid int) *Job {
	return &Job{
		PGID: pid,
	}
}

var CurrJob *Job
