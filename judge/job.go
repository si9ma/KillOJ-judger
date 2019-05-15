package judge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"go.uber.org/zap/buffer"

	"github.com/si9ma/KillOJ-common/sandbox"
	"github.com/si9ma/KillOJ-common/utils"

	"github.com/si9ma/KillOJ-common/codelang"

	"github.com/si9ma/KillOJ-common/constants"

	"github.com/si9ma/KillOJ-common/kredis"

	"github.com/si9ma/KillOJ-common/judge"
	"github.com/si9ma/KillOJ-common/tip"

	"github.com/si9ma/KillOJ-common/model"
	otgrom "github.com/smacker/opentracing-gorm"

	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
)

const judgerWorkDir = "/tmp/kjudger" // working directory of kjudger
const sandboxLogPath = "log/sandbox.log"

var RunResultErr = errors.New("run result error")

type job struct {
	submitID        int
	Submit          model.Submit
	Problem         model.Problem
	ProblemTestCase []model.ProblemTestCase
	workDir         string
	ctx             context.Context
	tracer          opentracing.Tracer
	db              *gorm.DB
	redisdb         *redis.ClusterClient
	lang            codelang.Language
	sandboxCfg      sandbox.Config

	sandboxOut      buffer.Buffer
	sandboxErr      buffer.Buffer
	successTestCase int
	memSum          int64 // memory usage of all test case
	timeSum         int64 // time usage of all test case
}

// create working directory base on id
func (j *job) mkWorkDir() error {
	path := judgerWorkDir + "/" + strconv.Itoa(j.submitID)

	log.Bg().Info("creating job working directory", zap.Int("submitId", j.submitID))

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Bg().Error("create job working directory fail",
			zap.Int("submitId", j.submitID), zap.Error(err))
		return err
	}

	j.workDir = path
	return nil
}

func (j *job) clean() {
	// remove working directory
	log.Bg().Info("remove job working directory", zap.Int("submitId", j.submitID))
	if _, err := os.Stat(j.workDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(j.workDir); err != nil {
			log.Bg().Error("remove job working directory fail",
				zap.Int("submitId", j.submitID), zap.Error(err))
		}
	}
}

// query submit detail info from db
func (j *job) querySubmitDetail(submitId int) (err error) {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "querySubmitDetail")
	defer span.Finish()

	span.SetTag("submitId", submitId)
	db := otgrom.SetSpanToGorm(ctx, j.db)

	log.For(ctx).Info("query submit detail", zap.Int("submitId", j.submitID))

	// query submit info
	if err := db.First(&j.Submit, submitId).Error; err != nil {
		log.For(ctx).Error("query submit error ",
			zap.Int("submitId", submitId), zap.Error(err))
		return err
	}

	// query problem info
	if err := db.Model(&j.Submit).Related(&j.Problem).Error; err != nil {
		log.For(ctx).Error("query problem with submit info error ",
			zap.Int("submitId", submitId), zap.Error(err))
		return err
	}

	// query test case list
	if err := db.Model(&j.Problem).Related(&j.ProblemTestCase).Error; err != nil {
		log.For(ctx).Error("query test case list with problem info error ",
			zap.Int("problemId", j.Problem.ID), zap.Error(err))
		return err
	}

	log.For(ctx).Info("success query submit detail", zap.Int("submitId", submitId))
	return nil
}

func (j *job) saveSrcCode() error {
	lang, err := codelang.GetLangByCode(j.Submit.Language)
	if err != nil {
		log.Bg().Error("get language error",
			zap.Int("langCode", j.Submit.Language), zap.Error(err))
		return err
	}

	file := j.workDir + "/" + lang.FileName
	if err := ioutil.WriteFile(file, []byte(j.Submit.SourceCode), os.ModePerm); err != nil {
		log.Bg().Error("write file error",
			zap.String("file", file), zap.Error(err))
		return err
	}

	j.lang = *lang
	log.Bg().Info("save user source code successful", zap.String("file", file),
		zap.Int("submitId", j.submitID))
	return nil
}

func (j *job) compile() (err error) {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "compile")
	defer span.Finish()
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)

	// build command
	cmd, err := j.buildCmd()
	if err != nil {
		log.Bg().Error("build command fail", zap.Error(err))
		return err
	}
	subcmdArg := "compile"
	langArg := fmt.Sprintf("--lang=%s", j.lang.Name)
	dirArg := fmt.Sprintf("--dir=%s", j.workDir)
	srcArg := fmt.Sprintf("--src=%s", j.lang.FileName)
	cmd.Args = append(cmd.Args, subcmdArg, langArg, dirArg, srcArg)
	log.For(ctx).Info("build compile command successful",
		zap.String("cmd", strings.Join(cmd.Args, " ")))

	// compile
	if err := cmd.Run(); err != nil {
		log.For(ctx).Error("compile source code fail", zap.String("stderr", j.sandboxErr.String()), zap.Error(err))
		return err
	}

	log.Bg().Info("compile source code success",
		zap.Int("submitId", j.submitID),
		zap.String("stdout", j.sandboxOut.String()),
		zap.String("stderr", j.sandboxErr.String()))

	return nil
}

