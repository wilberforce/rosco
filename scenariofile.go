package rosco

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

type ScenarioFile struct {
	filePath string

	Name    string     `json:"Name"`
	Count   int        `json:"Count"`
	Date    time.Time  `json:"Date"`
	Summary string     `json:"Summary"`
	RawData []*RawData `json:"MemsData"`
}

func NewScenarioFile(filepath string) *ScenarioFile {
	scenario := &ScenarioFile{}
	scenario.filePath = filepath
	scenario.Date = time.Now()

	return scenario
}

func (scenario *ScenarioFile) ConvertLogToScenario(id string) error {
	// use the responder to load the data
	responder := NewResponder()
	filename := getScenarioPath(id)
	err := responder.LoadScenario(filename)

	if err == nil {
		scenario.Name = id
		scenario.Count = responder.Playbook.Count
		scenario.RawData = responder.RawData
	}

	return err
}

func (scenario *ScenarioFile) Write() error {
	json, err := json.MarshalIndent(scenario, "", " ")

	if err != nil {
		log.Errorf("Error generating scenario description (%s)", err)
	} else {
		err = ioutil.WriteFile(scenario.filePath, json, 0644)

		if err != nil {
			log.Errorf("Error writing scenario file %s (%s)", scenario.filePath, err)
		}
	}

	return err
}

func (scenario *ScenarioFile) Read() error {
	file, err := ioutil.ReadFile(scenario.filePath)

	if err != nil {
		log.Errorf("Error reading scenario file %s (%s)", scenario.filePath, err)
	}

	err = json.Unmarshal([]byte(file), &scenario)

	if err != nil {
		log.Errorf("Error reading scenario file %s (%s)", scenario.filePath, err)
	}

	log.Infof("scenario file: %+v", scenario)

	return err
}
