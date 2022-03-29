package rosco

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"math"
	"reflect"
	"time"
)

// ECUReaderInstance communication structure for MEMS
type ECUReaderInstance struct {
	ecuReader   ECUReader
	Status      *ECUStatus
	Diagnostics *DataframeAnalysis
	Datalogger  *MemsDataLogger
}

/*
// MemsConnectionStatus are we?
type MemsConnectionStatus struct {
	Emulated    bool   `json:"Emulated"`
	Connected   bool   `json:"Connected"`
	Initialised bool   `json:"Initialised"`
	ECUSerial   string `json:"ECUSerial"`
	ECUID       string `json:"ECUID"`
	IACPosition int    `json:"IACPosition"`
}
*/
// NewECUReaderInstance creates a new mems structure
func NewECUReaderInstance() *ECUReaderInstance {
	m := &ECUReaderInstance{}
	m.Status = &ECUStatus{}
	m.Diagnostics = NewDataframeAnalysis(20)
	m.resetStatus()

	return m
}

func (ecu *ECUReaderInstance) ConnectAndInitialiseECU(port string) (bool, error) {
	var err error
	var connected bool

	ecu.ecuReader = NewECUReader(port)

	if connected, err = ecu.connectToECU(); err == nil {
		if connected {
			ecu.Status.Connected = true
			// get the ecu id, serial number and iac position
			ecu.Status.ECUID, err = ecu.getECUID()
			ecu.Status.ECUSerial, err = ecu.getECUSerial()
			ecu.Status.IACPosition, err = ecu.GetIACPosition()
		}
	}

	return ecu.Status.Connected, err
}

func (ecu *ECUReaderInstance) Disconnect() error {
	var err error

	if err = ecu.ecuReader.Disconnect(); err == nil {
		log.Info("disconnected ecu")
	} else {
		log.Warnf("error disconnecting (%s)", err)
	}

	ecu.resetStatus()

	return err
}

// ResetDiagnostics clears and resets the diagnostic data
func (ecu *ECUReaderInstance) ResetDiagnostics() {
	// update the status
	log.Info("resetting ecu diagnostics")
	ecu.Diagnostics = NewDataframeAnalysis(20)
}

func (ecu *ECUReaderInstance) GetDataframes() MemsData {
	df := MemsData{}

	// read the raw dataframes
	log.Info("getting 0x7d and 0x80 dataframes")
	d80, d7d := ecu.readRawDataFrames()

	// create the dataframes from the raw binary df
	if df80, err := ecu.createDataframe80(d80); err == nil {
		if df7d, err := ecu.createDataframe7D(d7d); err == nil {
			// build the Mems Dataframe using the raw df and applying the relevant adjustments and calculations
			df = ecu.createMemsDataframe(df80, df7d)
			// include the raw df converted into string format
			df.Dataframe80 = hex.EncodeToString(d80)
			df.Dataframe7d = hex.EncodeToString(d7d)

			log.Infof("generated ecu df from dataframe (%+v)", df)

			ecu.Diagnostics.Analyse(df)
			df.Analytics = ecu.Diagnostics.Analysis
		}
	}

	ecu.writeToLog(df)

	return df

}

func (ecu *ECUReaderInstance) connectToECU() (bool, error) {
	return ecu.ecuReader.Connect()
}

