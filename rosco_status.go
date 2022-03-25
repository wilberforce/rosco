package rosco

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

// GetStatus returns the connection and ECU status
func (ecu *ECUReaderInstance) GetStatus() ECUStatus {
	log.Infof("getting ecu status (%+v)", ecu.Status)
	return *ecu.Status
}

func (ecu *ECUReaderInstance) SendHeartbeat() error {
	log.Info("sending ecu heartbeat")
	return ecu.updateECUState(MEMSHeartbeat)
}

// ResetAdjustments resets the adjustable values
func (ecu *ECUReaderInstance) ResetAdjustments() error {
	log.Info("resetting  ecu adjustable values ")
	return ecu.updateECUState(MEMSResetAdj)
}

// ResetECU clears fault codes. resets adjustable values and learnt values
func (ecu *ECUReaderInstance) ResetECU() error {
	log.Info("resetting ecu")
	return ecu.updateECUState(MEMSResetECU)
}

// ClearFaults clears fault codes
func (ecu *ECUReaderInstance) ClearFaults() error {
	log.Info("clearing ecu recorded faults ")
	return ecu.updateECUState(MEMSClearFaults)
}

// Updates ECU state, is used to clear the state for the reset commands or emitting a state keep-alive heartbeat
// Returns success of the operation
func (ecu *ECUReaderInstance) updateECUState(command []byte) error {
	var err error
	var data []byte

	if data, err = ecu.ecuReader.SendAndReceive(command); err == nil {
		log.Infof("updated ECU state with clear, reset or heartbeat (%X)", data)
	}

	return err
}

func (ecu *ECUReaderInstance) resetStatus() {
	ecu.Status.Connected = false
	ecu.Status.ECUID = ""
	ecu.Status.ECUSerial = ""
	ecu.Status.IACPosition = 0
}

func (ecu *ECUReaderInstance) getECUID() (string, error) {
	var data []byte
	var err error
	var ecuId string

	log.Info("reading ecu id")

	if data, err = ecu.ecuReader.SendAndReceive(MEMSInitECUID); err == nil {
		ecuId = fmt.Sprintf("%X", data[1:])
		log.Infof("ecu id %X received", ecuId)
	} else {
		log.Warnf("error recieving ecu id %X (%s)", data, err)
	}

	return ecuId, err
}

func (ecu *ECUReaderInstance) getECUSerial() (string, error) {
	var data []byte
	var err error
	var ecuSerial string

	log.Info("reading ecu serial")

	if data, err = ecu.ecuReader.SendAndReceive(MEMSGetECUSerial); err == nil {
		ecuSerial = fmt.Sprintf("%s%X", data[1:9], data[9:])
		log.Infof("ecu serial %s received", ecuSerial)
	} else {
		log.Warnf("error recieving ecu serial %X (%s)", data, err)
	}

	return ecuSerial, err
}

func (ecu *ECUReaderInstance) getIACPosition() (int, error) {
	var data []byte
	var err error

	log.Info("reading ecu iac position ")

	if data, err = ecu.ecuReader.SendAndReceive(MEMSGetIACPosition); err == nil {
		log.Infof("ecu iac position, received (%X)", data)
		return int(data[1]), err
	} else {
		log.Warnf("ecu iac position invalid, received (%X)", data)
		return MEMSIACPositionDefault, err
	}
}
