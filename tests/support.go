package tests

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getLogfilename() string {
	currentTime := time.Now()
	dateTime := currentTime.Format("2006-01-02 15:04:05")
	dateTime = strings.ReplaceAll(dateTime, ":", "")
	dateTime = strings.ReplaceAll(dateTime, " ", "-")
	filename := fmt.Sprintf("debug-%s.log", dateTime)
	return filepath.FromSlash(filename)
}

func init() {
	var logWriter io.Writer

	writeToFile := false

	if writeToFile {
		// write logs to file and console
		filename := getLogfilename()

		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}

		logWriter = io.MultiWriter(os.Stdout, f)
	} else {
		logWriter = os.Stdout
	}

	// Output to stdout instead of the default stderr
	// and to a log file
	log.SetOutput(logWriter)

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	// enable function logging for tests
	//log.SetReportCaller(true)
}

func GetVirtualPort() string {
	if runtime.GOOS == "darwin" {
		// ensure memsulator is running for tests to pass
		homeFolder, _ := homedir.Dir()
		path := fmt.Sprintf("%s/ttyecu", homeFolder)
		log.Infof("using port %s", path)
		return filepath.FromSlash(path)
	}

	if runtime.GOOS == "windows" {
		path := "COM1"
		log.Infof("using port %s", path)
		return path
	}

	return ""
}
