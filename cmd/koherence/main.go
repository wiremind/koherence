package main

import (
	"fmt"
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

	err := app.Run(os.Args)
	if err != nil {
		// TODO do better
		panic(fmt.Sprintf("main: %s", err))
	}
}
