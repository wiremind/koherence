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
	},
}

func debugBs(clicontext *cli.Context) error {
	var err error

	infosType := clicontext.String("type")
	infos := bs.ExtractBsInfos(infosType)

	b, err := json.Marshal(infos)
	if err != nil {
		slog.Error("Cannot encode in JSON")
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))

	return nil
}
