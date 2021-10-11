package rosco

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"runtime"
)

const (
	MemsFolder = "MemsFCR"
	LogsFolder = "Logs"
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

	return filepath.FromSlash(dir)
}

func GetAppFolder() string {
	// get the application binary current directory
	dir, _ := os.Getwd()
	return filepath.FromSlash(dir)
}

func GetLogFolder() string {
	dir := GetHomeFolder()
	return fmt.Sprintf("%s/%s/%s/", dir, MemsFolder, LogsFolder)
}
