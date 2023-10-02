package main

import (
	"github.com/urfave/cli"
	"github.com/wiremind/koherence/cmd/koherence/debug"
)

var debugCommand = cli.Command{
	Name:  "debug",
	Usage: "debug utilities",
	Subcommands: []cli.Command{
		debug.MachineCommand,
		debug.BlockStorageCommand,
	},
}
