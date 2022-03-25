package rosco

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

// AdjustShortTermFuelTrim increments or decrements by the number of steps
func (ecu *ECUReaderInstance) AdjustShortTermFuelTrim(steps int) (int, error) {
	return ecu.applyAdjustment(MEMSSTFTIncrement, MEMSSTFTDecrement, MEMSFuelTrimDefault, steps)
}

// AdjustLongTermFuelTrim increments or decrements by the number of steps
func (ecu *ECUReaderInstance) AdjustLongTermFuelTrim(steps int) (int, error) {
	return ecu.applyAdjustment(MEMSLTFTIncrement, MEMSLTFTDecrement, MEMSFuelTrimDefault, steps)
}

// AdjustIdleDecay increments or decrements by the number  of steps
func (ecu *ECUReaderInstance) AdjustIdleDecay(steps int) (int, error) {
	return ecu.applyAdjustment(MEMSIdleDecayIncrement, MEMSIdleDecayDecrement, MEMSIdleDecayDefault, steps)
}

// AdjustIdleSpeed increments or decrements by the number of steps
func (ecu *ECUReaderInstance) AdjustIdleSpeed(steps int) (int, error) {
	return ecu.applyAdjustment(MEMSIdleSpeedIncrement, MEMSIdleSpeedDecrement, MEMSIdleSpeedDefault, steps)
}

// AdjustIgnitionAdvanceOffset increments or decrements by the number of steps
func (ecu *ECUReaderInstance) AdjustIgnitionAdvanceOffset(steps int) (int, error) {
	return ecu.applyAdjustment(MEMSIgnitionAdvanceOffsetIncrement, MEMSIgnitionAdvanceOffsetDecrement, MEMSIgnitionAdvanceOffsetDefault, steps)
}

// AdjustIACPosition increments or decrements by the number of steps
func (ecu *ECUReaderInstance) AdjustIACPosition(steps int) (int, error) {
	return ecu.applyAdjustment(MEMSIACIncrement, MEMSIACDecrement, MEMSIACPositionDefault, steps)
}

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

//
// Private functions
//

// Increment or Decrement the adjustment by n steps
// Returns the final value of the adjustment
func (ecu *ECUReaderInstance) applyAdjustment(incrementCommand []byte, decrementCommand []byte, defaultValue int, steps int) (int, error) {
	var err error

	// no adjustment required
	if steps == 0 {
		err = fmt.Errorf("0 step adjustment requested, ignoring.")
		log.Warnf("%s", err)
		return defaultValue, err
	}

	if steps > 0 {
		// if the steps are positive then increment the adjustment by n steps.
		return ecu.incementAdjustment(incrementCommand, steps)
	} else {
		// if the steps are negative then decrement the adjustment by n steps.
		return ecu.decrementAdjustment(decrementCommand, steps)
	}
}

func (ecu *ECUReaderInstance) decrementAdjustment(cmd []byte, steps int) (int, error) {
	var err error
	var data []byte

	log.Infof("decrementing adjustable command %X by %d steps", data, steps)
	for step := steps; step < 0; step++ {
		if data, err = ecu.ecuReader.SendAndReceive(cmd); err == nil {
			log.Infof("command %X deccremented to %X", cmd, data)
		}
	}

	if err == nil {
		return int(data[1]), err
	} else {
		return 0, err
	}
}

func (ecu *ECUReaderInstance) incementAdjustment(cmd []byte, steps int) (int, error) {
	var err error
	var data []byte

	log.Infof("incrementing adjustable command %X by %d steps", data, steps)

	for step := 0; step < steps; step++ {
		if data, err = ecu.ecuReader.SendAndReceive(cmd); err == nil {
			log.Infof("command %X incremented to %X", cmd, data)
		}
	}

	if err == nil {
		return int(data[1]), err
	} else {
		return 0, err
	}
}
