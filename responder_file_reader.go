package rosco

import (
	"fmt"
	"strings"
)

type ResponderFileInfo struct {
	Data        []*RawData
	Description ScenarioDescription
}

type ResponderFileReader interface {
	Load() (ResponderFileInfo, error)
}

func NewResponderFileReader(filepath string) (ResponderFileReader, error) {
	var err error
	var r ResponderFileReader

	if isCSVFile(filepath) {
		r = NewScenarioCSVReader(filepath)
	}

	if isFCRFile(filepath) {
		r = NewScenarioFCRReader(filepath)
	}

	if r == nil {
		err = fmt.Errorf("file reader for %s is not supported", filepath)
	}

	return r, err
}

func isCSVFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".csv")
}

func isFCRFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".fcr")
}
