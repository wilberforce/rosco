package rosco

import log "github.com/sirupsen/logrus"

// TestFuelPump test
func (ecu *ECUReaderInstance) TestFuelPump(activate bool) error {
	return ecu.activateActuator(MEMSFuelPumpOn, MEMSFuelPumpOff, activate)
}

// PTCRelay test
func (ecu *ECUReaderInstance) TestPTCRelay(activate bool) error {
	return ecu.activateActuator(MEMSPTCRelayOn, MEMSPTCRelayOff, activate)
}

// ACRelay test
func (ecu *ECUReaderInstance) TestACRelay(activate bool) error {
	return ecu.activateActuator(MEMSACRelayOn, MEMSACRelayOff, activate)
}

// TestPurgeValve test
func (ecu *ECUReaderInstance) TestPurgeValve(activate bool) error {
	return ecu.activateActuator(MEMSPurgeValveOn, MEMSPurgeValveOff, activate)
}

// TestO2Heater test
func (ecu *ECUReaderInstance) TestO2Heater(activate bool) error {
	return ecu.activateActuator(MEMSO2HeaterOn, MEMSO2HeaterOff, activate)
}

// TestBoostValve test
func (ecu *ECUReaderInstance) TestBoostValve(activate bool) error {
	return ecu.activateActuator(MEMSBoostValveOn, MEMSBoostValveOff, activate)
}

// TestFan1 test
func (ecu *ECUReaderInstance) TestFan1(activate bool) error {
	return ecu.activateActuator(MEMSFan1On, MEMSFan1Off, activate)
}

// TestFan2 test
func (ecu *ECUReaderInstance) TestFan2(activate bool) error {
	return ecu.activateActuator(MEMSFan2On, MEMSFan2Off, activate)
}

// TestInjectors test, the activate state is ignored on this test
func (ecu *ECUReaderInstance) TestInjectors(activate bool) error {
	return ecu.activateActuator(MEMSTestInjectors, MEMSTestInjectors, activate)
}

// TestCoil test, the activate state is ignored on this test
func (ecu *ECUReaderInstance) TestCoil(activate bool) error {
	return ecu.activateActuator(MEMSFireCoil, MEMSFireCoil, activate)
}

// Switches on or off the actuator
// Returns the success of the operation
func (ecu *ECUReaderInstance) activateActuator(activateCommand []byte, deactivateCommand []byte, activate bool) error {
	var err error
	var data []byte

	if activate {
		if data, err = ecu.ecuReader.SendAndReceive(activateCommand); err == nil {
			log.Infof("actuator %X activated (%X)", activateCommand, data)
		}
	} else {
		if data, err = ecu.ecuReader.SendAndReceive(deactivateCommand); err == nil {
			log.Infof("actuator %X deactivated (%X)", deactivateCommand, data)
		}
	}

	return err
}
