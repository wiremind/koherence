package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/version"
)

func main() {
	var exitValue int

	app := cli.NewApp()
	app.Name = "koherence"
	app.Usage = "Openstack/Kube/Machine coherence checker"
	app.Version = version.Version

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "exit, e",
			Value:       1,
			Usage:       "value returned on error",
			Destination: &exitValue,
		},
	}

	app.Commands = []cli.Command{
		debugCommand,
		checkCommand,
	}

	err := app.Run(os.Args)
	if err != nil {
		slog.Error(
			"main() error",
			slog.String("error", err.Error()),
		)
		os.Exit(exitValue)
	}
}
