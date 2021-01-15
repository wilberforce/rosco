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

	port := "/dev/null"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Connected, is.False())
	then.AssertThat(t, mems.Initialised, is.False())
}

func TestConnectAndInitialiseScenario(t *testing.T) {
	port := "scenario.csv"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Connected, is.True())
	then.AssertThat(t, mems.Initialised, is.True())
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
