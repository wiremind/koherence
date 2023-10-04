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

var OpenstackCommand = cli.Command{
	Name:   "openstack",
	Usage:  "display openstack information",
	Action: debugOpenstack,
}

func debugOpenstack(clicontext *cli.Context) error {
	infos, err := machine.ReadFsInfos()
	if err != nil {
		return err
	}

	bs, err := bs.OpenstackGetBlockStorage(infos)
	if err != nil {
		return err
	}

	b, err := json.Marshal(bs)
	if err != nil {
		slog.Error("Cannot encode in JSON")
		return err
	}

	fmt.Fprintln(os.Stdout, string(b))

	return nil
}
