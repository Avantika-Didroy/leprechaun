package workers

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kilgaloon/leprechaun/context"
	"github.com/kilgaloon/leprechaun/log"
)

// AsyncMarker is string in step that we use to know
// does that command need to be done async
const AsyncMarker = "->"

// Worker is single worker and information about it
// worker is processing all steps
type Worker struct {
	StartedAt      time.Time
	WorkingOn      string
	Steps          []string
	Context        *context.Context
	Logs           log.Logs
	DoneChan       chan string
	ErrorChan      chan *Worker
	Name           string
	TasksPerformed int
	Cmd            map[string]*exec.Cmd
	Err            error
	Stdout         *os.File
}

// Run starts worker
func (w *Worker) Run(steps []string) {
	w.Steps = steps
	w.StartedAt = time.Now()

	for _, step := range w.Steps {
		if (w.Err != nil) {
			// Worker had some kind of error, don't run any steps
			return
		}
		w.Logs.Info("Step %s is in progress... \n", step)
		// replace variables
		parts := strings.Fields(step)

		if parts[0] == AsyncMarker {
			step = w.Context.Transpile(strings.Join(parts[1:], " "))
			go w.workOnStep(step)
		} else {
			step = w.Context.Transpile(step)
			w.workOnStep(step)
		}
	}
}

func (w *Worker) workOnStep(step string) {
	cmd := exec.Command("bash", "-c", step)
	w.Cmd[step] = cmd

	var stderr bytes.Buffer
	cmd.Stdout = w.Stdout
	cmd.Stderr = &stderr

	w.WorkingOn = step
	w.Err = cmd.Run()
	if w.Err != nil {
		w.ErrorChan <- w
		return
	}

	w.Logs.Info("Step %s finished... \n\n", step)
	// there is output, write it to info
	// if len(out.String()) > 0 {
	// 	out.WriteTo(w.Stdout)
	// }
	// command finished executing
	// delete it, and let it rest in pepperonies
	delete(w.Cmd, step)
	w.Done()
}

// Kill all commands that worker is working on
func (w *Worker) Kill() {
	for step, cmd := range w.Cmd {
		if err := cmd.Process.Kill(); err != nil {
			w.Logs.Error("Failed to kill process on step %s: %s", step, err)
		}
	}

	w.DoneChan <- w.Name
}

// Done signals that this worker is done and send his id for cleaner
func (w *Worker) Done() {
	w.TasksPerformed++
	// worker performed all tasks, and can be done
	if w.TasksPerformed == len(w.Steps) {
		w.DoneChan <- w.Name
	}
}
