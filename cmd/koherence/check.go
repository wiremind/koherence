package main

import (
	"github.com/urfave/cli"
	"github.com/wiremind/koherence/cmd/koherence/check"
)

var checkCommand = cli.Command{
	Name:  "check",
	Usage: "check utilities",
	Subcommands: []cli.Command{
		check.MachineCommand,
	},
}
