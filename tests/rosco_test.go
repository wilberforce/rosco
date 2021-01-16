package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"io/ioutil"
	"log"
	"testing"
)

func TestConnectAndInitialise(t *testing.T) {
	// disable internal logging when running tests
	log.SetOutput(ioutil.Discard)

	// ensure memsulator is running and change the port
	// to the ttyecu path
	port := "/Users/ajackson/ttyecu"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Connected, is.True())
	then.AssertThat(t, mems.Status.Initialised, is.True())

	mems.Disconnect()
	then.AssertThat(t, mems.Status.Connected, is.False())
	then.AssertThat(t, mems.Status.Initialised, is.False())
}

func TestConnectAndInitialiseScenario(t *testing.T) {
	port := "scenario.csv"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Connected, is.True())
	then.AssertThat(t, mems.Status.Initialised, is.True())
}

func TestScenarioGetDataframe(t *testing.T) {
	port := "scenario.csv"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	data := mems.GetDataframes()

	then.AssertThat(t, data.BatteryVoltage, is.GreaterThan(0.0))
}

func TestScenarioSendCommand(t *testing.T) {
	port := "scenario.csv"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	data, _ := mems.SendCommand(rosco.MEMSGetIACPosition)

	then.AssertThat(t, data[0], is.EqualTo(rosco.MEMSGetIACPosition[0]))
}
