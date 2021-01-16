package rosco

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/tarm/serial"
)

// MemsCommandResponse communication pair
type MemsCommandResponse struct {
	Command       []byte   `json:"Command"`
	Response      []byte   `json:"Response"`
	MemsDataFrame MemsData `json:"MemsData"`
}

// MemsConnection communication structure for MEMS
type MemsConnection struct {
	SerialPort      *serial.Port
	CommandResponse *MemsCommandResponse
	Diagnostics     *MemsDiagnostics
	responder       *Responder
	Status          *MemsConnectionStatus
}

// MemsConnectionStatus are we?
type MemsConnectionStatus struct {
	Emulated    bool   `json:"Emulated"`
	Connected   bool   `json:"Connected"`
	Initialised bool   `json:"Initialised"`
	ECUID       string `json:"ECUID"`
	IACPosition int    `json:"IACPosition"`
}

// package init function
func init() {
	// Response formats for commands that do not respond with the format [COMMAND][VALUE]
	// Generally these are either part of the initialisation sequence or are ECU data frames
	responseMap["0a"] = []byte{0x0A}
	responseMap["ca"] = []byte{0xCA}
	responseMap["75"] = []byte{0x75}

	// Format for DataFrames starts with [Command Echo][Data Size][Data Bytes (28 for 0x80 and 32 for 0x7D)]
	responseMap["80"] = []byte{0x80, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B}
	responseMap["7d"] = []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F}
	responseMap["d0"] = []byte{0xD0, 0x99, 0x00, 0x03, 0x03}

	// generic response, expect command and single byte response
	responseMap["0f"] = []byte{0x0f, 0x00}
	responseMap["00"] = []byte{0x00, 0x00}
}

// NewMemsConnection creates a new mems structure
func NewMemsConnection() *MemsConnection {
	m := &MemsConnection{}
	//m.TxECU = make(chan MemsCommandResponse)
	//m.RxECU = make(chan MemsCommandResponse)
	m.CommandResponse = &MemsCommandResponse{}
	// engine diagnostics
	m.Diagnostics = NewMemsDiagnostics()
	// set status
	m.Status = &MemsConnectionStatus{}
	m.Status.Connected = false
	m.Status.Initialised = false
	m.Status.Emulated = false
	m.Status.ECUID = ""
	m.Status.IACPosition = m.Diagnostics.Analysis.IACPosition

	return m
}

// ConnectAndInitialiseECU connect and initialise the ECU
func (mems *MemsConnection) ConnectAndInitialiseECU(port string) {
	if mems.isScenario(port) {
		// emulate ECU if scenario file is supplied
		mems.Status.Emulated = true
		mems.responder = NewResponder()
	}

	if !mems.Status.Connected {
		mems.connect(port)
		if mems.Status.Connected {
			mems.initialise()
		}
	}

	// update status
	mems.Status.IACPosition = mems.Diagnostics.Analysis.IACPosition
}

// Disconnect from the ECU
func (mems *MemsConnection) Disconnect() MemsConnectionStatus {
	// close the connection
	mems.SerialPort.Flush()
	mems.SerialPort.Close()

	// update the status
	mems.Status.Connected = false
	mems.Status.Initialised = false
	mems.Status.Emulated = false
	mems.Status.ECUID = ""
	mems.Status.IACPosition = 0

	return *mems.Status
}

// ResetDiagnostics clears and resets the diagnostic data
func (mems *MemsConnection) ResetDiagnotics() {
	// update the status
	mems.Diagnostics = NewMemsDiagnostics()
}

// GetStatus returns the connection and ECU status
func (mems *MemsConnection) GetStatus() MemsConnectionStatus {
	return *mems.Status
}

// SendCommand sends a command and returns the response
func (mems *MemsConnection) SendCommand(cmd []byte) ([]byte, error) {
	mems.writeSerial(cmd)
	response, e := mems.readSerial()

	if e != nil {
		LogW.Printf("%s command send/receive fault %v", ECUResponseTrace, e)
	}

	return response, e
}

