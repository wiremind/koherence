package check

import (
	"log/slog"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/machine"
)

var MachineCommand = cli.Command{
	Name:   "machine",
	Usage:  "check machine informations",
	Action: machineChecker,
	Flags:  []cli.Flag{},
}

func machineChecker(clicontext *cli.Context) error {
	var infosFs *machine.MachineInfos
	var infosProvider *machine.MachineInfos
	var err error

	if infosFs, err = machine.ReadFsInfos(); err != nil {
		return err
	}

	switch infosFs.SysVendor {
	case "OpenStack Foundation":
		if infosProvider, err = machine.ReadOpenstackInfos(); err != nil {
			return err
		}
	default:
		slog.Error(
			"This provider is not supported.",
			slog.String("provider", infosFs.SysVendor),
		)
		return nil
	}

	if infosFs.Uuid != infosProvider.Uuid || infosFs.Hostname != infosProvider.Hostname {
		slog.Error(
			"Machine informations mismatch.",
			slog.Any("fs", *infosFs),
			slog.Any("provider", *infosProvider),
		)
	} else {
		slog.Info(
			"Machine informations OK.",
			slog.Any("infos", *infosFs),
		)
	}

	return nil
}
