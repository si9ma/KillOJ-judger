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
	configPath = "config.yml"
	workerTag  string
	app        *cli.App
	cfg        *config.Config // app config
	err        error
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
	}
	app.Before = func(context *cli.Context) error {
		// Init log and configuration
		if cfg, err = Init(configPath); err != nil {
			log.Bg().Fatal("initialize fail", zap.String("configPath", configPath), zap.Error(err))
		}
		return err
	}
	app.Commands = []cli.Command{
		{
			Name:  "judger",
			Usage: "launch judge judger",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "tag",
					Value:       "kjudger",
					Destination: &workerTag,
					Usage:       "tag(name) of this judger",
				},
			},
			Action: func(c *cli.Context) error {
				var judger *judge.Judger
				// new judger
				if judger, err = judge.NewJudger(*cfg, workerTag); err != nil {
					log.Bg().Fatal("create judger fail", zap.Error(err))
				}

				defer judger.Close()
				if err = judger.Judge(); err != nil {
					log.Bg().Fatal("run judger fail", zap.Error(err))
				}
				return err
			},
		},
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Bg().Fatal("run app fail", zap.Error(err))
	}
}

//func send() error {
//	tracer, closer := tracing.NewTracer("sender")
//	opentracing.SetGlobalTracer(tracer)
//	defer closer.Close()
//
//	var judgeTask tasks.Signature
//
//	span, ctx := opentracing.StartSpanFromContext(context.Background(), "send")
//	defer span.Finish()
//
//	batchID := uuid.New().String()
//	span.SetBaggageItem("batch.id", batchID)
//	mlog.For(ctx).Info("", zap.String("batch.id", batchID))
//
//	judgeTask = tasks.Signature{
//		Name:       "judge",
//		RoutingKey: constants.ProjectName,
//	}
//
//	_, err := asyncjob.Server().SendTaskWithContext(ctx, &judgeTask)
//	if err != nil {
//		return fmt.Errorf("Could not send task: %s", err.Error())
//	}
//
//	return nil
//}