func (j *job) run() (err error) {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "batchRun")
	defer span.Finish()
	defer func() {
		if err != nil {
			log.For(ctx).Error("run user code fail", zap.Error(err))
		}
	}()

	// tracing
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)
	span.SetTag("TestCaseNum", len(j.ProblemTestCase))

	// log
	log.For(ctx).Info("judge user code",
		zap.Int("submitId", j.submitID),
		zap.Int("problemId", j.Problem.ID),
		zap.Int("TestCaseNum", len(j.ProblemTestCase)))

	// run all test case
	for _, testCase := range j.ProblemTestCase {
		// tracing
		innerSpan, innerCtx := opentracing.StartSpanFromContext(ctx, "run")
		innerSpan.SetTag("submitId", j.submitID)
		innerSpan.SetTag("problemId", j.Problem.ID)
		innerSpan.SetTag("caseId", testCase.ID)

		// log
		log.For(innerCtx).Info("run test case",
			zap.Int("submitId", j.submitID),
			zap.Int("problemId", j.Problem.ID),
			zap.Int("caseId", testCase.ID))

		var cmd *exec.Cmd
		cmd, err = j.buildRunnerCmd(testCase)
		if err != nil {
			log.Bg().Error("build run command fail", zap.Error(err))
			innerSpan.Finish()
			return err
		}
		log.For(innerCtx).Info("build run command success",
			zap.String("cmd", strings.Join(cmd.Args, " ")))

		// run
		if err = cmd.Run(); err != nil {
			// log
			log.For(innerCtx).Info("run case fail",
				zap.Int("submitId", j.submitID),
				zap.Int("problemId", j.Problem.ID),
				zap.Int("caseId", testCase.ID),
				zap.Error(err),
				zap.String("stderr", j.sandboxErr.String()))

			innerSpan.Finish()
			return err
		}

		if err = j.handSingleRunResult(innerCtx); err != nil {
			innerSpan.Finish()
			return err
		}

		innerSpan.Finish()
	}

	return nil
}

func (j *job) handSingleRunResult(ctx context.Context) error {
	// handle result
	resStr := j.sandboxOut.Bytes()
	var innerRes judge.InnerResult
	if err := json.Unmarshal(resStr, &innerRes); err != nil {
		j.handleSystemError(err)
		log.For(ctx).Error("unmarshal result fail", zap.Error(err))
		return err
	}

	// log
	switch innerRes.Status {
	case judge.FAIL:
		log.For(ctx).Error("run user code fail",
			zap.Int("submitId", j.submitID),
			zap.String("result", string(resStr)),
			zap.String("stderr", j.sandboxErr.String()))
		return RunResultErr
	case judge.SUCCESS:
		log.For(ctx).Info("run user code success",
			zap.Int("submitId", j.submitID), zap.String("result", string(resStr)))
		j.timeSum += innerRes.TimeLimit
		j.memSum += innerRes.MemLimit
		j.successTestCase++
		return nil
	}

	return nil
}

func (j *job) buildRunnerCmd(testCase model.ProblemTestCase) (*exec.Cmd, error) {
	// build command
	cmd, err := j.buildCmd()
	if err != nil {
		return nil, err
	}

	// arguments
	subcmdArg := "run"
	cmdArg := "--cmd=/Main"
	inputArg := fmt.Sprintf("--input=%s", testCase.InputData)
	dirArg := fmt.Sprintf("--dir=%s", j.workDir)
	expectedArg := fmt.Sprintf("--expected=%s", testCase.ExpectedOutput)
	scmpArg := "--seccomp"
	timeArg := fmt.Sprintf("--timeout=%d", j.Problem.TimeLimit)
	memArg := fmt.Sprintf("--memory=%d", j.Problem.MemoryLimit)

	// java is different
	if j.lang == codelang.LangJava {
		subcmdArg = "java"
		cmdArg = "--class=Main"

		// no --seccomp
		cmd.Args = append(cmd.Args, subcmdArg, cmdArg, inputArg,
			dirArg, expectedArg, timeArg, memArg)
	} else {
		cmd.Args = append(cmd.Args, subcmdArg, cmdArg, inputArg,
			dirArg, expectedArg, scmpArg, timeArg, memArg)
	}

	return cmd, nil
}

// System error when error
func (j *job) handleSystemError(err error) error {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "handleSystemError")
	defer span.Finish()
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)
	log.For(ctx).Error("system error arise, handle system error", zap.Error(err))

	result := judge.OuterResult{
		ID:         string(j.submitID),
		Status:     judge.SystemErrorStatus,
		Message:    tip.SystemErrorTip.String(),
		IsComplete: true,
	}

	return j.saveResult(ctx, &result)
}

