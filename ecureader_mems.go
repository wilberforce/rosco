package rosco

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"time"
)

type MEMSReader struct {
	connected  bool
	port       string
	ecuId      string
	ecuSerial  string
	serialPort *serial.Port
}

func NewMEMSReader(connection string) *MEMSReader {
	log.Infof("created mems ecu reader")

	r := &MEMSReader{}
	r.port = connection
	return r
}

func (r *MEMSReader) Connect() (bool, error) {
	r.connected = false

	if err := r.connectToSerialPort(r.port); err != nil {
		log.Errorf("error opening serial serialPort (%s) status : (%+v)", r.port, err)
		// connect failure if we cannot open the serialPort
		return false, err
	}

	if err := r.initialiseMemsECU(); err != nil {
		log.Errorf("error opening serial serialPort (%s) status : (%+v)", r.port, err)
		// connect failure if we cannot initialise successfully
		// disconnect from the ecu
		_ = r.Disconnect()
		return false, err
	}

	// connected, no errors
	r.connected = true

	return r.connected, nil
}

func (r *MEMSReader) SendAndReceive(command []byte) ([]byte, error) {
	var response []byte
	var err error

	if r.serialPort != nil {
		if r.connected {
			r.writeSerial(command)
			response, err = r.readSerial(command)
		} else {
			err = fmt.Errorf("ecu is not connected, unable to send %X", command)
			log.Errorf("%s", err)
		}
	} else {
		err = fmt.Errorf("serial connnection not initialised, unable to send %X", command)
		log.Errorf("%s", err)
	}

	return response, err
}

func (r *MEMSReader) Disconnect() error {
	var err error

	// don't try and close an uninitialised serial serialPort
	// the serial library throws and ugly fatal if that happens
	if r.serialPort != nil {
		if r.connected {
			if err = r.serialPort.Flush(); err != nil {
				log.Warnf("error flushing serial serialPort (%+v)", err)
			}

			if err = r.serialPort.Close(); err != nil {
				log.Warnf("error closing serial serialPort (%+v)", err)
			} else {
				log.Infof("serial port closed successully")
			}
		}
	}

	r.connected = false
	return err
}

func (r *MEMSReader) connectToSerialPort(port string) error {
	var err error

	log.Infof("attempting to open serial serialPort %s", port)

	// connect to the ecu, timeout if we don't get data after a couple of seconds
	c := &serial.Config{Name: port, Baud: 9600, ReadTimeout: time.Millisecond * 2000}

	if r.serialPort, err = serial.OpenPort(c); err != nil {
		log.Errorf("error opening serial port (%s)", err)
	}

	return err
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
func (r *MEMSReader) initialiseMemsECU() error {
	_ = r.serialPort.Flush()

	log.Infof("initialising ecu")

	r.writeSerial(MEMSInitCommandA)
	if response, err := r.readSerial(MEMSInitCommandA); err != nil {
		// abandon initialisation if error occurred
		log.Errorf("mems initialisation failed command %X (%s)", MEMSInitCommandA, err)
		return err
	} else {
		// if we get the command echoed back we can assume good connection and proceed.
		// This is to work around the issue  in Windows where the serialPort always connects even if it's not available.
		if len(response) == 0 {
			err = fmt.Errorf("0 bytes received, serial serialPort read error, timeout? (%s)", err)
			log.Errorf("%s", err)
			return err
		}

		if response[0] == MEMSInitCommandA[0] {
			r.writeSerial(MEMSInitCommandB)

			if response, err = r.readSerial(MEMSInitCommandB); err != nil {
				// abandon initialisation if error occurred
				log.Errorf("mems initialisation failed command %X (%s)", MEMSInitCommandB, err)
				return err
			}

			r.writeSerial(MEMSHeartbeat)
			if response, err = r.readSerial(MEMSHeartbeat); err != nil {
				// abandon initialisation if error occurred
				log.Errorf("mems initialisation failed command %X (%s)", MEMSHeartbeat, err)
				return err
			}

			r.writeSerial(MEMSInitECUID)
			if response, err = r.readSerial(MEMSInitECUID); err != nil {
				// abandon initialisation if error occurred
				log.Errorf("mems initialisation failed command %X (%s)", MEMSInitECUID, err)
				return err
			} else {
				log.Infof("mems initialisation ECU ID %X successful", response)
			}
		}
	}

	log.Infof("mems initialised")
	return nil
}

// readSerial read from MEMS
// read 1 byte at a time until we have all the expected bytes
func (r *MEMSReader) readSerial(command []byte) ([]byte, error) {
	var bytesRead int
	var err error

	size, err := getResponseSize(command)

	// serial read buffer
	b := make([]byte, size)

	//  receivedBytes frame buffer
	receivedBytes := make([]byte, 0)

	if r.serialPort != nil {
		// read all the expected bytes before returning the receivedBytes
		for count := 0; count < size; {
			// wait for a response from MEMS
			bytesRead, err = r.serialPort.Read(b)

			if bytesRead == 0 {
				err = fmt.Errorf("0 bytes received, serial serialPort read error, timeout? (%s)", err)
				log.Errorf("%s", err)
				// drop out of loop, send back a 0x00 byte array response
				// this prevents the loop getting blocked on a read error
				count = size
				receivedBytes = append(receivedBytes, b...)
			} else {
				// append the read bytes to the receivedBytes frame
				receivedBytes = append(receivedBytes, b[:bytesRead]...)
			}

			// increment by the number of bytes read
			count = count + bytesRead
			if count > size {
				err = fmt.Errorf("received dataframe size mismatch, received %d, expected %d", count, size)
				log.Errorf("%s", err)
			}
		}
	}

	log.Infof("received %X from ecu, %d bytes", receivedBytes, bytesRead)

	if err == nil {
		err = r.commandMatchesResponse(command, receivedBytes)
	}

	return receivedBytes, err
}

func (r *MEMSReader) commandMatchesResponse(command []byte, receivedBytes []byte) error {
	var err error

	if command[0] != receivedBytes[0] {
		err = fmt.Errorf("expecting command echo of %X, received %X", command[0], receivedBytes[0])
		log.Errorf("%s", err)
	}

	return err
}

// writeSerial write to MEMS
func (r *MEMSReader) writeSerial(data []byte) {
	if r.serialPort != nil {
		bytesWritten, err := r.serialPort.Write(data)

		if err != nil {
			log.Errorf("error sending %X to the ecu (%s)", data, err)
		}

		if bytesWritten > 0 {
			log.Infof("sent %X to ecu", data)
		}
	}
}