func (mems *MemsConnection) GetDataframes() MemsData {
	LogI.Printf("%s getting x7d and x80 dataframes", ECUCommandTrace)

	// read the raw dataframes
	d80, d7d, e := mems.readRawDataFrames()

	if e != nil {
		LogE.Printf("%s Unable to create memsdata, corrupt dataframes", ECUResponseTrace)
	}

	// populate the DataFrame structure for command 0x80
	r := bytes.NewReader(d80)
	var df80 DataFrame80

	if err := binary.Read(r, binary.BigEndian, &df80); err != nil {
		LogE.Printf("%s dataframe x80 binary.Read failed: %v", ECUCommandTrace, err)
	}

	// populate the DataFrame structure for command 0x7d
	r = bytes.NewReader(d7d)
	var df7d DataFrame7d

	if err := binary.Read(r, binary.BigEndian, &df7d); err != nil {
		LogE.Printf("%s dataframe x7d binary.Read failed: %v", ECUCommandTrace, err)
	}

	t := time.Now()

	// build the Mems Data frame using the raw data and applying the relevant
	// adjustments and calculations
	memsdata := MemsData{
		Time:                     t.Format("15:04:05.000"),
		EngineRPM:                int(df80.EngineRpm),
		CoolantTemp:              int(df80.CoolantTemp) - 55,
		AmbientTemp:              int(df80.AmbientTemp) - 55,
		IntakeAirTemp:            int(df80.IntakeAirTemp) - 55,
		FuelTemp:                 int(df80.FuelTemp) - 55,
		ManifoldAbsolutePressure: float32(df80.ManifoldAbsolutePressure),
		BatteryVoltage:           float32(df80.BatteryVoltage) / 10,
		ThrottlePotSensor:        roundTo2DecimalPoints(float32(df80.ThrottlePotSensor) * 0.02),
		IdleSwitch:               bool(df80.IdleSwitch&IdleSwitchActive != 0),
		AirconSwitch:             bool(df80.AirconSwitch != 0),
		ParkNeutralSwitch:        bool(df80.ParkNeutralSwitch != 0),
		DTC0:                     df80.Dtc0,
		DTC1:                     df80.Dtc1,
		IdleSetPoint:             int(df80.IdleSetPoint),
		IdleHot:                  int(df80.IdleHot) - 35,
		IACPosition:              int(df80.IacPosition),
		IdleSpeedDeviation:       int(df80.IdleSpeedDeviation),
		IgnitionAdvanceOffset80:  int(df80.IgnitionAdvanceOffset80),
		IgnitionAdvance:          (float32(df80.IgnitionAdvance) / 2) - 24,
		CoilTime:                 roundTo2DecimalPoints(float32(df80.CoilTime) * 0.002),
		CrankshaftPositionSensor: bool(df80.CrankshaftPositionSensor != 0),
		CoolantTempSensorFault:   bool(df80.Dtc0&CoolantSensorFaultCode != 0),
		IntakeAirTempSensorFault: bool(df80.Dtc0&AirSensorFaultCode != 0),
		FuelPumpCircuitFault:     bool(df80.Dtc1&FuelPumpFaultCode != 0),
		ThrottlePotCircuitFault:  bool(df80.Dtc1&ThrottlePotFaultCode != 0),
		IgnitionSwitch:           bool(df7d.IgnitionSwitch != 0),
		ThrottleAngle:            int(math.Round(float64(df7d.ThrottleAngle) * 6 / 10)),
		AirFuelRatio:             float32(df7d.AirFuelRatio) / 10,
		DTC2:                     df7d.Dtc2,
		LambdaVoltage:            int(df7d.LambdaVoltage) * 5,
		LambdaFrequency:          int(df7d.LambdaFrequency),
		LambdaDutycycle:          int(df7d.LambdaDutyCycle),
		LambdaStatus:             int(df7d.LambdaStatus),
		ClosedLoop:               bool(df7d.LoopIndicator != 0),
		LongTermFuelTrim:         int(df7d.LongTermFuelTrim) - 128,
		ShortTermFuelTrim:        int(df7d.ShortTermFuelTrim),
		FuelTrimCorrection:       int(df7d.ShortTermFuelTrim) - 100,
		CarbonCanisterPurgeValve: int(df7d.CarbonCanisterPurgeValve),
		DTC3:                     df7d.Dtc3,
		IdleBasePosition:         int(df7d.IdleBasePos),
		DTC4:                     df7d.Dtc4,
		IgnitionAdvanceOffset7d:  int(df7d.IgnitionAdvanceOffset7d) - 48,
		IdleSpeedOffset:          (int(df7d.IdleSpeedOffset) - 128) * 25,
		DTC5:                     df7d.Dtc5,
		JackCount:                int(df7d.JackCount),
		Dataframe80:              hex.EncodeToString(d80),
		Dataframe7d:              hex.EncodeToString(d7d),
	}

	// add the data for diagnostics
	mems.Diagnostics.Add(memsdata)
	mems.Diagnostics.Analyse()

	LogI.Printf("%s built mems dataframe %v", ECUCommandTrace, memsdata)

	return memsdata
}

