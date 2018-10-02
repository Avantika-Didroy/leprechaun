package api

import (
	"errors"

	"github.com/kilgaloon/leprechaun/agent"
)

// Command is closure that will be called to execute command
type Command func(args ...string) ([][]string, error)
type tabel [][]string
type column []string

// Registrator is agent that will be registered
// with regi
type Registrator struct {
	Agent    agent.Agent
	Commands map[string]Command
}

// Command Set command to registrator
func (r *Registrator) Command(name string, command Command) {
	r.Commands[name] = command
}

// Call specified command
func (r Registrator) Call(name string, args ...string) ([][]string, error) {
	if command, exist := r.Commands[name]; exist {
		return command(args...)
	}

	return nil, errors.New("Command does not exists, or it's not registered")
}

// CreateRegistrator create registrator struct
// to be pushed to Socket registry
func CreateRegistrator(agent agent.Agent) *Registrator {
	r := &Registrator{
		Agent:    agent,
		Commands: make(map[string]Command),
	}

	r.Command("workers:list", r.WorkersList)
	r.Command("workers:kill", r.KillWorker)

	return r
}
