package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
)

func Test_isValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:        5999,
		CoolantTemp:      119,
		IntakeAirTemp:    79,
		IdleBasePosition: 249,
		DTC5:             255,
	}

	valid := d.isValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		EngineRPM:        6000,
		CoolantTemp:      120,
		IntakeAirTemp:    80,
		IdleBasePosition: 250,
		DTC5:             254,
	}

	valid = d.isValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_EngineRPMIsValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM: 5999,
	}

	valid := d.isEngineRPMValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		EngineRPM: 6000,
	}

	valid = d.isValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_CoolantIsValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTemp: 119,
	}

	valid := d.isCoolantTempValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		CoolantTemp: 120,
	}

	valid = d.isValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_isIntakeAirTempValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		IntakeAirTemp: 79,
	}

	valid := d.isIntakeAirTempValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		IntakeAirTemp: 80,
	}

	valid = d.isIntakeAirTempValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_isIdleBasePositionValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		IdleBasePosition: 249,
	}

	valid := d.isIdleBasePositionValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		IdleBasePosition: 250,
	}

	valid = d.isIdleBasePositionValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_isDTC5Valid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		DTC5: 255,
	}

	valid := d.isDTC5Valid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		DTC5: 254,
	}

	valid = d.isDTC5Valid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_addToDataset(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM: 1,
	}

	d.addToDataset(data)

	then.AssertThat(t, len(d.dataset), is.EqualTo(1))
}

func Test_fillAddToDataset(t *testing.T) {
	maxDataframeItems := 30
	d := NewDataframeAnalysis(maxDataframeItems)

	data := MemsData{
		EngineRPM: 1,
	}

	// add the first valid item
	d.addToDataset(data)
	then.AssertThat(t, len(d.dataset), is.EqualTo(1))

	data = MemsData{
		EngineRPM: 5999,
	}

	for i := 1; i < maxDataframeItems; i++ {
		d.addToDataset(data)
		then.AssertThat(t, len(d.dataset), is.EqualTo(i+1))
	}

	// add another item, buffer should remain at 30
	d.addToDataset(data)
	then.AssertThat(t, len(d.dataset), is.EqualTo(maxDataframeItems))
	//  first item should be dropped which had an engine rpm of 1
	then.AssertThat(t, d.dataset[0].EngineRPM, is.EqualTo(5999))
}

func Test_getExpectedEngineWarmTime(t *testing.T) {
	d := NewDataframeAnalysis(1)
	expectedWarmTime, _ := time.Parse("15:04:05.000", "12:01:50.000")

	data := MemsData{
		Time:        "12:00:00.000",
		CoolantTemp: 70,
	}

	warmAt := d.getExpectedEngineWarmTime(data)
	then.AssertThat(t, warmAt, is.EqualTo(expectedWarmTime))
}

func Test_isEngineRunning(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM: 0,
	}

	running := d.isEngineRunning(data)
	then.AssertThat(t, running, is.False())

	data = MemsData{
		EngineRPM: 1,
	}

	running = d.isEngineRunning(data)
	then.AssertThat(t, running, is.True())
}

func Test_isEngineWarming(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTemp: 77,
	}

	result := d.isEngineWarming(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		CoolantTemp: 78,
	}

	result = d.isEngineWarming(data)
	then.AssertThat(t, result, is.False())
}

func Test_isEngineWarm(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTemp: 77,
	}

	result := d.isEngineWarm(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		CoolantTemp: 78,
	}

	result = d.isEngineWarm(data)
	then.AssertThat(t, result, is.True())
}

func Test_isBatteryVoltageLow(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		BatteryVoltage: 12.7,
	}

	result := d.isBatteryVoltageLow(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		BatteryVoltage: 13,
	}

	result = d.isBatteryVoltageLow(data)
	then.AssertThat(t, result, is.False())
}

func Test_isLoopClosed(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		ClosedLoop: false,
	}

	result := d.isLoopClosed(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		ClosedLoop: true,
	}

	result = d.isLoopClosed(data)
	then.AssertThat(t, result, is.True())
}

func Test_isThrottleActive(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		ThrottleAngle: 14,
	}

	result := d.isThrottleActive(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		ThrottleAngle: 15,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM: 1250,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		EngineRPM: 1301,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		ThrottleAngle: 15,
		EngineRPM:     1250,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		ThrottleAngle: 14,
		EngineRPM:     1301,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())
}

func Test_isEngineIdle(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		ThrottleAngle: 14,
		EngineRPM:     1,
	}

	result := d.isEngineIdle(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM:     1,
		ThrottleAngle: 15,
	}

	result = d.isEngineIdle(data)
	then.AssertThat(t, result, is.False())
}