func (ecu *ECUReaderInstance) createMemsDataframe(df80 DataFrame80, df7d DataFrame7d) MemsData {
	t := time.Now()

	memsdata := MemsData{
		Time:                     t.Format("2006-01-02 15:04:05.000"),
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
		IdleHot:                  int(df80.IdleHot), // was (hotidle - 35) but don't understand why this offset
		IACPosition:              int(df80.IacPosition),
		IdleSpeedDeviation:       int(df80.IdleSpeedDeviation),
		IgnitionAdvanceOffset80:  int(df80.IgnitionAdvanceOffset80),
		IgnitionAdvance:          (float32(df80.IgnitionAdvance) / 2) - 24,
		CoilTime:                 roundTo2DecimalPoints(float32(df80.CoilTime) * 0.002),
		CrankshaftPositionSensor: int(df80.CrankshaftPositionSensor),
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
		ShortTermFuelTrim:        int(df7d.ShortTermFuelTrim) - 100,
		FuelTrimCorrection:       int(df7d.ShortTermFuelTrim) - 100,
		CarbonCanisterPurgeValve: int(df7d.CarbonCanisterPurgeValve),
		DTC3:                     df7d.Dtc3,
		IdleBasePosition:         int(df7d.IdleBasePos),
		DTC4:                     df7d.Dtc4,
		IgnitionAdvanceOffset7d:  int(df7d.IgnitionAdvanceOffset7d) - 48,
		IdleSpeedOffset:          int(df7d.IdleSpeedOffset), // - 128) * 25,
		DTC5:                     df7d.Dtc5,
		JackCount:                int(df7d.JackCount),
	}

	return memsdata
}

func (ecu *ECUReaderInstance) createDataframe7D(d7d []byte) (DataFrame7d, error) {
	var err error
	var df7d DataFrame7d

	defer func() {
		if err := recover(); err != nil {
			log.Warnf("dataframe conversion panic occurred %s", err)
		}
	}()

	// populate the DataFrame structure for command 0x7d
	byteReader := bytes.NewReader(d7d)

	if err = binary.Read(byteReader, binary.BigEndian, &df7d); err != nil {
		log.Errorf("error reading dataframe x7d (%s)", err)
	} else {
		log.Infof("dataframe x7d received (data: %X dataframe: %+v)", byteReader, df7d)
	}

	return df7d, err
}

func (ecu *ECUReaderInstance) createDataframe80(d80 []byte) (DataFrame80, error) {
	var err error
	var df80 DataFrame80

	defer func() {
		if err := recover(); err != nil {
			log.Warnf("dataframe conversion panic occurred %s", err)
		}
	}()

	// populate the DataFrame structure for command 0x80
	byteReader := bytes.NewReader(d80)

	if err = binary.Read(byteReader, binary.BigEndian, &df80); err != nil {
		log.Errorf("error reading dataframe x80 (%s)", err)
	} else {
		log.Infof("dataframe x80 received (data: %X dataframe: %+v)", byteReader, df80)
	}

	return df80, err
}

func (ecu *ECUReaderInstance) readRawDataFrames() ([]byte, []byte) {
	var err error
	var dataframe7d, dataframe80 []byte

	if dataframe80, err = ecu.ecuReader.SendAndReceive(MEMSReqData80); err != nil {
		log.Errorf("error recieving dataframe 0x80 (%s)", err)
	}
	if dataframe7d, err = ecu.ecuReader.SendAndReceive(MEMSReqData7D); err != nil {
		log.Errorf("error recieving dataframe 0x7D (%s)", err)
	}

	return dataframe80, dataframe7d
}

func (ecu *ECUReaderInstance) writeToLog(df MemsData) {
	if ecu.Datalogger != nil {
		if reflect.TypeOf(ecu.ecuReader) == reflect.TypeOf(&MEMSReader{}) {
			// write to a logfile if the ecu reader is a real (or virtual) ECU
			go ecu.Datalogger.WriteMemsDataToFile(df)
		}
	}
}

