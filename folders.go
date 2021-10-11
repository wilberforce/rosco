package rosco

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"runtime"
)

const (
	MemsFolder  = "MemsFCR"
	LogsFolder  = "Logs"
	DebugFolder = "Debug"
)

func GetHomeFolder() string {
	var dir string

	// get the home directory
	if runtime.GOOS == "darwin" {
		// sandbox folder
		dir = "./Documents"
	} else {
		dir, _ = homedir.Dir()
	}

	dir = fmt.Sprintf("%s/%s", dir, MemsFolder)

	return filepath.FromSlash(dir)
}

func GetAppFolder() string {
	// get the application binary current directory
	dir, _ := os.Getwd()
	return filepath.FromSlash(dir)
}

func GetDebugFolder() string {
	dir := GetHomeFolder()
	dir = fmt.Sprintf("%s/%s/%s", dir, MemsFolder, LogsFolder)
	return filepath.FromSlash(dir)
}

func GetLogFolder() string {
	dir := GetHomeFolder()
	dir = fmt.Sprintf("%s/%s/%s", dir, MemsFolder, DebugFolder)
	return filepath.FromSlash(dir)
}
