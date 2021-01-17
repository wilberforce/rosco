package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func getPort(useScenario bool) string {
	if useScenario {
		return "scenario.csv"
	}

	// ensure memsulator is running and change the port
	// to the ttyecu path
	return "/Users/ajackson/ttyecu"
}

func TestConnectAndInitialise(t *testing.T) {
	// disable internal logging when running tests
	//log.SetOutput(ioutil.Discard)

	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Connected, is.True())
	then.AssertThat(t, mems.Status.Initialised, is.True())

	mems.Disconnect()
	then.AssertThat(t, mems.Status.Connected, is.False())
	then.AssertThat(t, mems.Status.Initialised, is.False())
}

func TestConnectAndInitialiseScenario(t *testing.T) {
	port := getPort(true)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Connected, is.True())
	then.AssertThat(t, mems.Status.Initialised, is.True())
}

func TestScenarioGetDataframe(t *testing.T) {
	port := getPort(true)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	data := mems.GetDataframes()

	then.AssertThat(t, data.BatteryVoltage, is.GreaterThan(0.0))
}

func TestAdjustSTFT(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	trim := mems.AdjustShortTermFuelTrim(1)
	then.AssertThat(t, trim, is.EqualTo(rosco.MEMSFuelTrimDefault+1))

	trim = mems.AdjustShortTermFuelTrim(-1)
	then.AssertThat(t, trim, is.EqualTo(rosco.MEMSFuelTrimDefault-1))

	trim = mems.AdjustShortTermFuelTrim(0)
	then.AssertThat(t, trim, is.EqualTo(rosco.MEMSFuelTrimDefault))
}

func TestResetAdjustments(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	success := mems.ResetAdjustments()
	then.AssertThat(t, success, is.True())
}

func TestResetECU(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	success := mems.ResetECU()
	then.AssertThat(t, success, is.True())
}

func TestClearFaults(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	success := mems.ClearFaults()
	then.AssertThat(t, success, is.True())
}

func TestGetIACPosition(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	success := mems.GetIACPosition()
	then.AssertThat(t, success, is.EqualTo(rosco.MEMSIACPositionDefault))
}

func TestIdleDecay(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	value := mems.AdjustIdleDecay(1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIdleDecayDefault+1))

	value = mems.AdjustIdleDecay(-1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIdleDecayDefault-1))

	value = mems.AdjustIdleDecay(0)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIdleDecayDefault))
}

func TestIdleSpeed(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	value := mems.AdjustIdleSpeed(1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIdleSpeedDefault+1))

	value = mems.AdjustIdleSpeed(-1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIdleSpeedDefault-1))

	value = mems.AdjustIdleSpeed(0)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIdleSpeedDefault))
}

func TestIgnitionAdvanceOffset(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	value := mems.AdjustIgnitionAdvanceOffset(1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIgnitionAdvanceOffsetDefault+1))

	value = mems.AdjustIgnitionAdvanceOffset(-1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIgnitionAdvanceOffsetDefault-1))

	value = mems.AdjustIgnitionAdvanceOffset(0)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIgnitionAdvanceOffsetDefault))
}

func TestIACPosition(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	value := mems.AdjustIACPosition(1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIACPositionDefault+1))

	value = mems.AdjustIACPosition(-1)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIACPositionDefault-1))

	value = mems.AdjustIACPosition(0)
	then.AssertThat(t, value, is.EqualTo(rosco.MEMSIACPositionDefault))
}

func TestHeartbeat(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	value := mems.SendHeartbeat()
	then.AssertThat(t, value, is.True())
}

func TestActuators(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	value := mems.TestFuelPump(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestFuelPump(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestPTCRelay(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestPTCRelay(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestACRelay(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestACRelay(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestPurgeValve(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestPurgeValve(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestO2Heater(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestO2Heater(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestBoostValve(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestBoostValve(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestInjectors(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestInjectors(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestCoil(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestCoil(false)
	then.AssertThat(t, value, is.True())

	value = mems.TestFan1(true)
	then.AssertThat(t, value, is.True())

	// Testing Fan1 deactivate command causes issues on the emulated
	// serial port, assume this is a special character

	//value = mems.TestFan1(false)
	//then.AssertThat(t, value, is.True())

	value = mems.TestFan2(true)
	then.AssertThat(t, value, is.True())

	value = mems.TestFan2(false)
	then.AssertThat(t, value, is.True())
}
