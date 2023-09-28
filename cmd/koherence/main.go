package main

import (
	"os"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "koherence"
	app.Usage = "Openstack/Kube/Machine coherence checker"
	app.Version = version.Version

	app.Commands = []cli.Command{
		debugCommand,
		checkCommand,
	}

	app.Run(os.Args)
}
