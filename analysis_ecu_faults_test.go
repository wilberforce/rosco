package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_analyseECUFaults(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTempSensorFault:   false,
		IntakeAirTempSensorFault: false,
		FuelPumpCircuitFault:     false,
		ThrottlePotCircuitFault:  false,
	}

	d.analyseECUFaults(data)
	then.AssertThat(t, d.Analysis.CoolantTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.IntakeAirTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.FuelPumpCircuitFault, is.False())
	then.AssertThat(t, d.Analysis.ThrottlePotCircuitFault, is.False())

	data = MemsData{
		CoolantTempSensorFault:   true,
		IntakeAirTempSensorFault: true,
		FuelPumpCircuitFault:     true,
		ThrottlePotCircuitFault:  true,
	}

	d.analyseECUFaults(data)
	then.AssertThat(t, d.Analysis.CoolantTempSensorFault, is.True())
	then.AssertThat(t, d.Analysis.IntakeAirTempSensorFault, is.True())
	then.AssertThat(t, d.Analysis.FuelPumpCircuitFault, is.True())
	then.AssertThat(t, d.Analysis.ThrottlePotCircuitFault, is.True())
}

func Test_isIntakeAirTempSensorFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		IntakeAirTempSensorFault: true,
	}

	result := d.isIntakeAirTempSensorFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		IntakeAirTempSensorFault: false,
	}

	result = d.isIntakeAirTempSensorFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isThrottlePotCircuitFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		ThrottlePotCircuitFault: true,
	}

	result := d.isThrottlePotCircuitFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		ThrottlePotCircuitFault: false,
	}

	result = d.isThrottlePotCircuitFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isFuelPumpCircuitFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		FuelPumpCircuitFault: true,
	}

	result := d.isFuelPumpCircuitFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		FuelPumpCircuitFault: false,
	}

	result = d.isFuelPumpCircuitFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isCoolantSensorFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTempSensorFault: true,
	}

	result := d.isCoolantSensorFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		CoolantTempSensorFault: false,
	}

	result = d.isCoolantSensorFaulty(data)
	then.AssertThat(t, result, is.False())
}
