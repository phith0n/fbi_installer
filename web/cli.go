package web

import (
	"os"
	"path/filepath"

	"fbi_installer/logging"

	"github.com/urfave/cli/v2"
)

var logger = logging.GetSugar()
var WebCommand = &cli.Command{
	Name:  "start",
	Usage: "start the FBI Remote Installer daemon",
	Action: func(c *cli.Context) (err error) {
		listen := c.String("listen")
		DataDir, err = filepath.Abs(c.String("data-dir"))
		if err != nil {
			return cli.Exit("data dir must exist", 1)
		}
		if !IsFile(DataDir) {
			err = os.MkdirAll(DataDir, 0o755) //nolint:gosec
		}

		if err != nil {
			return cli.Exit("create data dir error: "+err.Error(), 1)
		}
		BaseURL = c.String("base-url")
		return StartGin(listen)
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "listen",
			Aliases: []string{"l"},
			Usage:   "listen address",
			Value:   ":8080",
		},
		&cli.StringFlag{
			Name:     "data-dir",
			Aliases:  []string{"d"},
			Usage:    "CIA data directory",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "base-url",
			Aliases:  []string{"u"},
			Usage:    "base url for this server",
			Required: false,
		},
	},
}
