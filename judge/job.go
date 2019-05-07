package judge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	successTestCase int
	memSum          int64 // memory usage of all test case
	timeSum         int64 // time usage of all test case
}

// create working directory base on id
func (j *job) mkWorkDir() error {
	path := judgerWorkDir + "/" + string(j.submitID)

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Bg().Error("create job working directory fail",
			zap.Int("problemId", j.submitID), zap.Error(err))
		return err
	}

	j.workDir = path
	return nil
}

func (j *job) clean() {
	// remove working directory
	if _, err := os.Stat(j.workDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(j.workDir); err != nil {
			log.Bg().Error("remove job working directory fail",
				zap.Int("problemId", j.submitID), zap.Error(err))
		}
	}
}

// query submit detail info from db
func (j *job) querySubmitDetail(submitId int) (err error) {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "querySubmitDetail")
	defer span.Finish()

	span.SetTag("submitId", submitId)
	db := otgrom.SetSpanToGorm(ctx, j.db)

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

	log.For(ctx).Info("success query submit info", zap.Int("submitId", submitId))
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
	log.Bg().Info("write file success", zap.String("file", file))
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
	langArg := fmt.Sprintf("--lang %s", j.lang.Name)
	dirArg := fmt.Sprintf("--dir %s", j.workDir)
	srcArg := fmt.Sprintf("--src %s", j.lang.FileName)
	cmd.Args = append(cmd.Args, subcmdArg, langArg, dirArg, srcArg)
	log.For(ctx).Info("build compile command success",
		zap.String("cmd", strings.Join(cmd.Args, " ")))

	// compile
	if err := cmd.Run(); err != nil {
		log.For(ctx).Error("compile fail", zap.Error(err))
		return err
	}

	return nil
}

