// refer: https://github.com/RichardKnop/machinery/blob/master/example/machinery.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/si9ma/KillOJ-common/asyncjob"

	"github.com/RichardKnop/machinery/v1/tasks"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/constants"

	"github.com/si9ma/KillOJ-common/tracing"

	"github.com/google/uuid"

	"github.com/RichardKnop/machinery/v1/log"
	mlog "github.com/si9ma/KillOJ-common/log"
	mytasks "github.com/si9ma/KillOJ-judger/tasks"
	"github.com/urfave/cli"
)

var (
	app        *cli.App
	configPath string
)

func init() {
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "kjudger"
	app.Usage = "kjudger is a judge worker for KillOJ(https://github.com/si9ma/KillOJ)"
	app.Author = "si9ma"
	app.Email = "si9may@tom.com"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "c",
			Value:       "",
			Destination: &configPath,
			Usage:       "Path to a configuration file",
		},
	}
}

func init() {

	cfg := asyncjob.Config{
		Broker:        "amqp://si9ma:rabbitmq@localhost:5672/",
		DefaultQueue:  "judger",
		Exchange:      constants.ProjectName,
		ExchangeType:  "direct",
		BindingKey:    constants.ProjectName,
		PrefetchCount: 3,
	}

	if err := asyncjob.Init(cfg); err != nil {
		mlog.Bg().Error("init machinery fail")
	}
}

func main() {
	// Set the CLI app commands
	app.Commands = []cli.Command{
		{
			Name:  "worker",
			Usage: "launch judge worker",
			Action: func(c *cli.Context) error {
				if err := worker(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
		{
			Name:  "send",
			Usage: "send example tasks ",
			Action: func(c *cli.Context) error {
				if err := send(); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
	}

	// Run the CLI app
	app.Run(os.Args)
}

func worker() error {
	tracer, closer := tracing.NewTracer("worker")
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	if err := asyncjob.Server().RegisterTask("judge", mytasks.Judge); err != nil {
		mlog.Bg().Error("register task fail", zap.Error(err))
	}
	worker := asyncjob.Server().NewWorker("worker", 0)

	// Here we inject some custom code for error handling,
	// start and end of task hooks, useful for metrics for example.
	errorhandler := func(err error) {
		log.ERROR.Println("I am an error handler:", err)
	}

	pretaskhandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am a start of task handler for:", signature.Name)
	}

	posttaskhandler := func(signature *tasks.Signature) {
		log.INFO.Println("I am an end of task handler for:", signature.Name)
	}

	worker.SetPostTaskHandler(posttaskhandler)
	worker.SetErrorHandler(errorhandler)
	worker.SetPreTaskHandler(pretaskhandler)

	return worker.Launch()
}

func send() error {
	tracer, closer := tracing.NewTracer("sender")
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	var judgeTask tasks.Signature

	span, ctx := opentracing.StartSpanFromContext(context.Background(), "send")
	defer span.Finish()

	batchID := uuid.New().String()
	span.SetBaggageItem("batch.id", batchID)
	mlog.For(ctx).Info("", zap.String("batch.id", batchID))

	judgeTask = tasks.Signature{
		Name: "judge",
	}

	_, err := asyncjob.Server().SendTaskWithContext(ctx, &judgeTask)
	if err != nil {
		return fmt.Errorf("Could not send task: %s", err.Error())
	}

	return nil
}
