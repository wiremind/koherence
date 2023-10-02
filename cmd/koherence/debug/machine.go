package debug

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/machine"
)

var MachineCommand = cli.Command{
	Name:   "machine",
	Usage:  "display machine information",
	Action: debugMachine,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type",
			Usage: "Read infos from specified type",
			Value: "fs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "Format the output using the given Go template, e.g, '{{json .}}'",
		},
	},
}

func debugMachine(clicontext *cli.Context) error {
	var infosType string
	var infos *machine.MachineInfos
	var err error

	switch infosType = clicontext.String("type"); infosType {
	case "fs":
		if infos, err = machine.ReadFsInfos(); err != nil {
			return err
		}
	case "openstack":
		if infos, err = machine.ReadOpenstackInfos(); err != nil {
			return err
		}
	default:
		slog.Error("This infos type is not supported")
		return nil
	}

	b, err := json.Marshal(infos)
	if err != nil {
		slog.Error("Cannot encode in JSON")
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))

	return nil
}