//
// Private functions
//

// connect to MEMS via serial port
func (mems *MemsConnection) connect(port string) {
	var err error
	var s *serial.Port

	// assume not connected or initialised
	mems.Status.Connected = false
	mems.Status.Initialised = false

	if mems.Status.Emulated {
		err = mems.responder.LoadScenario(port)
	} else {
		// connect to the ecu, timeout if we don't get data after a couple of seconds
		c := &serial.Config{Name: port, Baud: 9600, ReadTimeout: time.Millisecond * 2000}

		LogI.Println("opening ", port)

		s, err = serial.OpenPort(c)
		if s != nil {
			mems.SerialPort = s
		}
	}

	if err == nil {
		LogI.Println("connected to ", port)
		mems.Status.Connected = true
	} else {
		LogE.Printf("error opening port (%s)", err)
		mems.Status.Connected = false
		mems.Status.Initialised = false
	}
}

// check if the port is a CSV file, if so then a scenario emulation
// has been requested rather than a real serial connection
func (mems *MemsConnection) isScenario(port string) bool {
	return strings.HasSuffix(port, ".csv")
}

// checks the first byte of the response against the sent command
func (mems *MemsConnection) isCommandEcho() bool {
	return mems.CommandResponse.Command[0] == mems.CommandResponse.Response[0]
}

// initialises the connection to the ECU
// The initialisation sequence is as follows:
//
// 1. Send command CA (MEMS_InitCommandA)
// 2. Recieve response CA
// 3. Send command 75 (MEMS_InitCommandB)
// 4. Recieve response 75
// 5. Send request ECU ID command D0 (MEMS_InitECUID)
// 6. Recieve response D0 XX XX XX XX
//
func (mems *MemsConnection) initialise() {
	// assume not initialised
	mems.Status.Initialised = false

	if mems.Status.Emulated {
		mems.Status.Initialised = true
	} else {
		if mems.SerialPort != nil {
			mems.SerialPort.Flush()

			mems.writeSerial(MEMSInitCommandA)
			_, _ = mems.readSerial()

			mems.writeSerial(MEMSInitCommandB)
			_, _ = mems.readSerial()

			mems.writeSerial(MEMSHeartbeat)
			_, _ = mems.readSerial()

			mems.writeSerial(MEMSInitECUID)
			ECUID, _ := mems.readSerial()
			mems.Status.ECUID = fmt.Sprintf("%X", ECUID)

			// get the IAC position
			mems.writeSerial(MEMSGetIACPosition)
			response, _ := mems.readSerial()
			iac, _ := binary.Uvarint(response)
			mems.Diagnostics.Analysis.IACPosition = int(iac)

			mems.Status.Initialised = true
		}
	}
}

