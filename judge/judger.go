package judge

import (
	"context"
	"io"

	"github.com/si9ma/KillOJ-common/sandbox"

	"github.com/go-redis/redis"

	"github.com/si9ma/KillOJ-common/kredis"
	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/jinzhu/gorm"

	"github.com/RichardKnop/machinery/v1"
	"github.com/si9ma/KillOJ-common/tracing"
	"github.com/si9ma/KillOJ-judger/config"

	"github.com/opentracing/opentracing-go"
	"github.com/si9ma/KillOJ-common/asyncjob"
	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"
)

const serverName = "KillOj_Judger"

type Judger struct {
	tracer          opentracing.Tracer
	closer          io.Closer
	db              *gorm.DB
	redisdb         *redis.ClusterClient
	machineryServer *machinery.Server
	tag             string
	concurrency     int
	sandboxCfg      sandbox.Config
}

func NewJudger(cfg config.Config, workerTag string) (*Judger, error) {
	var (
		machineryServer *machinery.Server
		err             error
		db              *gorm.DB
		redisdb         *redis.ClusterClient
	)

	// init asyncjob server
	if machineryServer, err = asyncjob.Init(cfg.AsyncJob); err != nil {
		log.Bg().Error("Init asyncjob server fail", zap.Error(err))
		return nil, err
	}

	// init tracer
	tracer, closer := tracing.NewTracer(serverName)
	opentracing.SetGlobalTracer(tracer)

	// init db
	if db, err = mysql.InitDB(cfg.Mysql); err != nil {
		log.Bg().Error("Init mysql fail", zap.Error(err))
		return nil, err
	}

	// init redis
	if redisdb, err = kredis.Init(cfg.Redis); err != nil {
		log.Bg().Error("Init redis fail", zap.Error(err))
		return nil, err
	}

	return &Judger{
		tracer:          tracer,
		closer:          closer,
		machineryServer: machineryServer,
		tag:             workerTag,
		db:              db,
		concurrency:     cfg.Concurrency,
		redisdb:         redisdb,
		sandboxCfg:      cfg.Sandbox,
	}, nil
}

// close tracer
func (j *Judger) Close() (err error) {
	// close db
	if j.db != nil {
		if err = j.db.Close(); err != nil {
			log.Bg().Error("close db fail", zap.Error(err))
		}
	}

	// close redis
	if j.redisdb != nil {
		if err = j.redisdb.Close(); err != nil {
			log.Bg().Error("close redis fail", zap.Error(err))
		}
	}

	// close tracer
	if err = j.closer.Close(); err != nil {
		log.Bg().Error("close tracer fail", zap.Error(err))
	}

	return
}

func (j *Judger) Judge() error {
	if err := j.machineryServer.RegisterTask("judge", j.do); err != nil {
		log.Bg().Error("register task fail", zap.Error(err))
		return err
	}
	worker := j.machineryServer.NewWorker(j.tag, j.concurrency)

	return worker.Launch()
}

// just do it
func (j *Judger) do(ctx context.Context, submitId int) (err error) {
	// new job
	jb := &job{
		ctx:        ctx,
		tracer:     j.tracer,
		redisdb:    j.redisdb,
		db:         j.db,
		submitID:   submitId,
		sandboxCfg: j.sandboxCfg,
	}
	// *. clean
	defer jb.clean()

	// 1. query submit detail
	if err = jb.querySubmitDetail(submitId); err != nil {
		return jb.handleSystemError(err)
	}

	// 2.create job working directory
	if err = jb.mkWorkDir(); err != nil {
		log.For(ctx).Error("create job working directory fail", zap.Error(err))
		return jb.handleSystemError(err)
	}

	// 3. save source code
	if err = jb.saveSrcCode(); err != nil {
		log.For(ctx).Error("save source code fail", zap.Error(err))
		return jb.handleSystemError(err)
	}

	// 4. compile
	if err := jb.compile(); err != nil {
		return jb.handleSystemError(err)
	}

	// 5. handle compile result
	if err := jb.handleSandboxResult(); err != nil {
		return err
	}

	// 6. run
	if err := jb.run(); err != nil && err != RunResultErr {
		// treat error as system error,
		// only when err not equal RunResultErr
		return jb.handleSystemError(err)
	}

	// 7. handle run result
	if err := jb.handleSandboxResult(); err != nil {
		return err
	}

	return nil
}
