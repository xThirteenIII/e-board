package shell

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"syscall"
)

type shell struct {
	fgJob  Job          // active foreground job
	bgJobs []Job        // list of jobs in the background
	pid    int          // shell process id
	pgid   int          // shell group process id
	mu     sync.RWMutex // mutex for synchronizing job addition/deletion
}

var miniShell *shell // miniShell global var to be accessed acrossed packages

func InitMiniShell() {
	miniShell = &shell{
		bgJobs: make([]Job, 0, 32),
		pid:    os.Getpid(),
		pgid:   syscall.Getpgrp(),
		mu:     sync.RWMutex{},
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
	miniShell.fgJob = Job{}
}

func AddBackgroundJob(j Job) error {
	return addBackgroundJob(j)
}

func addForegroundJob(j Job) {
	miniShell.fgJob = j
}

func addBackgroundJob(j Job) error {
	if len(miniShell.bgJobs) == 32 {
		return fmt.Errorf("Max job capacity reached")
	}
	miniShell.bgJobs = append(miniShell.bgJobs, j)
	return nil
}

func (ms *shell) GetBackgroundJobs() []Job {
	return miniShell.bgJobs
}

func (ms *shell) GetForegroundJob() Job {
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
func (ms *shell) reapZombies() {
	// waitstatus checks the status which the child terminated with
	// WNOHANG tells the wait4 call NOT TO BLOCK waiting for zombie children
	// This means that with each for iteration, even if there's no zombies,
	// the syscall will just return 0, it will not wait for some process to become a
	// zombie.
	// If we dindt use WNOHANG, the for{} cycle would block after just one terminated child
	var ws syscall.WaitStatus
	pid, _ := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
	for i, job := range ms.bgJobs {
		// Since we put pgid = pid when we created children,
		// they will hold the same value

		// Mark job done
		if job.Pgid == pid {
			if ws.Exited() || ws.Signaled() {
				ms.bgJobs[i].Status = "Done"
				ms.bgJobs[i].Finished = true
			}
		}
	}
}

func (ms *shell) ReapZombies() {
	ms.reapZombies()
}

func GetMiniShellPid() int {
	return miniShell.pid
}

func GetMiniShellPgid() int {
	return miniShell.pgid
}

// PrintDoneJobs prints all background jobs that have finished execution.
// Not handling "Terminated" status for now.
// It also reaps them from the slice.
func (ms *shell) PrintDoneJobs() {
	ms.mu.Lock()
	for i, job := range ms.bgJobs {
		if job.Status == "Done" && job.Finished {
			// TODO: handle + or - signs next to [n]
			fmt.Printf("[%d]  Done\t\t\t%s %s\n", i+1, job.Command.Cmd.Argv[0], strings.Join(job.Command.Cmd.Argv[1:], " "))
		}
	}
	// Delete job after printing
	ms.bgJobs = slices.DeleteFunc(ms.bgJobs, func(j Job) bool { return j.Status == "Done" && j.Finished })
	ms.mu.Unlock()
}
