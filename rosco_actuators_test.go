package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_actuators_activateActuator(t *testing.T) {
	var err error
	var connected bool

	virtualPort := getVirtualPort()

	r := NewECUReaderInstance()
	connected, err = r.ConnectAndInitialiseECU(virtualPort)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	// fuel pump
	err = r.activateActuator(MEMSFuelPumpOn, MEMSFuelPumpOff, true)
	then.AssertThat(t, err, is.Nil())

	err = r.activateActuator(MEMSFuelPumpOn, MEMSFuelPumpOff, false)
	then.AssertThat(t, err, is.Nil())
}

func Test_adjustments_AllActuators(t *testing.T) {
	var err error
	var connected bool

	virtualPort := getVirtualPort()

	r := NewECUReaderInstance()
	connected, err = r.ConnectAndInitialiseECU(virtualPort)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	err = r.TestFuelPump(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestPTCRelay(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestACRelay(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestPurgeValve(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestO2Heater(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestBoostValve(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestFan1(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestFan2(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestInjectors(true)
	then.AssertThat(t, err, is.Nil())

	err = r.TestCoil(true)
	then.AssertThat(t, err, is.Nil())

	_ = r.Disconnect()
}