func (j *job) run() (err error) {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "batchRun")
	defer span.Finish()
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)
	span.SetTag("caseNum", len(j.ProblemTestCase))

	defer func() {
		if err != nil {
			log.For(ctx).Error("judge fail", zap.Error(err))
		}
	}()

	// log
	log.For(ctx).Info("judge problem",
		zap.Int("submitId", j.submitID),
		zap.Int("problemId", j.Problem.ID),
		zap.Int("caseNum", len(j.ProblemTestCase)))

	for _, testCase := range j.ProblemTestCase {
		// tracing
		innerSpan, innerCtx := opentracing.StartSpanFromContext(ctx, "run")
		innerSpan.SetTag("submitId", j.submitID)
		innerSpan.SetTag("problemId", j.Problem.ID)
		innerSpan.SetTag("caseId", testCase.ID)

		// log
		log.For(innerCtx).Info("run case",
			zap.Int("submitId", j.submitID),
			zap.Int("problemId", j.Problem.ID),
			zap.Int("caseId", testCase.ID))

		var cmd *exec.Cmd
		cmd, err = j.buildRunnerCmd(testCase)
		if err != nil {
			log.Bg().Error("build command fail", zap.Error(err))
			innerSpan.Finish()
			return err
		}
		log.For(innerCtx).Info("build run command success",
			zap.String("cmd", strings.Join(cmd.Args, " ")))

		if err = cmd.Run(); err != nil {
			// log
			log.For(innerCtx).Info("run case fail",
				zap.Int("submitId", j.submitID),
				zap.Int("problemId", j.Problem.ID),
				zap.Int("caseId", testCase.ID),
				zap.Error(err))
			innerSpan.Finish()
			return err
		}

		if err = j.handSingleRunResult(innerCtx); err != nil {
			innerSpan.Finish()
			if err != RunResultErr {
				return err
			}

			// break
			// let handlesandboxresult() to handle error
			return nil
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
		log.For(ctx).Error("unmarshal fail", zap.Error(err))
		return err
	}

	// log
	switch innerRes.Status {
	case judge.FAIL:
		log.For(ctx).Error("run fail", zap.ByteString("result", resStr))
		return RunResultErr
	case judge.SUCCESS:
		log.For(ctx).Info("run success", zap.ByteString("result", resStr))
		j.timeSum += innerRes.TimeLimit
		j.memSum += innerRes.MemLimit
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
	cmdArg := "--cmd /Main"
	inputArg := fmt.Sprintf("--input %s", testCase.InputData)
	dirArg := fmt.Sprintf("--dir %s", j.workDir)
	expectedArg := fmt.Sprintf("--expected %s", testCase.ExpectedOutput)
	scmpArg := "--seccomp"
	timeArg := fmt.Sprintf("--timeout %d", j.Problem.TimeLimit)
	memArg := fmt.Sprintf("--memory %d", j.Problem.MemoryLimit)

	// java is different
	if j.lang == codelang.LangJava {
		subcmdArg = "java"
		cmdArg = "--class Main"

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
func (j *job) handleSystemError(err error) {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "handleSystemError")
	defer span.Finish()
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)
	log.Bg().Error("judge system error", zap.Error(err))

	client := kredis.WrapRedisClient(ctx, j.redisdb)
	result := judge.OuterResult{
		ID:         string(j.submitID),
		Status:     judge.SystemErrorStatus,
		Message:    tip.SystemErrorTip.String(),
		IsComplete: true,
	}

	if v, err := json.Marshal(result); err != nil {
		log.For(ctx).Error("marshal result error", zap.Error(err))
	} else {
		key := constants.SubmitStatusKeyPrefix + string(j.submitID)
		val := string(v)
		client.Set(key, val, constants.SubmitStatusTimeout)
	}
}

func (j *job) buildCmd() (*exec.Cmd, error) {
	cmd := exec.Command(j.sandboxCfg.ExePath)
	j.sandboxOut = buffer.Buffer{} // new buffer
	cmd.Stdout = &j.sandboxOut

	// --id
	idArg := fmt.Sprintf("--id %d", j.submitID)
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
	logArg := fmt.Sprintf("--log %s", logPath)
	logfmtArg := fmt.Sprintf("--log-format %s", j.sandboxCfg.LogFormat)

	cmd.Args = append(cmd.Args, logArg, logfmtArg)
	return cmd, nil
}

func (j *job) handleSandboxResult() error {
	span, ctx := opentracing.StartSpanFromContext(j.ctx, "handleSandboxResult")
	defer span.Finish()
	span.SetTag("submitId", j.submitID)
	span.SetTag("problemId", j.Problem.ID)

	resStr := j.sandboxOut.Bytes()
	var innerRes judge.InnerResult
	if err := json.Unmarshal(resStr, &innerRes); err != nil {
		j.handleSystemError(err)
		log.For(ctx).Error("unmarshal fail", zap.Error(err))
		return err
	}

	// log
	switch innerRes.Status {
	case judge.FAIL:
		log.For(ctx).Error("fail", zap.ByteString("result", resStr))
	case judge.SUCCESS:
		log.For(ctx).Info("success", zap.ByteString("result", resStr))
	}

	// compile success
	if innerRes.Status == judge.SUCCESS &&
		innerRes.ResultType == judge.CompileResType {
		return nil
	}

	testCaseNum := len(j.ProblemTestCase)
	outerResult := innerRes.ToOuterResult()
	outerResult.IsComplete = true
	outerResult.TestCaseNum = testCaseNum
	outerResult.SuccessTestCase = j.successTestCase

	// replace time and memory usage with average value
	if innerRes.Status == judge.SUCCESS &&
		innerRes.ResultType == judge.RunResType {
		outerResult.Memory = j.memSum / int64(testCaseNum)
		outerResult.Runtime = j.timeSum / int64(testCaseNum)
	}

	client := kredis.WrapRedisClient(ctx, j.redisdb)
	if v, err := json.Marshal(outerResult); err != nil {
		log.For(ctx).Error("marshal result error", zap.Error(err))
	} else {
		key := constants.SubmitStatusKeyPrefix + string(j.submitID)
		val := string(v)
		client.Set(key, val, constants.SubmitStatusTimeout)
	}

	if innerRes.Status == judge.FAIL {
		return errors.New("sandbox error")
	}

	return nil
}
