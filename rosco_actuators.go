package rosco

import log "github.com/sirupsen/logrus"

// Switches on or off the actuator
// Returns the success of the operation
func (ecu *ECUReaderInstance) activateActuator(activateCommand []byte, deactivateCommand []byte, activate bool) error {
	var err error
	var cmd []byte
	var data []byte

	if activate {
		if data, err = ecu.ecuReader.SendAndReceive(activateCommand); err == nil {
			log.Infof("actuator %X activated (%X)", cmd, data)
		}
	} else {
		if data, err = ecu.ecuReader.SendAndReceive(deactivateCommand); err == nil {
			log.Infof("actuator %X deactivated (%X)", cmd, data)
		}
	}

	return err
}
