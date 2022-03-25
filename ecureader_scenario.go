package rosco

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type ScenarioReader struct {
	connected    bool
	scenarioFile string
}

func NewScenarioReader(filename string) *ScenarioReader {
	log.Infof("created scenario playback ecu reader")

	r := &ScenarioReader{}

	// expand to full path
	filename = fmt.Sprintf("%s/%s", GetLogFolder(), filename)
	filename = filepath.FromSlash(filename)

	r.scenarioFile = filename
	log.Infof("loading scenario file %s", r.scenarioFile)

	return r
}

func (r *ScenarioReader) Connect() (bool, error) {
	var err error
	return r.connected, err
}

func (r *ScenarioReader) SendAndReceive(command []byte) ([]byte, error) {
	var err error
	var bytes []byte

	return bytes, err
}

func (r *ScenarioReader) Disconnect() error {
	var err error
	return err
}