func (j *job) buildCmd() (*exec.Cmd, error) {
	cmd := exec.Command(j.sandboxCfg.ExePath)
	j.sandboxOut = buffer.Buffer{} // new buffer
	j.sandboxErr = buffer.Buffer{}
	cmd.Stdout = &j.sandboxOut
	cmd.Stderr = &j.sandboxErr

	// --id
	idArg := fmt.Sprintf("--id=%d", j.submitID)
	cmd.Args = append(cmd.Args, idArg)

	if !j.sandboxCfg.EnableLog {
		// disable log
		return cmd, nil
	}

	// sandbox log
	logPath, err := utils.MkDirAll4RelativePath(j.sandboxCfg.LogPath)
	if err != nil {
		log.Bg().Error("init sandbox log fail",
			zap.String("relativeLogPath", j.sandboxCfg.LogPath),
			zap.Error(err))
		return nil, err
	}
	logArg := fmt.Sprintf("--log=%s", logPath)
	logfmtArg := fmt.Sprintf("--log-format=%s", j.sandboxCfg.LogFormat)

	cmd.Args = append(cmd.Args, logArg, logfmtArg)
	return cmd, nil
}

func (j *job) handleSandboxResult() error {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "handleSandboxResult")
	defer span.Finish()
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)
	log.For(ctx).Info("handle the result of sandbox",
		zap.Int("submitId", j.submitID), zap.Int("problemId", j.Problem.ID))

	resStr := j.sandboxOut.Bytes()
	var innerRes judge.InnerResult
	if err := json.Unmarshal(resStr, &innerRes); err != nil {
		j.handleSystemError(err)
		log.For(ctx).Error("unmarshal result fail", zap.Error(err))
		return err
	}

	// log
	switch innerRes.Status {
	case judge.FAIL:
		log.For(ctx).Error("the result of sandbox run is fail",
			zap.String("result", string(resStr)),
			zap.String("stderr", j.sandboxErr.String()))
	case judge.SUCCESS:
		log.For(ctx).Info("run sandbox success", zap.String("result", string(resStr)))
	}

	// if compile success
	if innerRes.Status == judge.SUCCESS &&
		innerRes.ResultType == judge.CompileResType {
		return nil
	}

	// InnerResult convert to OuterResult
	testCaseNum := len(j.ProblemTestCase)
	outerResult := innerRes.ToOuterResult()
	outerResult.IsComplete = true
	outerResult.TestCaseNum = testCaseNum
	outerResult.SuccessTestCase = j.successTestCase

	// replace time and memory usage with average value when successful
	if innerRes.Status == judge.SUCCESS &&
		innerRes.ResultType == judge.RunResType {
		outerResult.Memory = j.memSum / int64(testCaseNum)
		outerResult.Runtime = j.timeSum / int64(testCaseNum)
	}

	if err := j.saveResult(ctx, &outerResult); err != nil {
		return err
	}

	if innerRes.Status == judge.FAIL {
		return errors.New("sandbox error")
	}

	return nil
}

func (j *job) saveResult(ctx context.Context, result *judge.OuterResult) error {
	client := kredis.WrapRedisClusterClient(ctx, j.redisdb)
	db := otgrom.SetSpanToGorm(ctx, j.db)

	// save result
	if v, err := json.Marshal(result); err != nil {
		log.For(ctx).Error("marshal result error", zap.Error(err))
		return err
	} else {
		key := constants.SubmitStatusKeyPrefix + strconv.Itoa(j.submitID)
		val := string(v)
		if err := client.Set(key, val, constants.SubmitStatusTimeout).Err(); err != nil {
			log.For(ctx).Error("save result to redis fail", zap.Error(err))
			return err
		}
		log.For(ctx).Info("save result to redis success",
			zap.String("key", key), zap.String("result", val))
	}

	// save is complete
	k := constants.UserProblemSubmitIsCompletePrefix + strconv.Itoa(j.Submit.UserID) + "_" + strconv.Itoa(j.Submit.ProblemID)
	if err := client.Set(k, true, constants.SubmitStatusTimeout).Err(); err != nil {
		log.For(ctx).Error("save is complete of submit to redis fail", zap.Error(err),
			zap.Int("submidID", j.submitID))
		return err
	}

	// update db
	j.Submit.Result = result.Status.Code
	j.Submit.RunTime = int(result.Runtime)
	j.Submit.MemoryUsage = int(result.Memory)
	j.Submit.IsComplete = true
	err := db.Save(&j.Submit).Error
	if err != nil {
		log.For(ctx).Error("save run sandbox result to db fail", zap.Error(err),
			zap.Int("submitID", j.submitID))
		return err
	}

	return nil
}
