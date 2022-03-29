package rosco

import (
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type ScenarioReader struct {
	connected    bool
	scenarioFile string
	Responder    *ScenarioResponder
}

func NewScenarioReader(filename string) *ScenarioReader {
	log.Infof("created scenario playback ecu reader")
	r := &ScenarioReader{}

	// initialise the responseMap
	responseMap = createResponseMap()

	// expand to full path
	//filename = fmt.Sprintf("%s/%s", GetLogFolder(), filename)
	filename = filepath.FromSlash(filename)
	r.scenarioFile = filename

	return r
}

func (r *ScenarioReader) Connect() (bool, error) {
	var err error

	log.Infof("connecting to scenario playback file %s", r.scenarioFile)

	r.Responder = NewResponder()

	if err = r.loadScenario(); err == nil {
		r.connected = true
	}

	return r.connected, err
}

func (r *ScenarioReader) SendAndReceive(command []byte) ([]byte, error) {
	var err error
	var data []byte

	data = r.Responder.GetECUResponse(command)
	log.Infof("read (%X) from scenario playback file", data)

	return data, err
}

func (r *ScenarioReader) Disconnect() error {
	var err error

	log.Infof("disconnected scenario playback file %s", r.scenarioFile)

	// disconnect
	r.connected = false
	r.scenarioFile = ""

	return err
}

func (r *ScenarioReader) loadScenario() error {
	log.Infof("loading scenario playback file %s", r.scenarioFile)
	return r.Responder.LoadScenario(r.scenarioFile)
}
