package rosco

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
)

const (
	MemsFolder  = "MemsFCR"
	LogsFolder  = "Logs"
	DebugFolder = "Debug"
	sandbox     = false
)

func GetHomeFolder() string {
	var dir string
	var err error

	// get the home directory
	dir, err = homedir.Dir()

	if err != nil {
		log.Warnf("error getting home folder %s", err)
	}

	// override if sandboxed
	if runtime.GOOS == "darwin" && sandbox {
		// sandbox folder
		dir = "./Documents"
	}

	dir = fmt.Sprintf("%s/%s", dir, MemsFolder)

	return filepath.FromSlash(dir)
}

func GetAppFolder() string {
	// get the application binary current directory
	dir, err := os.Getwd()

	if err != nil {
		log.Warnf("error getting app folder %s", err)
	}

	return filepath.FromSlash(dir)
}

func GetLogFolder() string {
	dir := GetHomeFolder()
	dir = fmt.Sprintf("%s/%s", dir, LogsFolder)
	return filepath.FromSlash(dir)
}

func GetDebugFolder() string {
	dir := GetHomeFolder()
	dir = fmt.Sprintf("%s/%s", dir, DebugFolder)
	return filepath.FromSlash(dir)
}
