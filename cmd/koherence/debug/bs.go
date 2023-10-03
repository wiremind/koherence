package debug

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/bs"
	"github.com/wiremind/koherence/machine"
)

var BlockStorageCommand = cli.Command{
	Name:   "bs",
	Usage:  "display block-storage information from filesystem",
	Action: debugBs,
}

func debugBs(clicontext *cli.Context) error {
	var err error

	machineInfos, err := machine.ReadFsInfos()
	if err != nil {
		return err
	}

	infos := bs.ExtractBsInfos(machineInfos)

	b, err := json.Marshal(infos)
	if err != nil {
		slog.Error("Cannot encode in JSON")
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))

	return nil
}
