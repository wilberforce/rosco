package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"io/ioutil"
	"log"
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
	log.SetOutput(ioutil.Discard)

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

func TestScenarioSendCommand(t *testing.T) {
	port := getPort(true)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	data, _ := mems.SendCommand(rosco.MEMSGetIACPosition)

	then.AssertThat(t, data[0], is.EqualTo(rosco.MEMSGetIACPosition[0]))
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
	then.AssertThat(t, success, is.EqualTo(true))
}

func TestResetECU(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	success := mems.ResetECU()
	then.AssertThat(t, success, is.EqualTo(true))
}

func TestClearFaults(t *testing.T) {
	port := getPort(false)

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	success := mems.ClearFaults()
	then.AssertThat(t, success, is.EqualTo(true))
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
