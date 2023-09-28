package debug

import (
	"log/slog"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/machine"
)

var MachineCommand = cli.Command{
	Name:   "machine",
	Usage:  "display machine information",
	Action: info,
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

func info(clicontext *cli.Context) error {
	var infosType string
	var infos *machine.MachineInfos
	var err error

	switch infosType = clicontext.String("type"); infosType {
	case "fs":
		if infos, err = machine.ReadFsInfos(); err != nil {
			return err
		}
	default:
		slog.Error("This infos type is not supported")
		return nil
	}

	slog.Info(
		"Fetched infos",
		slog.String("type", infosType),
		slog.String("struct", "MachineInfos"),
		slog.Any("infos", *infos),
	)

	return nil
}
