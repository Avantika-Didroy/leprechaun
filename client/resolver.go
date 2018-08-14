package client

import "github.com/kilgaloon/leprechaun/cmd"

// Resolve resolves which command to call
func Resolve(command string, args []string) error {
	switch command {
	case "rainbow:transfer":
		go cmd.Transfer(args)
	}

	return nil
}