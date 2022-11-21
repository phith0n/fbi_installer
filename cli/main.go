package main

import (
	"log"
	"os"

	"fbi_installer/logging"
	"fbi_installer/web"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

var logger = logging.GetSugar()

func main() {
	app := cli.App{
		Name:  "fbi_installer",
		Usage: "FBI Remote Installer is a tool help 3DS player to install games easier",
		Commands: []*cli.Command{
			web.WebCommand,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug mode",
				Value:   false,
			},
		},
		Before: func(context *cli.Context) error {
			debug := context.Bool("debug")
			err := logging.InitLogger(debug)
			if err != nil {
				return err
			}

			if debug {
				gin.SetMode(gin.DebugMode)
			} else {
				gin.SetMode(gin.ReleaseMode)
			}

			logger.Infof("debug mode = %v", debug)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
