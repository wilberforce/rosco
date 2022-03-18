package rosco

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type LoopbackReader struct {
	responseMap map[string][]byte
}

func NewLoopbackReader() *LoopbackReader {
	r := &LoopbackReader{}
	r.createResponseMap()
	return r
}

func (r *LoopbackReader) Open(connection string) error {
	var err error

	if connection != "loopback" {
		return fmt.Errorf("Cannot open %s as a loppback", connection)
	}

	return err
}

func (r *LoopbackReader) Read(b []byte) (int, error) {
	var err error
	var n int

	// convert the command code to a string
	command := hex.EncodeToString(b)
	command = strings.ToUpper(command)

	return n, err
}

func (r *LoopbackReader) Write(b []byte) (int, error) {
	var err error
	var n int

	return n, err
}

func (r *LoopbackReader) Flush() {
}

func (r *LoopbackReader) Close() error {
	var err error
	return err
}

func (r *LoopbackReader) createResponseMap() {
	r.responseMap = make(map[string][]byte)
	// Response formats for commands that do not respond with the format [COMMAND][VALUE]
	// Generally these are either part of the initialisation sequence or are ECU data frames
	r.responseMap["0A"] = []byte{0x0A}
	r.responseMap["CA"] = []byte{0xCA}
	r.responseMap["75"] = []byte{0x75}

	// Format for DataFrames starts with [Command Echo][Data Size][Data Bytes (28 for 0x80 and 32 for 0x7D)]
	r.responseMap["80"] = []byte{0x80, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B}
	r.responseMap["7D"] = []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F}
	r.responseMap["D0"] = []byte{0xD0, 0x99, 0x00, 0x03, 0x03}
	r.responseMap["D1"] = []byte{0xD1, 0x41, 0x42, 0x4E, 0x4D, 0x50, 0x30, 0x30, 0x33, 0x99, 0x00, 0x03, 0x03}

	// heartbeat
	r.responseMap["F4"] = []byte{0xf4, 0x00}

	// adjustments
	r.responseMap["79"] = []byte{0x79, 0x8b} // increment STFT (default is 138)
	r.responseMap["7A"] = []byte{0x7a, 0x89} // decrement STFT (default is 138)
	r.responseMap["7B"] = []byte{0x7b, 0x1f} // increment LTFT (default is 30)
	r.responseMap["7C"] = []byte{0x7c, 0x1d} // decrement LTFT (default is 30)
	r.responseMap["89"] = []byte{0x89, 0x24} // increment Idle Decay (default is 35)
	r.responseMap["8A"] = []byte{0x8a, 0x22} // decrement Idle Decay (default is 35)
	r.responseMap["91"] = []byte{0x91, 0x81} // increment Idle Speed  (default is 128)
	r.responseMap["92"] = []byte{0x92, 0x7f} // decrement Idle Speed (default is 128)
	r.responseMap["93"] = []byte{0x93, 0x81} // increment Ignition Advance Offset (default is 128)
	r.responseMap["94"] = []byte{0x94, 0x7f} // decrement Ignition Advance Offset (default is 128)
	r.responseMap["FD"] = []byte{0xfd, 0x81} // increment IAC (default is 128)
	r.responseMap["FE"] = []byte{0xfe, 0x7f} // decrement IAC (default is 128)

	//resets
	r.responseMap["0F"] = []byte{0x0f, 0x00} // clear all adjustments
	r.responseMap["CC"] = []byte{0xcc, 0x00} // clear faults
	r.responseMap["FA"] = []byte{0xfa, 0x00} // reset ecu
	r.responseMap["FB"] = []byte{0xfb, 0x80} // Idle Air Control Position

	// actuators
	r.responseMap["11"] = []byte{0x11, 0x00} // fuel pump on
	r.responseMap["01"] = []byte{0x01, 0x00} // fuel pump off
	r.responseMap["12"] = []byte{0x12, 0x00} // ptc relay on
	r.responseMap["02"] = []byte{0x02, 0x00} // ptc relay off
	r.responseMap["13"] = []byte{0x13, 0x00} // ac relay on
	r.responseMap["03"] = []byte{0x03, 0x00} // ac relay off
	r.responseMap["18"] = []byte{0x18, 0x00} // purge valve on
	r.responseMap["08"] = []byte{0x08, 0x00} // purge vavle off
	r.responseMap["19"] = []byte{0x19, 0x00} // O2 heater on
	r.responseMap["09"] = []byte{0x09, 0x00} // O2 heater off
	r.responseMap["1B"] = []byte{0x1b, 0x00} // boost valve on
	r.responseMap["0B"] = []byte{0x0b, 0x00} // boost valve off
	r.responseMap["1D"] = []byte{0x1d}       // fan 1 on
	r.responseMap["0D"] = []byte{0x0d, 0x00} // fan 1 off
	r.responseMap["1E"] = []byte{0x1e}       // fan 2 on
	r.responseMap["0E"] = []byte{0x0e, 0x00} // fan 2 off
	r.responseMap["EF"] = []byte{0xef, 0x03} // test mpi injectors
	r.responseMap["F7"] = []byte{0xf7, 0x03} // test injectors
	r.responseMap["F8"] = []byte{0xf8, 0x02} // fire coil

	// unknown command Responses
	r.responseMap["65"] = []byte{0x65, 0x00}
	r.responseMap["6D"] = []byte{0x6d, 0x00}
	r.responseMap["7E"] = []byte{0x7e, 0x08}
	r.responseMap["7F"] = []byte{0x7f, 0x05}
	r.responseMap["82"] = []byte{0x82, 0x09, 0x9E, 0x1D, 0x00, 0x00, 0x60, 0x05, 0xFF, 0xFF}
	r.responseMap["CB"] = []byte{0xcb, 0x00}
	r.responseMap["CD"] = []byte{0xcd, 0x01}
	r.responseMap["D2"] = []byte{0xd2, 0x02, 0x01, 0x00, 0x01}
	r.responseMap["D3"] = []byte{0xd3, 0x02, 0x01, 0x00, 0x02}
	r.responseMap["E7"] = []byte{0xe7, 0x02}
	r.responseMap["E8"] = []byte{0xe8, 0x05, 0x26, 0x01, 0x00, 0x01}
	r.responseMap["ED"] = []byte{0xed, 0x00}
	r.responseMap["EE"] = []byte{0xee, 0x00}
	r.responseMap["F0"] = []byte{0xf0, 0x05}
	r.responseMap["F3"] = []byte{0xf3, 0x00}
	r.responseMap["F5"] = []byte{0xf5, 0x00}
	r.responseMap["F6"] = []byte{0xf6, 0x00}
	r.responseMap["FC"] = []byte{0xfc, 0x00}

	// generic response, expect command and single byte response
	r.responseMap["00"] = []byte{0x00, 0x00}
}
