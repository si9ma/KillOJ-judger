package utils

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/si9ma/KillOJ-common/constants"
)

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)

	// Config file not found
	if err != nil {
		return nil, fmt.Errorf("Open file error: %s", err)
	}

	// Config file found, let's try to read it
	data := make([]byte, 10000)
	count, err := file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("Read from file error: %s", err)
	}

	return data[:count], nil
}

func IsDebug() bool {
	// on debug mode, write log to stdout at the same time
	debug := false
	var err error

	if e := os.Getenv(constants.EnvDebug); e != "" {
		// if error , return false
		if debug, err = strconv.ParseBool(e); err != nil {
			return false
		}
	}

	return debug
}

func MkDirAll4RelativePath(relativePath string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absolutePath := strings.Join([]string{pwd, relativePath}, "/")
	err = MkDirAll4Path(absolutePath)

	return absolutePath, err
}

func MkDirAll4Path(p string) error {
	dir := path.Dir(p)

	// create directory if directory not exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
