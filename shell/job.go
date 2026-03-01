package shell

// Job is any program interactively started by the shell.
// A job is a unit of work. It can have multiple processes and has its own ID.
// Added to jobs it's only background jobs, since foreground commands wait() for the
// child to be done, even though this behaviour is theoretically incorrect.
/*
	Unix shells use the abstraction of a job to represent the processes that are created
	as a result of evaluating a single command line. At any point in time, there is at
	most one foreground job and zero or more background jobs. For example, typing
	linux> ls | sort
	creates a foreground job consisting of two processes connected by a Unix pipe: one
	running the ls program, the other running the sort program. The shell creates
	a separate process group for each job. Typically, the process group ID is taken
	from one of the parent processes in the job.
	We avoid this heredity since parent process is the mini shell itself and killing its
	group ID would mean shutting the program.
*/
type Job struct {
	Pgid     int // list of process group ID
	JID      int
	Status   string
	Command  CommandUnit
	Finished bool
}
