package shell

type Shell struct {
	Jobs      []Job
	Terminal  int //tty fd
	ShellPGID int //shell process group
}
