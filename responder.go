package rosco

import (
	"encoding/hex"
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
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
	Position    int
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

type ScenarioDetails struct {
	First   PlaybookResponse
	Current PlaybookResponse
	Last    PlaybookResponse
}

// ScenarioDescription describes the scenario for the ui
type ScenarioDescription struct {
	Name     string          `json:"name"`
	Count    int             `json:"Count"`
	Duration string          `json:"Duration"`
	Position int             `json:"Position"`
	Date     time.Time       `json:"Date"`
	Details  ScenarioDetails `json:"Details"`
	Summary  string          `json:"Summary"`
}

// ScenarioResponder struct
type ScenarioResponder struct {
	fileReader  ResponderFileReader
	file        *os.File
	RawData     []*RawData
	Playbook    Playbook
	Description ScenarioDescription
}

// NewResponder creates an instance of a Responder
func NewResponder() *ScenarioResponder {
	responder := &ScenarioResponder{}
	return responder
}

// LoadScenario loads a scenario for playing from the ECU
func (responder *ScenarioResponder) LoadScenario(filepath string) error {
	var err error
	var timestamp time.Time
	var info ResponderFileInfo

	if responder.fileReader, err = NewResponderFileReader(filepath); err == nil {
		if info, err = responder.fileReader.Load(); err == nil {
			responder.RawData = info.Data
			responder.Description = info.Description
			// reset the Position of the Playbook
			responder.Playbook.Position = 0
			responder.Playbook.Count = info.Description.Count
			responder.Playbook.servedDataframe7d = false
			responder.Playbook.servedDataframe80 = false

			// iterate the scenario extracting the raw dataframes into a sequential Playbook
			for i := 0; i < len(responder.RawData); i++ {
				pr := PlaybookResponse{}

				// attempt to convert to time
				timestamp, err = ConvertTimeFieldToDate(responder.RawData[i].Time)

				pr.Timestamp = timestamp
				pr.Dataframe7d = responder.convertHexStringToByteArray(responder.RawData[i].Dataframe7d)
				pr.Dataframe80 = responder.convertHexStringToByteArray(responder.RawData[i].Dataframe80)

				responder.Playbook.Responses = append(responder.Playbook.Responses, pr)
			}
		}
	}

	return err
}

func ConvertTimeFieldToDate(timeField string) (time.Time, error) {
	var err error
	var timestamp time.Time

	// attempt to convert to time
	if timestamp, err = time.Parse("2006-01-02 15:04:05.000", timeField); err != nil {
		if timestamp, err = time.Parse("15:04:05.000", timeField); err != nil {
			if timestamp, err = time.Parse("15:04:05", timeField); err != nil {
				if timestamp, err = time.Parse("04:05.0", timeField); err != nil {
					log.Warnf("unable to parse timestamp %s, defaulting to current time", timeField)
					timestamp = time.Now()
				}
			}
		}
	}

	return timestamp, err
}

// MovePositionToLocation finds and moves the position in the playbook to
// the time location specified
func (responder *ScenarioResponder) MovePositionToLocation(timelocation time.Time) {
	for i, r := range responder.Playbook.Responses {
		if timelocation.Before(r.Timestamp) {
			log.Infof("moving position from %v to %v", responder.Playbook.Position, i)
			// set the position to the location after the specified time location
			responder.Playbook.Position = i - 1
			// exit loop
			break
		}
	}
}

func (responder *ScenarioResponder) MoveToPosition(position int) {
	if position < 0 {
		position = 0
	}

	if position < responder.Playbook.Count {
		log.Infof("moving position from %v to %v", responder.Playbook.Position, position)
		// set the position to the specified location
		responder.Playbook.Position = position
	}
}

func (responder *ScenarioResponder) GetFirst() (PlaybookResponse, error) {
	r, err := responder.getPlaybookRecord(0)
	return r, err
}

func (responder *ScenarioResponder) GetLast() (PlaybookResponse, error) {
	r, err := responder.getPlaybookRecord(responder.Playbook.Count - 1)
	return r, err
}

func (responder *ScenarioResponder) GetCurrent() (PlaybookResponse, error) {
	r, err := responder.getPlaybookRecord(responder.Playbook.Position)
	return r, err
}

func (responder *ScenarioResponder) getPlaybookRecord(i int) (PlaybookResponse, error) {
	var e error

	if len(responder.Playbook.Responses) > 0 {
		r := responder.Playbook.Responses[i]
		r.Position = i
		return r, e
	} else {
		return PlaybookResponse{}, errors.New("empty response file")
	}
}

// GetECUResponse returns an emulated response byte string
func (responder *ScenarioResponder) GetECUResponse(cmd []byte) []byte {
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
		// if a command request is made whilst we're replaying
		// generate a relevant response from the response map
		data = generateECUResponse(command)
	}

	return data
}

// determines where the command code is a dataframe request
func (responder *ScenarioResponder) isDataframeRequest(command string) bool {
	return (command == "80" || command == "7D")
}

// converts the hex string to a byte array
func (responder *ScenarioResponder) convertHexStringToByteArray(response string) []byte {
	// convert to byte array
	data, _ := hex.DecodeString(response)

	return data
}
