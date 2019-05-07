package config

import (
	"github.com/si9ma/KillOJ-common/asyncjob"
	"github.com/si9ma/KillOJ-common/kredis"
	"github.com/si9ma/KillOJ-common/mysql"
	"github.com/si9ma/KillOJ-common/sandbox"
)

type Config struct {
	AsyncJob    asyncjob.Config `yaml:"asyncJob"`
	Concurrency int             `yaml:"concurrency"` // the concurrency of judger
	Mysql       mysql.Config    `yaml:"mysql"`
	Redis       kredis.Config   `yaml:"redis"`
	Sandbox     sandbox.Config  `yaml:"sandbox"`
}
