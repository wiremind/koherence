package check

import (
	"log/slog"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/machine"
)

var MachineCommand = cli.Command{
	Name:   "machine",
	Usage:  "check machine informations",
	Action: machineCheckerCommand,
	Flags:  []cli.Flag{},
}

func machineCheckerCommand(clicontext *cli.Context) error {
	_, err := MachineChecker()
	if err != nil {
		slog.Error(
			"Could not check machine infos.",
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func MachineChecker() (*machine.MachineInfos, error) {
	var infosFs *machine.MachineInfos
	var infosProvider *machine.MachineInfos
	var err error

	if infosFs, err = machine.ReadFsInfos(); err != nil {
		return nil, err
	}

	switch infosFs.SysVendor {
	case machine.ProviderOpenstack:
		if infosProvider, err = machine.ReadOpenstackInfos(); err != nil {
			return nil, err
		}
	default:
		slog.Error(
			"This provider is not supported.",
			slog.String("provider", infosFs.SysVendor),
		)
		panic("provider not supported")
	}

	if infosFs.Uuid != infosProvider.Uuid || infosFs.Hostname != infosProvider.Hostname {
		slog.Error(
			"Machine informations mismatch.",
			slog.Any("fs", *infosFs),
			slog.Any("provider", *infosProvider),
		)
		panic("Machine informations mismatch.")
	}

	return infosFs, nil
}
