package check

import (
	"encoding/json"
	"errors"
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

func GetBsProvider(machineInfos *machine.MachineInfos) (map[uuid.UUID]*bs.BlockStorageInfos, error) {
	switch machineInfos.SysVendor {
	case machine.ProviderOpenstack:
		return bs.OpenstackGetBlockStorage(machineInfos)
	default:
		slog.Error(
			"Provider not supported.",
			slog.String("provider", machineInfos.SysVendor),
		)
		err := errors.New("provider not supported")
		return nil, err
	}
}

func BsMerge(bsFs map[uuid.UUID]*bs.BlockStorageInfos, bsProvider map[uuid.UUID]*bs.BlockStorageInfos) ([]byte, error) {
	var b []byte
	var err error

	bsDiff := map[uuid.UUID]*bsTuple{}
	BsMerged := map[uuid.UUID]*bs.BlockStorageInfos{}

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
			// XXX Provider think it is detached but it is not... Just create
			// an empty struct to be stored in bsDiff
			p = &bs.BlockStorageInfos{}
		}

		// Check coherence...
		// TODO check size too...
		if v.MachineId != p.MachineId || v.Device != filepath.Base(p.Device) || v.Size != p.Size {
			bsDiff[v.Uuid] = &bsTuple{Fs: v, Provider: p}

			slog.Debug(
				"Found incoherence.",
				slog.Any("fs", *v),
				slog.Any("provider", *p),
			)
		} else {
			BsMerged[v.Uuid] = &bs.BlockStorageInfos{
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

	all := map[string]interface{}{"merged": BsMerged, "diff": bsDiff}
	b, err = json.Marshal(all)
	if err != nil {
		slog.Error("Cannot encode in JSON")
		return nil, err
	}

	if len(bsDiff) > 0 {
		return b, &IncoherenceError{}
	}

	return b, nil
}

func bsCheckerCommand(clicontext *cli.Context) error {
	var err error

	machineInfos, err := MachineChecker()
	if err != nil {
		return err
	}

	bsFs := bs.ExtractBsInfos(machineInfos)

	bsProvider, err := GetBsProvider(machineInfos)
	if err != nil {
		return err
	}

	b, err := BsMerge(bsFs, bsProvider)

	if b != nil {
		fmt.Fprintln(os.Stdout, string(b))
	}

	return err
}
