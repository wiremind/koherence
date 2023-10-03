package check

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/urfave/cli"
	"github.com/wiremind/koherence/bs"
	"github.com/wiremind/koherence/machine"
)

type bsTuple struct {
	Fs       *bs.BlockStorageInfos `json:"fs"`
	Provider *bs.BlockStorageInfos `json:"provider"`
}

type IncoherenceError struct{}

func (e *IncoherenceError) Error() string {
	return "Incoherence found on the machine !"
}

var BlockStorageCommand = cli.Command{
	Name:   "bs",
	Usage:  "check block-storage informations",
	Action: bsCheckerCommand,
}

func bsCheckerCommand(clicontext *cli.Context) error {
	var bsProvider map[uuid.UUID]*bs.BlockStorageInfos
	var err error

	machineInfos, err := MachineChecker()
	if err != nil {
		return err
	}

	bsFs := bs.ExtractBsInfos(machineInfos)

	switch machineInfos.SysVendor {
	case machine.ProviderOpenstack:
		bsProvider, err = bs.OpenstackGetBlockStorage(machineInfos.Uuid)
		if err != nil {
			return err
		}
	default:
		slog.Error(
			"Provider not supported.",
			slog.String("provider", machineInfos.SysVendor),
		)
		panic("Provider not supported.")
	}

	bsDiff := map[uuid.UUID]*bsTuple{}
	bsMerged := map[uuid.UUID]*bs.BlockStorageInfos{}

	for k, v := range bsFs {
		// Ignore boot block-storage
		if v.Uuid == uuid.Nil {
			continue
		}

		p, ok := bsProvider[k]
		if !ok {
			slog.Error(
				"Volume not found on provider.",
				slog.String("uuid", k.String()),
			)
			panic("Block-storage not found on provider API...")
		}

		// Check coherence...
		// TODO check size too...
		//if v.MachineId != p.MachineId || v.Device != filepath.Base(p.Device) || v.Size != p.Size {
		if v.MachineId != p.MachineId || v.Device != filepath.Base(p.Device) {
			bsDiff[v.Uuid] = &bsTuple{Fs: v, Provider: p}

			slog.Debug(
				"Found incoherence.",
				slog.Any("fs", *v),
				slog.Any("provider", *p),
			)
		} else {
			bsMerged[v.Uuid] = &bs.BlockStorageInfos{
				Uuid:      v.Uuid,
				FullUuid:  p.FullUuid,
				MachineId: v.MachineId,
				Size:      v.Size,
				Device:    v.Device,
				Type:      v.Type,
				BlockDev:  v.BlockDev,
				Status:    p.Status,
				Metadata:  p.Metadata,
			}
		}
	}

	var b []byte

	all := map[string]interface{}{"merged": bsMerged, "diff": bsDiff}
	b, err = json.Marshal(all)
	if err != nil {
		slog.Error("Cannot encode in JSON")
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))

	if len(bsDiff) > 0 {
		return &IncoherenceError{}
	}

	return nil
}
