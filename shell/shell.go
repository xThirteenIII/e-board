package shell

import (
	"fmt"
	"os"
)

type shell struct {
	fgJob  *Job  // active foreground job
	bgJobs []Job // list of jobs in the background
	pid    int   // shell process id
	pgid   int   // shell group process id
}

var miniShell *shell // miniShell global var to be accessed acrossed packages

func InitMiniShell() {
	miniShell = &shell{
		bgJobs: make([]Job, 0, 32),
		pid:    os.Getpid(),
		pgid:   os.Getgid(),
	}
}

func GetMiniShell() *shell {
	return miniShell
}

func (ms *shell) AddForegroundJob(j Job) {
	addForegroundJob(j)
}

func (ms *shell) RemoveForegroundJob() {
	removeFgJob()
}

func removeFgJob() {
	miniShell.fgJob = nil
}

func AddBackgroundJob(j Job) error {
	return addBackgroundJob(j)
}

func addForegroundJob(j Job) {
	miniShell.fgJob = &j
}

func addBackgroundJob(j Job) error {
	if len(miniShell.bgJobs) == 32 {
		return fmt.Errorf("Max job capacity reached")
	}
	miniShell.bgJobs = append(miniShell.bgJobs, j)
	return nil
}

func (ms shell) GetBackgroundJobs() []Job {
	return miniShell.bgJobs
}

func (ms shell) GetForegroundJob() *Job {
	return miniShell.fgJob
}

// GetUniqueFgPgids returns a slice of unique pgids, from the foreground Jobs slice.
/*
func GetUniqueFgPgids() []int {

	// If there is no foreground jobs, return nil.
	if len(miniShell.fgJobs) == 0 {
		return nil
	}

	// pgidsMap holds unique pgid
	pgidsMap := make(map[int]int)

	// Go through fgJobs slice.
	// Add each foreground job pgid as a key to the map
	// If key already exists, it will just update the value,
	// we don't care about that
	for i := 0; i < len(miniShell.fgJobs); i++ {
		pgidsMap[miniShell.fgJobs[i].Pgid]++
	}

	pgids := make([]int, 0, len(pgidsMap))
	fmt.Println("\npgid map:", pgidsMap)
	for k, _ := range pgidsMap {
		pgids = append(pgids, k)
	}

	return pgids

}
*/

func GetMiniShellPid() int {
	return miniShell.pid
}

func GetMiniShellPgid() int {
	return miniShell.pgid
}
