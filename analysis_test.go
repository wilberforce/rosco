package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
)

const (
	engineRunning             = 1
	engineStopped             = 0
	rpmIdle                   = 1250
	rpmCruising               = 1350
	warmEngineTemperature     = 80
	coldEngineTemperature     = 50
	lowBattery                = 12.7
	goodBattery               = 13
	invalidRPM                = 6000
	invalidCoolantTemperature = 120
	invalidIntakeTemperature  = 80
	expectedDTC5              = 255
	invalidDTC5               = 254
	invalidIdleBasePosition   = 250
	highIdleBasePosition      = 100
	goodIdleBasePosition      = 40
	goodIntakeTemperature     = 30
	idleThrottleAngle         = 14
	activeThrottleAngle       = 15
	highIdleMAP               = 46
	goodIdleMap               = 35
	highCoilTime              = 4.1
	goodCoilTime              = 4
	activeLambdaStatus        = 1
	inactiveLambdaStatus      = 0
	lowIdleHot                = 5
	goodIdleHot               = 40
	highIdleHot               = 60
	goodLambdaValue           = 435
	goodCASPosition           = 15
)

func Test_Analyse(t *testing.T) {
	d := NewDataframeAnalysis(30)

	// valid data
	data := MemsData{
		EngineRPM:                engineRunning,
		CoolantTemp:              warmEngineTemperature,
		IntakeAirTemp:            goodIntakeTemperature,
		IdleBasePosition:         goodIdleBasePosition,
		DTC5:                     expectedDTC5,
		BatteryVoltage:           goodBattery,
		CoolantTempSensorFault:   false,
		IntakeAirTempSensorFault: false,
		FuelPumpCircuitFault:     false,
		ThrottlePotCircuitFault:  false,
	}

	d.Analyse(data)
	then.AssertThat(t, d.dataset[0].EngineRPM, is.EqualTo(engineRunning))
	then.AssertThat(t, d.Analysis.CoolantTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.IntakeAirTempSensorFault, is.False())
	then.AssertThat(t, d.Analysis.FuelPumpCircuitFault, is.False())
	then.AssertThat(t, d.Analysis.ThrottlePotCircuitFault, is.False())
	then.AssertThat(t, d.Analysis.BatteryFault, is.False())

	// invalid data, shouldn't get added to the dataset
	data = MemsData{
		EngineRPM:                invalidRPM,
		CoolantTemp:              invalidCoolantTemperature,
		IntakeAirTemp:            invalidIntakeTemperature,
		IdleBasePosition:         invalidIdleBasePosition,
		DTC5:                     invalidDTC5,
		BatteryVoltage:           goodBattery,
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
		EngineRPM:                engineStopped,
		CoolantTemp:              coldEngineTemperature,
		IntakeAirTemp:            goodIntakeTemperature,
		IdleBasePosition:         highIdleBasePosition,
		DTC5:                     expectedDTC5,
		BatteryVoltage:           lowBattery,
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

	data = MemsData{
		Time:                     "12:00:00.000",
		EngineRPM:                engineRunning,
		CoolantTemp:              70,
		IntakeAirTemp:            goodIntakeTemperature,
		IdleBasePosition:         goodIdleBasePosition,
		DTC5:                     expectedDTC5,
		BatteryVoltage:           goodBattery,
		CoolantTempSensorFault:   false,
		IntakeAirTempSensorFault: false,
		FuelPumpCircuitFault:     false,
		ThrottlePotCircuitFault:  false,
	}

	d.Analyse(data)
	expectedWarmTime, _ := time.Parse("15:04:05.000", "12:01:50.000")
	then.AssertThat(t, d.expectedTimeEngineWarm, is.EqualTo(expectedWarmTime))
}

func Test_addToDataset(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM: engineRunning,
	}

	d.addToDataset(data)

	then.AssertThat(t, len(d.dataset), is.EqualTo(1))
}

func Test_setDatasetLength(t *testing.T) {
	// try a 0 length dataset, should default to a miniumum of 1
	d := NewDataframeAnalysis(0)
	then.AssertThat(t, d.datasetLength, is.EqualTo(1))

	d.setDatasetLength(0)
	then.AssertThat(t, d.datasetLength, is.EqualTo(1))
}

func Test_fillAddToDataset(t *testing.T) {
	maxDataframeItems := 30
	d := NewDataframeAnalysis(maxDataframeItems)

	data := MemsData{
		EngineRPM: engineRunning,
	}

	// add the first valid item
	d.addToDataset(data)
	then.AssertThat(t, len(d.dataset), is.EqualTo(1))

	data = MemsData{
		EngineRPM: engineRunning + 1,
	}

	for i := 1; i < maxDataframeItems; i++ {
		d.addToDataset(data)
		then.AssertThat(t, len(d.dataset), is.EqualTo(i+1))
	}

	// add another item, buffer should remain at 30
	d.addToDataset(data)
	then.AssertThat(t, len(d.dataset), is.EqualTo(maxDataframeItems))
	//  first item should be dropped which had an engine rpm of 1
	then.AssertThat(t, d.dataset[0].EngineRPM, is.EqualTo(engineRunning+1))
}

func Test_isValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:        engineRunning,
		CoolantTemp:      warmEngineTemperature,
		IntakeAirTemp:    goodIntakeTemperature,
		IdleBasePosition: goodIdleBasePosition,
		DTC5:             expectedDTC5,
	}

	valid := d.isValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		EngineRPM:        invalidRPM,
		CoolantTemp:      invalidCoolantTemperature,
		IntakeAirTemp:    invalidIntakeTemperature,
		IdleBasePosition: invalidIdleBasePosition,
		DTC5:             invalidDTC5,
	}

	valid = d.isValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_EngineRPMIsValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM: engineRunning,
	}

	valid := d.isEngineRPMValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		EngineRPM: invalidRPM,
	}

	valid = d.isValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_CoolantIsValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTemp: warmEngineTemperature,
	}

	valid := d.isCoolantTempValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		CoolantTemp: invalidCoolantTemperature,
	}

	valid = d.isValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_isIntakeAirTempValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		IntakeAirTemp: goodIntakeTemperature,
	}

	valid := d.isIntakeAirTempValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		IntakeAirTemp: invalidIntakeTemperature,
	}

	valid = d.isIntakeAirTempValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_isIdleBasePositionValid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		IdleBasePosition: goodIdleBasePosition,
	}

	valid := d.isIdleBasePositionValid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		IdleBasePosition: invalidIdleBasePosition,
	}

	valid = d.isIdleBasePositionValid(data)
	then.AssertThat(t, valid, is.False())
}

func Test_isDTC5Valid(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		DTC5: expectedDTC5,
	}

	valid := d.isDTC5Valid(data)
	then.AssertThat(t, valid, is.True())

	data = MemsData{
		DTC5: invalidDTC5,
	}

	valid = d.isDTC5Valid(data)
	then.AssertThat(t, valid, is.False())
}
