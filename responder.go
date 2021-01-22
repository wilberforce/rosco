package rosco

import (
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
)

// RawData represents the raw data from the log file
type RawData struct {
	Dataframe7d string `csv:"0x7d_raw"`
	Dataframe80 string `csv:"0x80_raw"`
}

// PlaybookResponse type
type PlaybookResponse struct {
	Dataframe7d []byte
	Dataframe80 []byte
}

// Playbook struct
type Playbook struct {
	Responses         []PlaybookResponse
	Position          int
	Count             int
	servedDataframe7d bool
	servedDataframe80 bool
}

// ScenarioDescription describes the scenario for the ui
type ScenarioDescription struct {
	Name       string `json:"name"`
	Count      int    `json:"Count"`
	Position   int    `json:"Position"`
	Status     string `json:"status"`
	SampleData []int  `json:"sample"`
}

// Responder struct
type Responder struct {
	file     *os.File
	RawData  []*RawData
	Playbook Playbook
	//responseMap map[string][]byte
}

// NewResponder creates an instance of a responder
func NewResponder() *Responder {
	responder := &Responder{}
	return responder
}

// Open the CSV scenario file
func (responder *Responder) openFile(filepath string) error {
	var err error

	responder.file, err = os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error opening scenario file")
	}

	return err
}

// Load the scenario
func (responder *Responder) loadScenarioCSV(filepath string) error {
	var err error

	_ = responder.openFile(filepath)

	if err = gocsv.Unmarshal(responder.file, &responder.RawData); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error parsing scenario file")
	} else {
		log.WithFields(log.Fields{"Count": len(responder.RawData)}).Info("scenario loaded successfully")
	}

	return err
}

// LoadScenario loads a scenario for playing from the ECU
func (responder *Responder) LoadScenario(filepath string) error {
	err := responder.loadScenarioCSV(filepath)

	if err == nil {
		// reset the Position of the Playbook
		responder.Playbook.Position = 0
		responder.Playbook.Count = len(responder.RawData)
		responder.Playbook.servedDataframe7d = false
		responder.Playbook.servedDataframe80 = false

		// iterate the scenario extracting the raw dataframes into a sequential Playbook
		for i := 0; i < len(responder.RawData); i++ {
			pr := PlaybookResponse{}
			pr.Dataframe7d = responder.convertHexStringToByteArray(responder.RawData[i].Dataframe7d)
			pr.Dataframe80 = responder.convertHexStringToByteArray(responder.RawData[i].Dataframe80)

			responder.Playbook.Responses = append(responder.Playbook.Responses, pr)
		}
	}

	return err
}

// GetECUResponse returns an emulated response byte string
func (responder *Responder) GetECUResponse(cmd []byte) []byte {
	var data []byte

	// convert the command code to a string
	command := hex.EncodeToString(cmd)
	command = strings.ToUpper(command)

	// if the command is a dataframe request and we have a response file
	// then use the response file
	if responder.isDataframeRequest(command) {

		position := responder.Playbook.Position

		if command == "7D" {
			data = responder.Playbook.Responses[position].Dataframe7d
			// truncate to the right size
			data = data[:33]
			responder.Playbook.servedDataframe7d = true
		}

		if command == "80" {
			data = responder.Playbook.Responses[position].Dataframe80
			// truncate to the right size
			data = data[:29]
			responder.Playbook.servedDataframe80 = true
		}

		// served both dataframes from this Position, index on to the next Position
		if responder.Playbook.servedDataframe7d && responder.Playbook.servedDataframe80 {
			responder.Playbook.servedDataframe7d = false
			responder.Playbook.servedDataframe80 = false

			responder.Playbook.Position = responder.Playbook.Position + 1
			log.WithFields(log.Fields{"index": responder.Playbook.Position, "Count": responder.Playbook.Count}).Info("both dataframes served from scenario")

			// if we've reached the end then loop back to the start
			if responder.Playbook.Position >= responder.Playbook.Count {
				responder.Playbook.Position = 0
				log.Info("reached end of scenario, restarting from beginning")
			}
		}
	} else {
		// generate the relevant response
		data = responder.generateECUResponse(command)
	}

	return data
}

// determines where the command code is a dataframe request
func (responder *Responder) isDataframeRequest(command string) bool {
	return (command == "80" || command == "7D")
}

// converts the hex string to a byte array
func (responder *Responder) convertHexStringToByteArray(response string) []byte {
	// convert to byte array
	data, _ := hex.DecodeString(response)

	return data
}

// if we're responding to a command that isn't a dataframe request
// then generate the correct response
func (responder *Responder) generateECUResponse(command string) []byte {
	command = strings.ToUpper(command)

	r := responseMap[command]

	if r == nil {
		r = responseMap["00"]
		copy(r[0:], command)
	}

	log.WithFields(log.Fields{"response": fmt.Sprintf("%x", r), "command": command}).Info("generated response for command")
	return r
}