// readSerial read from MEMS
// read 1 byte at a time until we have all the expected bytes
func (mems *MemsConnection) readSerial() ([]byte, error) {
	var n int
	var e error

	size := mems.getResponseSize(mems.CommandResponse.Command)

	// serial read buffer
	b := make([]byte, size)

	//  data frame buffer
	data := make([]byte, 0)

	if mems.Status.Emulated {
		// emulate the response
		data = mems.responder.GetECUResponse(mems.CommandResponse.Command)
		LogI.Printf("%s data read for emulation %x", EmulatorTrace, data)
	} else {
		if mems.SerialPort != nil {
			// read all the expected bytes before returning the data
			for count := 0; count < size; {
				// wait for a response from MEMS
				n, _ = mems.SerialPort.Read(b)

				if n == 0 {
					LogW.Printf("serial port read error, timeout?")
					// drop out of loop, send back a 0x00 byte array response
					// this prevents the loop getting blocked on a read error
					count = size
					data = append(data, b...)
					e = errors.New("serial port read error")
				} else {
					// append the read bytes to the data frame
					data = append(data, b[:n]...)
				}

				// increment by the number of bytes read
				count = count + n
				if count > size {
					LogW.Printf("%s dataframe size mismatch (received %d, expected %d)", ECUResponseTrace, count, size)
					e = errors.New("size mismatch")
				}
			}
		}
	}

	LogI.Printf("%s received data from ECU [%d] < %x", ECUResponseTrace, n, data)
	mems.CommandResponse.Response = data

	if !mems.isCommandEcho() {
		LogW.Printf("%s expecting command echo (%x)\n", ECUResponseTrace, mems.CommandResponse.Command)
		e = errors.New("command mismatch")
	}

	return data, e
}

// writeSerial write to MEMS
func (mems *MemsConnection) writeSerial(data []byte) {
	if mems.Status.Emulated {
		LogI.Printf("%s data stored for emulation %x", EmulatorTrace, data)
		mems.CommandResponse.Command = data
	} else {
		if mems.SerialPort != nil {
			// save the sent command
			mems.CommandResponse.Command = data

			// write the response to the code reader
			n, e := mems.SerialPort.Write(data)

			if e != nil {
				LogE.Printf("%s error sending data to serial port (%s)", ECUCommandTrace, e)
			}

			if n > 0 {
				LogI.Printf("%s data sent to serial port %x", ECUCommandTrace, data)
			}
		}
	}
}

func roundTo2DecimalPoints(x float32) float32 {
	return float32(math.Round(float64(x)*100) / 100)
}

// readRawDataFrames reads dataframe 80 and then dataframe 7d as raw byte arrays
func (mems *MemsConnection) readRawDataFrames() ([]byte, []byte, error) {
	mems.writeSerial(MEMSReqData80)
	dataframe80, e := mems.readSerial()

	if e != nil {
		LogW.Printf("%s dataframe80 command send/receive fault %v", ECUResponseTrace, e)
	}

	mems.writeSerial(MEMSReqData7D)
	dataframe7d, e := mems.readSerial()

	if e != nil {
		LogW.Printf("%s dataframe7d command send/receive fault %v", ECUResponseTrace, e)
	}

	return dataframe80, dataframe7d, e
}

// getResponseSize returns the expected number of bytes for a given command
// The 'response' variable contains the formats for each command response pattern
// by default the response size is 2 bytes unless the command has a special format.
func (mems *MemsConnection) getResponseSize(command []byte) int {
	size := 2

	c := hex.EncodeToString(command)
	r := responseMap[c]

	if r != nil {
		size = len(r)
	} else {
		r = responseMap["00"]
		copy(r[0:], command)
	}

	LogI.Printf("%s expecting %x (%d)", ECUResponseTrace, r, size)
	return size
}
