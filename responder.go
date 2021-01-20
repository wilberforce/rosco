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
	file        *os.File
	RawData     []*RawData
	Playbook    Playbook
	responseMap map[string][]byte
}

// NewResponder creates an instance of a responder
func NewResponder() *Responder {
	responder := &Responder{}
	responder.responseMap = make(map[string][]byte)

	responder.buildResponseMap()

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

	r := responder.responseMap[command]

	if r == nil {
		r = responder.responseMap["00"]
		copy(r[0:], command)
	}

	log.WithFields(log.Fields{"response": fmt.Sprintf("%x", r), "command": command}).Info("generated response for command")
	return r
}

// build the response map for generated Responses
func (responder *Responder) buildResponseMap() {
	// Response formats for commands that do not respond with the format [COMMAND][VALUE]
	// Generally these are either part of the initialisation sequence or are ECU data frames
	responder.responseMap["0A"] = []byte{0x0A}
	responder.responseMap["CA"] = []byte{0xCA}
	responder.responseMap["75"] = []byte{0x75}

	// Format for DataFrames starts with [Command Echo][Data Size][Data Bytes (28 for 0x80 and 32 for 0x7D)]
	responder.responseMap["80"] = []byte{0x80, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B}
	responder.responseMap["7D"] = []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F}
	responder.responseMap["D0"] = []byte{0xD0, 0x99, 0x00, 0x03, 0x03}

	// heatbeat
	responder.responseMap["F4"] = []byte{0xf4, 0x00}
	responder.responseMap["FB"] = []byte{0xfb, 0x00}

	// adjustments
	responder.responseMap["7A"] = []byte{0x7a, 0x89}
	responder.responseMap["7B"] = []byte{0x7b, 0x1e}
	responder.responseMap["7C"] = []byte{0x7c, 0x8a}
	responder.responseMap["79"] = []byte{0x79, 0x8a}
	responder.responseMap["7B"] = []byte{0x7b, 0x8a}
	responder.responseMap["8A"] = []byte{0x8a, 0x23}
	responder.responseMap["89"] = []byte{0x89, 0x23}
	responder.responseMap["92"] = []byte{0x92, 0x80}
	responder.responseMap["91"] = []byte{0x91, 0x80}
	responder.responseMap["94"] = []byte{0x94, 0x80}
	responder.responseMap["93"] = []byte{0x93, 0x80}

	//resets
	responder.responseMap["FA"] = []byte{0xfa, 0x00}
	responder.responseMap["0F"] = []byte{0x0f, 0x00}
	responder.responseMap["CC"] = []byte{0xcc, 0x00}

	// generic response, expect command and single byte response
	responder.responseMap["00"] = []byte{0x00, 0x00}
}
