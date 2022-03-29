package rosco

import (
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"strings"
)

type ECUStatus struct {
	Connected   bool   `json:"Connected"`
	ECUID       string `json:"ECUID"`
	ECUSerial   string `json:"ECUSerial"`
	IACPosition int    `json:"IACPosition"`
}

type ECUReader interface {
	Connect() (connected bool, err error)
	SendAndReceive(command []byte) (response []byte, err error)
	Disconnect() (err error)
}

// global response map
var responseMap = make(map[string][]byte)

// ECU Reader factory
func NewECUReader(connection string) ECUReader {
	var r ECUReader

	// prepare the response map for synthetic ECUs
	responseMap = createResponseMap()

	// determine the type of reader from the connection string
	isFile := strings.HasSuffix(connection, ".csv") || strings.HasSuffix(connection, ".fcr")
	isLoopback := strings.Contains(connection, "loopback")

	if isLoopback {
		r = NewLoopbackReader()
	}

	if isFile {
		r = NewScenarioReader(connection)
	}

	// default to a Mems Reader
	if r == nil {
		r = NewMEMSReader(connection)
	}

	return r
}

// getResponseSize returns the expected number of bytes for a given command
// The 'response' variable contains the formats for each command response pattern
// by default the response size is 2 bytes unless the command has a special format.
func getResponseSize(command []byte) (int, error) {
	var err error
	var size int

	c := hex.EncodeToString(command)
	c = strings.ToUpper(c)
	response := responseMap[c]

	if response == nil {
		// default response size of 2 bytes, usually command and
		size = 2
		err = fmt.Errorf("unable to find response for command %s", command)
		log.Warnf("%s", err)
	} else {
		size = len(response)
		log.Infof("mapped command %s to %X, expecting a response of %d bytes", c, response, size)
	}

	return size, err
}

// if we're responding to a command that isn't a dataframe request
// then generate the correct response
func generateECUResponse(command string) []byte {
	if len(responseMap) == 0 {
		// create the response map, if it's not already been initialised
		responseMap = createResponseMap()
	}

	command = strings.ToUpper(command)
	response := responseMap[command]

	if response == nil {
		response = responseMap["00"]
		copy(response[0:], command)
	}

	log.Infof("generated a response %X for command %X", response, command)

	return response
}

