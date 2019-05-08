// refer: https://github.com/RichardKnop/machinery/blob/master/example/machinery.go
package main

import (
	"os"

	"github.com/si9ma/KillOJ-judger/judge"

	"github.com/si9ma/KillOJ-judger/config"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"
	"github.com/urfave/cli"
)

var (
	configPath = "conf/config.yml"
	workerTag  string
	app        *cli.App
	judger     *judge.Judger
)

func init() {
	app = cli.NewApp()
	app.Name = "kjudger"
	app.Usage = "kjudger is a judge judger for KillOJ(https://github.com/si9ma/KillOJ)"
	app.Author = "si9ma"
	app.Email = "si9may@tom.com"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "c",
			Value:       configPath,
			Destination: &configPath,
			Usage:       "Path to a configuration file",
		},
		cli.StringFlag{
			Name:        "tag",
			Value:       "kjudger",
			Destination: &workerTag,
			Usage:       "tag(name) of this judger",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		var (
			cfg *config.Config // app config
			err error
		)

		// Init log and configuration
		if cfg, err = Init(configPath); err != nil {
			log.Bg().Fatal("initialize fail", zap.String("configPath", configPath), zap.Error(err))
			return err
		}

		// new judger
		if judger, err = judge.NewJudger(*cfg, workerTag); err != nil {
			log.Bg().Fatal("create judger fail", zap.Error(err))
			return err
		}

		// judge
		if err = judger.Judge(); err != nil {
			log.Bg().Fatal("run judger fail", zap.Error(err))
			return err
		}

		return nil
	}

	app.After = func(context *cli.Context) error {
		return judger.Close()
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Bg().Fatal("run app fail", zap.Error(err))
	}
}
