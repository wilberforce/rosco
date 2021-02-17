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
		then.AssertThat(t, stats.Analysis.IsEngineIdle, is.False())
		then.AssertThat(t, stats.Analysis.IsEngineIdleFault, is.False())
		then.AssertThat(t, stats.Analysis.IsCruising, is.False())
		then.AssertThat(t, stats.Analysis.IsAtOperatingTemp, is.True())
		then.AssertThat(t, stats.Analysis.ClosedLoopFault, is.False())
		then.AssertThat(t, stats.Analysis.MapFault, is.False())
		then.AssertThat(t, stats.Analysis.VacuumFault, is.False())
		then.AssertThat(t, stats.Analysis.IdleAirControlFault, is.False())
		then.AssertThat(t, stats.Analysis.IACMinFault, is.False())
		then.AssertThat(t, stats.Analysis.IACMaxFault, is.False())
		then.AssertThat(t, stats.Analysis.LambdaSensorFault, is.False())
		then.AssertThat(t, stats.Analysis.LambdaOscillationFault, is.False())
		then.AssertThat(t, stats.Analysis.LambdaRangeFault, is.False())
		then.AssertThat(t, stats.Analysis.ThermostatFault, is.False())
		then.AssertThat(t, stats.Analysis.CoolantTempSensorFault, is.False())
		then.AssertThat(t, stats.Analysis.IntakeAirTempSensorFault, is.False())
		then.AssertThat(t, stats.Analysis.FuelPumpCircuitFault, is.False())
		then.AssertThat(t, stats.Analysis.ThrottlePotCircuitFault, is.False())

		then.AssertThat(t, stats.Stats["LambdaVoltage"].Mean, is.GreaterThanOrEqualTo(0.0))
	}
}