func createResponseMap() map[string][]byte {
	responseMap = make(map[string][]byte)
	// Response formats for commands that do not respond with the format [COMMAND][VALUE]
	// Generally these are either part of the initialisation sequence or are ECU data frames
	responseMap["0A"] = []byte{0x0A}
	responseMap["CA"] = []byte{0xCA}
	responseMap["75"] = []byte{0x75}

	// Format for DataFrames starts with [Command Echo][Data Size][Data Bytes (28 for 0x80 and 32 for 0x7D)]
	responseMap["80"] = []byte{0x80, 0x1c, 0x04, 0xa5, 0x4b, 0xff, 0x4c, 0xff, 0x31, 0x82, 0x22, 0x00, 0x20, 0x01, 0x00, 0x00, 0x00, 0x20, 0x84, 0x78, 0x00, 0x1d, 0x00, 0x44, 0x06, 0x59, 0x10, 0x00, 0x00}
	responseMap["7D"] = []byte{0x7d, 0x20, 0x10, 0x14, 0xff, 0x92, 0x40, 0x57, 0xff, 0xff, 0x01, 0x00, 0x80, 0x64, 0x00, 0xff, 0x64, 0xff, 0xff, 0x30, 0x80, 0x80, 0x0e, 0xff, 0x16, 0x80, 0x1b, 0x00, 0x22, 0x00, 0x31, 0xc0, 0x1f}

	responseMap["D0"] = []byte{0xD0, 0x99, 0x00, 0x03, 0x03}
	responseMap["D1"] = []byte{0xD1, 0x41, 0x42, 0x4E, 0x4D, 0x50, 0x30, 0x30, 0x33, 0x99, 0x00, 0x03, 0x03}

	// heartbeat
	responseMap["F4"] = []byte{0xf4, 0x00}

	// adjustments
	responseMap["79"] = []byte{0x79, 0x8b} // increment STFT (default is 138)
	responseMap["7A"] = []byte{0x7a, 0x89} // decrement STFT (default is 138)
	responseMap["7B"] = []byte{0x7b, 0x1f} // increment LTFT (default is 30)
	responseMap["7C"] = []byte{0x7c, 0x1d} // decrement LTFT (default is 30)
	responseMap["89"] = []byte{0x89, 0x24} // increment Idle Decay (default is 35)
	responseMap["8A"] = []byte{0x8a, 0x22} // decrement Idle Decay (default is 35)
	responseMap["91"] = []byte{0x91, 0x81} // increment Idle Speed  (default is 128)
	responseMap["92"] = []byte{0x92, 0x7f} // decrement Idle Speed (default is 128)
	responseMap["93"] = []byte{0x93, 0x81} // increment Ignition Advance Offset (default is 128)
	responseMap["94"] = []byte{0x94, 0x7f} // decrement Ignition Advance Offset (default is 128)
	responseMap["FD"] = []byte{0xfd, 0x81} // increment IAC (default is 128)
	responseMap["FE"] = []byte{0xfe, 0x7f} // decrement IAC (default is 128)

	//resets
	responseMap["0F"] = []byte{0x0f, 0x00} // clear all adjustments
	responseMap["CC"] = []byte{0xcc, 0x00} // clear faults
	responseMap["FA"] = []byte{0xfa, 0x00} // reset ecu
	responseMap["FB"] = []byte{0xfb, 0x80} // Idle Air Control Position

	// actuators
	responseMap["11"] = []byte{0x11, 0x00} // fuel pump on
	responseMap["01"] = []byte{0x01, 0x00} // fuel pump off
	responseMap["12"] = []byte{0x12, 0x00} // ptc relay on
	responseMap["02"] = []byte{0x02, 0x00} // ptc relay off
	responseMap["13"] = []byte{0x13, 0x00} // ac relay on
	responseMap["03"] = []byte{0x03, 0x00} // ac relay off
	responseMap["18"] = []byte{0x18, 0x00} // purge valve on
	responseMap["08"] = []byte{0x08, 0x00} // purge vavle off
	responseMap["19"] = []byte{0x19, 0x00} // O2 heater on
	responseMap["09"] = []byte{0x09, 0x00} // O2 heater off
	responseMap["1B"] = []byte{0x1b, 0x00} // boost valve on
	responseMap["0B"] = []byte{0x0b, 0x00} // boost valve off
	responseMap["1D"] = []byte{0x1d}       // fan 1 on
	responseMap["0D"] = []byte{0x0d, 0x00} // fan 1 off
	responseMap["1E"] = []byte{0x1e}       // fan 2 on
	responseMap["0E"] = []byte{0x0e, 0x00} // fan 2 off
	responseMap["EF"] = []byte{0xef, 0x03} // test mpi injectors
	responseMap["F7"] = []byte{0xf7, 0x03} // test injectors
	responseMap["F8"] = []byte{0xf8, 0x02} // fire coil

	// unknown command Responses
	responseMap["65"] = []byte{0x65, 0x00}
	responseMap["6D"] = []byte{0x6d, 0x00}
	responseMap["7E"] = []byte{0x7e, 0x08}
	responseMap["7F"] = []byte{0x7f, 0x05}
	responseMap["82"] = []byte{0x82, 0x09, 0x9E, 0x1D, 0x00, 0x00, 0x60, 0x05, 0xFF, 0xFF}
	responseMap["CB"] = []byte{0xcb, 0x00}
	responseMap["CD"] = []byte{0xcd, 0x01}
	responseMap["D2"] = []byte{0xd2, 0x02, 0x01, 0x00, 0x01}
	responseMap["D3"] = []byte{0xd3, 0x02, 0x01, 0x00, 0x02}
	responseMap["E7"] = []byte{0xe7, 0x02}
	responseMap["E8"] = []byte{0xe8, 0x05, 0x26, 0x01, 0x00, 0x01}
	responseMap["ED"] = []byte{0xed, 0x00}
	responseMap["EE"] = []byte{0xee, 0x00}
	responseMap["F0"] = []byte{0xf0, 0x05}
	responseMap["F3"] = []byte{0xf3, 0x00}
	responseMap["F5"] = []byte{0xf5, 0x00}
	responseMap["F6"] = []byte{0xf6, 0x00}
	responseMap["FC"] = []byte{0xfc, 0x00}

	// generic response, expect command and single byte response
	responseMap["00"] = []byte{0x00, 0x00}

	return responseMap
}

func roundTo2DecimalPoints(x float32) float32 {
	return float32(math.Round(float64(x)*100) / 100)
}
