package rosco

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"time"
)

type ScenarioFile struct {
	filePath string

	Name      string     `json:"Name"`
	Count     int        `json:"Count"`
	Date      time.Time  `json:"Date"`
	Summary   string     `json:"Summary"`
	ECUID     string     `json:"ECUID"`
	ECUSerial string     `json:"ECUSerial"`
	RawData   []*RawData `json:"MemsData"`
}

func NewScenarioFile(filepath string) *ScenarioFile {
	scenario := &ScenarioFile{}
	scenario.filePath = GetFullScenarioFilePath(filepath)
	scenario.Date = time.Now()

	return scenario
}

func (scenario *ScenarioFile) ConvertLogToScenario(id string) error {
	var err error

	// use the Responder to load the data
	responder := NewResponder()
	filename := GetFullScenarioFilePath(id)

	if err = responder.LoadScenario(filename); err == nil {
		name := strings.Replace(strings.ToLower(id), ".csv", ".fcr", 1)
		scenario.Name = name
		scenario.Count = responder.Playbook.Count
		scenario.RawData = responder.RawData
		scenario.Summary = fmt.Sprintf("Scenario file created from %s", id)
		if scenario.Count > 0 {
			scenario.Date, _ = ConvertTimeFieldToDate(responder.RawData[0].Time)
		}

		log.Infof("converted %s to %s", filename, scenario.filePath)
	} else {
		log.Errorf("error converting %s to %s", filename, scenario.filePath)
	}

	return err
}

func (scenario *ScenarioFile) Write() error {
	var err error
	var data []byte

	if data, err = json.MarshalIndent(scenario, "", " "); err != nil {
		log.Errorf("Error generating scenario description (%s)", err)
	} else {
		if err = ioutil.WriteFile(scenario.filePath, data, 0644); err != nil {
			log.Errorf("Error writing scenario file %s (%s)", scenario.filePath, err)
		}
	}

	return err
}

func (scenario *ScenarioFile) Read() error {
	var err error
	var data []byte

	if data, err = ioutil.ReadFile(scenario.filePath); err != nil {
		log.Errorf("Error reading scenario file %s (%s)", scenario.filePath, err)
		return err
	} else {
		if err = json.Unmarshal([]byte(data), &scenario); err != nil {
			log.Errorf("Error reading scenario file %s (%s)", scenario.filePath, err)
			return err
		}
	}

	log.Infof("scenario file: %+v", scenario)
	return err
}