func Test_Analyse(t *testing.T) {
	d := NewDataframeAnalysis(30)

	// valid data
	data := MemsData{
		EngineRPM:                5999,
		CoolantTemp:              119,
		IntakeAirTemp:            79,
		IdleBasePosition:         249,
		DTC5:                     255,
		BatteryVoltage:           13,
		CoolantTempSensorFault:   false,
		IntakeAirTempSensorFault: false,
		FuelPumpCircuitFault:     false,
		ThrottlePotCircuitFault:  false,
	}

	d.Analyse(data)
	then.AssertThat(t, d.dataset[0].EngineRPM, is.EqualTo(5999))
	then.AssertThat(t, d.Analysis.CoolantTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.IntakeAirTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.FuelPumpCircuitFault, is.False())
	then.AssertThat(t, d.Analysis.ThrottlePotCircuitFault, is.False())
	then.AssertThat(t, d.Analysis.BatteryFault, is.False())

	// invalid data, shouldn't get added to the dataset
	data = MemsData{
		EngineRPM:                6000,
		CoolantTemp:              120,
		IntakeAirTemp:            80,
		IdleBasePosition:         250,
		DTC5:                     254,
		BatteryVoltage:           13,
		CoolantTempSensorFault:   true,
		IntakeAirTempSensorFault: true,
		FuelPumpCircuitFault:     true,
		ThrottlePotCircuitFault:  true,
	}

	d.Analyse(data)
	then.AssertThat(t, len(d.dataset), is.EqualTo(1))
	then.AssertThat(t, d.Analysis.CoolantTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.IntakeAirTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.FuelPumpCircuitFault, is.False())
	then.AssertThat(t, d.Analysis.ThrottlePotCircuitFault, is.False())

	// engine not running, shouldn't get added to the dataset
	// but the ecu faults should be detected but
	// operational faults are not
	data = MemsData{
		EngineRPM:                0,
		CoolantTemp:              50,
		IntakeAirTemp:            20,
		IdleBasePosition:         100,
		DTC5:                     255,
		BatteryVoltage:           12.7,
		CoolantTempSensorFault:   true,
		IntakeAirTempSensorFault: true,
		FuelPumpCircuitFault:     true,
		ThrottlePotCircuitFault:  true,
	}

	d.Analyse(data)
	then.AssertThat(t, len(d.dataset), is.EqualTo(1))
	then.AssertThat(t, d.Analysis.CoolantTempSensorFault, is.True())
	then.AssertThat(t, d.Analysis.IntakeAirTempSensorFault, is.True())
	then.AssertThat(t, d.Analysis.FuelPumpCircuitFault, is.True())
	then.AssertThat(t, d.Analysis.ThrottlePotCircuitFault, is.True())
	then.AssertThat(t, d.Analysis.BatteryFault, is.False())
}

func Test_analyseOperationalStatus(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:     0,
		CoolantTemp:   77,
		ClosedLoop:    false,
		ThrottleAngle: 14,
	}

	d.analyseOperationalStatus(data)
	then.AssertThat(t, d.Analysis.IsEngineRunning, is.False())
	then.AssertThat(t, d.Analysis.IsEngineWarming, is.True())
	then.AssertThat(t, d.Analysis.IsAtOperatingTemp, is.False())
	then.AssertThat(t, d.Analysis.IsClosedLoop, is.False())
	then.AssertThat(t, d.Analysis.IsThrottleActive, is.False())
	then.AssertThat(t, d.Analysis.IsEngineIdle, is.False())

	data = MemsData{
		EngineRPM:     1,
		CoolantTemp:   78,
		ClosedLoop:    true,
		ThrottleAngle: 15,
	}

	d.analyseOperationalStatus(data)
	then.AssertThat(t, d.Analysis.IsEngineRunning, is.True())
	then.AssertThat(t, d.Analysis.IsEngineWarming, is.False())
	then.AssertThat(t, d.Analysis.IsAtOperatingTemp, is.True())
	then.AssertThat(t, d.Analysis.IsClosedLoop, is.True())
	then.AssertThat(t, d.Analysis.IsThrottleActive, is.True())
	then.AssertThat(t, d.Analysis.IsEngineIdle, is.False())

	data = MemsData{
		EngineRPM:     1,
		CoolantTemp:   78,
		ClosedLoop:    true,
		ThrottleAngle: 14,
	}

	d.analyseOperationalStatus(data)
	then.AssertThat(t, d.Analysis.IsEngineRunning, is.True())
	then.AssertThat(t, d.Analysis.IsEngineWarming, is.False())
	then.AssertThat(t, d.Analysis.IsAtOperatingTemp, is.True())
	then.AssertThat(t, d.Analysis.IsClosedLoop, is.True())
	then.AssertThat(t, d.Analysis.IsThrottleActive, is.False())
	then.AssertThat(t, d.Analysis.IsEngineIdle, is.True())
}

func Test_analyseOperationalFaults(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		BatteryVoltage: 12.7,
	}

	d.analyseOperationalFaults(data)
	then.AssertThat(t, d.Analysis.BatteryFault, is.True())

	data = MemsData{
		BatteryVoltage: 13,
	}

	d.analyseOperationalFaults(data)
	then.AssertThat(t, d.Analysis.BatteryFault, is.False())
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

func Test_analyseFaults(t *testing.T) {
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

func Test_isCoilFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		BatteryVoltage: 13,
		CoilTime:       4.1,
	}

	result := d.isCoilFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		BatteryVoltage: 13,
		CoilTime:       4,
	}

	result = d.isCoilFaulty(data)
	then.AssertThat(t, result, is.False())

	// battery low, high coil time ignored
	data = MemsData{
		BatteryVoltage: 12.7,
		CoilTime:       4.1,
	}

	result = d.isCoilFaulty(data)
	then.AssertThat(t, result, is.False())
}
