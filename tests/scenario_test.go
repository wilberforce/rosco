// +build scenario

package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestGetScenarios(t *testing.T) {
	scenarios, err := rosco.GetScenarios()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(scenarios), is.GreaterThan(0))
}

func TestConnectInitialiseScenario(t *testing.T) {
	port := "nofaults-warm.csv"

	mems := rosco.NewMemsConnection(".")
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Connected, is.True())
	then.AssertThat(t, mems.Status.Initialised, is.True())
}

func TestStatsScenario(t *testing.T) {
	port := "nofaults-warm.csv"
	mems := rosco.NewMemsConnection(".")
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Initialised, is.True())

	if mems.Status.Initialised {
		// get 30 data points
		for i := 0; i < 30; i++ {
			_ = mems.GetDataframes()

			then.AssertThat(t, mems.CommandResponse.Command, is.EqualTo(rosco.MEMSDataFrame))
			then.AssertThat(t, mems.CommandResponse.Response, is.EqualTo(rosco.MEMSDataFrame))
		}

		mems.Diagnostics.Analyse()
		stats := mems.Diagnostics

		then.AssertThat(t, stats.Analysis.IsEngineRunning, is.True())
		then.AssertThat(t, stats.Analysis.IsEngineWarming, is.False())
		then.AssertThat(t, stats.Analysis.IsAtOperatingTemp, is.True())

		then.AssertThat(t, stats.Stats["LambdaVoltage"].Mean, is.GreaterThanOrEqualTo(0.0))
	}
}