/*
// ConnectAndInitialiseECU connect and initialise the ECU
func (mems *ECUReaderInstance) ConnectAndInitialiseECU(serialPort string) {
	log.Infof("connecting to %s and initialising ecu", serialPort)

	if mems.isScenario(serialPort) {
		log.Info("ecu connected and initialised in emulation mode")
		// emulate ECU if scenario file is supplied
		mems.Status.Emulated = true

		// expand to full path
		serialPort = fmt.Sprintf("%s/%s", GetLogFolder(), serialPort)
		serialPort = filepath.FromSlash(serialPort)

		log.Infof("loading scenario file %s", serialPort)

		mems.Responder = NewResponder()
	}

	if !mems.Status.Connected {

		mems.connect(serialPort)

		if mems.Status.Connected {

			mems.initialise()

			if mems.Status.Initialised {
				log.Info("ecu connected and initialised successfully")
				// update status
				mems.Status.IACPosition = mems.Diagnostics.Analysis.IACPosition

				if !mems.Status.Emulated {
					// create a data log file
					mems.Datalogger = NewMemsDataLogger(GetLogFolder(), mems.Status.ECUID)
				}
			}
		}
	}
}

// sendCommandAndWaitResponse sends a command and returns the response
func (mems *ECUReaderInstance) sendCommandAndWaitResponse(cmd []byte) []byte {
	var response []byte

	mems.writeSerial(cmd)
	response = mems.readSerial()

	mems.CommandResponse.Command = cmd
	mems.CommandResponse.Response = response

	return response
}

// GetIACPosition returns the current IAC Position
func (mems *ECUReaderInstance) GetIACPosition() int {
	data, _ := mems.GetIACPosition()
	return data
}

// GetECUSerial returns the current ECU Serial and ID
func (mems *ECUReaderInstance) GetECUSerial() string {
	data, _ := mems.getECUSerial()
	return data
}

// connect to MEMS via serial serialPort
func (mems *ECUReaderInstance) connect(port string) {
	var err error
	var s *serial.Port

	// assume not connected or initialised
	mems.Status.Connected = false
	mems.Status.Initialised = false

	if mems.Status.Emulated {
		err = mems.Responder.LoadScenario(port)
	} else {
		// connect to the ecu, timeout if we don't get data after a couple of seconds
		c := &serial.Config{Name: port, Baud: 9600, ReadTimeout: time.Millisecond * 2000}

		log.Infof("attempting to open serial serialPort %s", port)
		s, err = serial.OpenPort(c)
	}

	if err != nil {
		log.Errorf("error opening serial serialPort (%s) status : (%+v)", err, mems.Status)
		mems.Status.Connected = false
		mems.Status.Initialised = false
	} else {
		mems.SerialPort = s
		mems.Status.Connected = true
		mems.Status.Initialised = false
		log.Errorf("opened serial serialPort %s (%+v)", port, mems.Status)
	}
}

// check if the serialPort is a CSV file, if so then a scenario emulation
// has been requested rather than a real serial connection
func (mems *ECUReaderInstance) isScenario(port string) bool {
	// return true if the URL starts file://
	if strings.HasPrefix(port, "file://") {
		return true
	}
	// return true if the file extension is CSV
	return strings.HasSuffix(port, ".csv")
}

// checks the first byte of the response against the sent command
func (mems *ECUReaderInstance) isCommandEcho() bool {
	if mems.CommandResponse.Response != nil {
		if len(mems.CommandResponse.Response) > 0 {
			return mems.CommandResponse.Command[0] == mems.CommandResponse.Response[0]
		}
	}

	return false
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
func (mems *ECUReaderInstance) initialise() {
	// assume not initialised
	mems.Status.Initialised = false

	if mems.Status.Emulated {
		mems.Status.Initialised = true
	} else {
		if mems.Status.Connected {
			_ = mems.SerialPort.Flush()

			mems.writeSerial(MEMSInitCommandA)
			response := mems.readSerial()
			log.Infof("init resp %+v", response)

			// if we get the command echoed back we can assume
			// a good connection and proceed. This is to work around the issue
			// in Windows where the serialPort always connects even if it's not available.
			if response[0] == MEMSInitCommandA[0] {
				mems.writeSerial(MEMSInitCommandB)
				_ = mems.readSerial()

				mems.writeSerial(MEMSHeartbeat)
				_ = mems.readSerial()

				mems.writeSerial(MEMSInitECUID)
				ECUID := mems.readSerial()
				mems.Status.ECUID = fmt.Sprintf("%X", ECUID)

				// get the IAC Position
				mems.writeSerial(MEMSGetIACPosition)
				response := mems.readSerial()
				iac, _ := binary.Uvarint(response)
				mems.Diagnostics.Analysis.IACPosition = int(iac)
				log.Infof("IAC Position %d", iac)

				// get the ECU Serial number
				mems.Status.ECUSerial = mems.GetECUSerial()
				mems.Status.Initialised = true
			} else {
				log.Error("timed out on initialisation sequence, closing connection")
				_ = mems.SerialPort.Flush()
				_ = mems.SerialPort.Close()
				mems.Status.Connected = false
				mems.Status.Initialised = false
			}
		}
	}

	log.WithFields(log.Fields{"connected": mems.Status.Connected, "initialised": mems.Status.Initialised}).Info("connected and initialised ECU")
}

// readSerial read from MEMS
// read 1 byte at a time until we have all the expected bytes
func (mems *ECUReaderInstance) readSerial() []byte {
	var n int
	var e error

	size := mems.getResponseSize(mems.CommandResponse.Command)

	// serial read buffer
	b := make([]byte, size)

	//  data frame buffer
	data := make([]byte, 0)

	if mems.Status.Emulated {
		// emulate the response
		data = mems.Responder.GetECUResponse(mems.CommandResponse.Command)
		log.Infof("read scenario data (%+v), %d bytes", data, n)
	} else {
		if mems.Status.Connected {
			if mems.SerialPort != nil {
				// read all the expected bytes before returning the data
				for count := 0; count < size; {
					// wait for a response from MEMS
					n, e = mems.SerialPort.Read(b)

					if n == 0 {
						log.Errorf("0 bytes received, serial serialPort read error, timeout? (%s)", e)
						// drop out of loop, send back a 0x00 byte array response
						// this prevents the loop getting blocked on a read error
						count = size
						data = append(data, b...)
					} else {
						// append the read bytes to the data frame
						data = append(data, b[:n]...)
					}

					// increment by the number of bytes read
					count = count + n
					if count > size {
						log.Errorf("received dataframe size mismatch, received %d, expected %d", count, size)
					}
				}
			}
		}
	}

	log.Infof("received data from ecu (%x), %d bytes", data, n)
	mems.CommandResponse.Response = data

	if !mems.isCommandEcho() {
		log.Warnf("expecting command echo of %x, recevied %x", mems.CommandResponse.Response, mems.CommandResponse.Command)
	}

	return data
}

// writeSerial write to MEMS
func (mems *ECUReaderInstance) writeSerial(data []byte) {
	if mems.Status.Emulated {
		log.WithFields(log.Fields{"data": fmt.Sprintf("%x", data)}).Info("data stored for emulation")
		mems.CommandResponse.Command = data
	} else {
		if mems.Status.Connected {
			if mems.SerialPort != nil {
				// save the sent command
				mems.CommandResponse.Command = data

				// write the response to the code reader
				n, e := mems.SerialPort.Write(data)

				if e != nil {
					log.WithFields(log.Fields{"error": e}).Error("error sending data to serial serialPort")
				}

				if n > 0 {
					log.WithFields(log.Fields{"data": fmt.Sprintf("%x", data)}).Info("data to serial serialPort")
				}
			}
		}
	}
}

// getResponseSize returns the expected number of bytes for a given command
// The 'response' variable contains the formats for each command response pattern
// by default the response size is 2 bytes unless the command has a special format.
func (mems *ECUReaderInstance) getResponseSize(command []byte) int {
	size := 2

	c := hex.EncodeToString(command)
	c = strings.ToUpper(c)
	r := responseMap[c]

	if r != nil {
		size = len(r)
	} else {
		log.Warn("unable to find a matching response in the map, assuming 2 byte response")
		r = responseMap["00"]
		copy(r[0:], command)
	}

	log.Infof("mapped command %s, expecting a response of %d bytes", fmt.Sprintf("%x", r), size)

	return size
}

// package init function
func init() {
	// Response formats for commands that do not respond with the format [COMMAND][VALUE]
	// Generally these are either part of the initialisation sequence or are ECU data frames
	responseMap["0A"] = []byte{0x0A}
	responseMap["CA"] = []byte{0xCA}
	responseMap["75"] = []byte{0x75}

	// Format for DataFrames starts with [Command Echo][Data Size][Data Bytes (28 for 0x80 and 32 for 0x7D)]
	responseMap["80"] = []byte{0x80, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B}
	responseMap["7D"] = []byte{0x7d, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F}
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
}
*/
