package rosco

import (
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

// RawData represents the raw data from the log file
type RawData struct {
	Time        string `csv:"#time"`
	Dataframe7d string `csv:"0x7d_raw"`
	Dataframe80 string `csv:"0x80_raw"`
}

// PlaybookResponse type
type PlaybookResponse struct {
	Timestamp   time.Time
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
	Name       string    `json:"name"`
	Count      int       `json:"Count"`
	Position   int       `json:"Position"`
	Status     string    `json:"status"`
	Date       time.Time `json:"Date"`
	SampleData []int     `json:"sample"`
}

// Responder struct
type Responder struct {
	file     *os.File
	RawData  []*RawData
	Playbook Playbook
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

	if err = responder.openFile(filepath); err == nil {
		if err = gocsv.Unmarshal(responder.file, &responder.RawData); err != nil {
			log.WithFields(log.Fields{"error": err}).Error("error parsing scenario file")
		} else {
			log.WithFields(log.Fields{"Count": len(responder.RawData)}).Info("scenario loaded successfully")
		}
	}

	return err
}

// LoadScenario loads a scenario for playing from the ECU
func (responder *Responder) LoadScenario(filepath string) error {
	var timestamp time.Time
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

			// attempt to convert to time
			if timestamp, err = time.Parse("15:04:05.000", responder.RawData[i].Time); err != nil {
				if timestamp, err = time.Parse("15:04:05", responder.RawData[i].Time); err != nil {
					if timestamp, err = time.Parse("04:05.0", responder.RawData[i].Time); err != nil {
						log.Warnf("unable to parse timestamp %s, defaulting to current time", responder.RawData[i].Time)
						timestamp = time.Now()
					}
				}
			}

			pr.Timestamp = timestamp
			pr.Dataframe7d = responder.convertHexStringToByteArray(responder.RawData[i].Dataframe7d)
			pr.Dataframe80 = responder.convertHexStringToByteArray(responder.RawData[i].Dataframe80)

			responder.Playbook.Responses = append(responder.Playbook.Responses, pr)
		}
	}

	return err
}

// MovePositionToLocation finds and moves the position in the playbook to
// the time location specified
func (responder *Responder) MovePositionToLocation(timelocation time.Time) {
	for i, r := range responder.Playbook.Responses {
		if timelocation.Before(r.Timestamp) {
			log.Printf("moving position from %v to %v", responder.Playbook.Position, i)
			// set the position to the location after the specified time location
			responder.Playbook.Position = i - 1
			// exit loop
			break
		}
	}
}

func (responder *Responder) GetFirst() PlaybookResponse {
	return responder.Playbook.Responses[0]
}

func (responder *Responder) GetLast() PlaybookResponse {
	return responder.Playbook.Responses[responder.Playbook.Count-1]
}

func (responder *Responder) GetCurrent() PlaybookResponse {
	return responder.Playbook.Responses[responder.Playbook.Position]
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
		// get the position of the next response
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
