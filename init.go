package main

import (
	"os"
	"strings"

	"github.com/si9ma/KillOJ-judger/config"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/utils"
	"gopkg.in/yaml.v2"

	"github.com/si9ma/KillOJ-common/log"
)

const logFilePath = "log/judger.log"

// init configuration
func Init(cfgPath string) (cfg *config.Config, err error) {
	var pwd string

	// get log path ( create parent directory is parent directory not exist)
	logPath, err := utils.MkDirAll4RelativePath(logFilePath)
	if err != nil {
		log.Bg().Error("Init log fail",
			zap.String("relativeLogPath", logFilePath), zap.Error(err))
		return nil, err
	}

	// init log
	if err := log.Init([]string{logPath}, log.Json); err != nil {
		log.Bg().Error("Init log fail",
			zap.String("logPath", logPath), zap.Error(err))
		return nil, err
	}

	if pwd, err = os.Getwd(); err != nil {
		log.Bg().Error("get current directory fail", zap.Error(err))
		return nil, err
	}

	// init configuration
	cfgPath = strings.Join([]string{pwd, cfgPath}, "/")
	if data, err := utils.ReadFile(cfgPath); err != nil {
		log.Bg().Error("Read config file fail",
			zap.String("cfgpath", cfgPath),
			zap.Error(err))
		return nil, err
	} else {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Bg().Error("Unmarshal YAML fail", zap.Error(err))
			return nil, err
		}
	}

	return cfg, nil
}
