package debug

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli"
	"github.com/wiremind/koherence/bs"
)

var BlockStorageCommand = cli.Command{
	Name:   "bs",
	Usage:  "display block-storage information",
	Action: debugBs,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:     "type",
			Usage:    "Read infos from specified type",
			Required: true,
		},
		cli.StringFlag{
			Name:     "name",
			Usage:    "Device name to check",
			Required: true,
		},
	},
}

func debugBs(clicontext *cli.Context) error {
	var infos *bs.BlockStorageInfos
	var err error

	infosType := clicontext.String("type")
	devName := clicontext.String("name")

	switch infosType = clicontext.String("type"); infosType {
	case "scsi":
		if infos, err = bs.ScsiDeviceInfos(devName); err != nil {
			return err
		}
	case "virtio-blk":
		if infos, err = bs.VirtioBlkDeviceInfos(devName); err != nil {
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
