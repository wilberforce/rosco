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

func Test_analyseOperationalStatus(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:     engineStopped,
		CoolantTemp:   coldEngineTemperature,
		ClosedLoop:    false,
		ThrottleAngle: idleThrottleAngle,
	}

	d.analyseOperationalStatus(data)
	then.AssertThat(t, d.Analysis.IsEngineRunning, is.False())
	then.AssertThat(t, d.Analysis.IsEngineWarming, is.True())
	then.AssertThat(t, d.Analysis.IsAtOperatingTemp, is.False())
	then.AssertThat(t, d.Analysis.IsClosedLoop, is.False())
	then.AssertThat(t, d.Analysis.IsThrottleActive, is.False())
	then.AssertThat(t, d.Analysis.IsEngineIdle, is.False())

	data = MemsData{
		EngineRPM:     engineRunning,
		CoolantTemp:   warmEngineTemperature,
		ClosedLoop:    true,
		ThrottleAngle: activeThrottleAngle,
	}

	d.analyseOperationalStatus(data)
	then.AssertThat(t, d.Analysis.IsEngineRunning, is.True())
	then.AssertThat(t, d.Analysis.IsEngineWarming, is.False())
	then.AssertThat(t, d.Analysis.IsAtOperatingTemp, is.True())
	then.AssertThat(t, d.Analysis.IsClosedLoop, is.True())
	then.AssertThat(t, d.Analysis.IsThrottleActive, is.True())
	then.AssertThat(t, d.Analysis.IsEngineIdle, is.False())

	data = MemsData{
		EngineRPM:     engineRunning,
		CoolantTemp:   warmEngineTemperature,
		ClosedLoop:    true,
		ThrottleAngle: idleThrottleAngle,
	}

	d.analyseOperationalStatus(data)
	then.AssertThat(t, d.Analysis.IsEngineRunning, is.True())
	then.AssertThat(t, d.Analysis.IsEngineWarming, is.False())
	then.AssertThat(t, d.Analysis.IsAtOperatingTemp, is.True())
	then.AssertThat(t, d.Analysis.IsClosedLoop, is.True())
	then.AssertThat(t, d.Analysis.IsThrottleActive, is.False())
	then.AssertThat(t, d.Analysis.IsEngineIdle, is.True())
}

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

func Test_analyseOperationalFaults(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:                engineRunning,
		BatteryVoltage:           lowBattery,
		ManifoldAbsolutePressure: highIdleMAP,
		CoilTime:                 highCoilTime,
		LambdaStatus:             inactiveLambdaStatus,
		CoolantTemp:              coldEngineTemperature,
		IdleBasePosition:         goodIdleBasePosition,
		LambdaVoltage:            highestLambdaValue + 1,
		JackCount:                highestJackCount - 1,
		CrankshaftPositionSensor: goodCASPosition,
	}

	d.analyseOperationalFaults(data)
	then.AssertThat(t, d.Analysis.BatteryFault, is.True())
	then.AssertThat(t, d.Analysis.MapFault, is.True())
	then.AssertThat(t, d.Analysis.CoilFault, is.False())
	then.AssertThat(t, d.Analysis.O2SystemFault, is.True())
	then.AssertThat(t, d.Analysis.IsEngineIdleFault, is.True())
	then.AssertThat(t, d.Analysis.VacuumFault, is.True())
	then.AssertThat(t, d.Analysis.LambdaRangeFault, is.True())
	then.AssertThat(t, d.Analysis.IdleAirControlJackFault, is.False())
	then.AssertThat(t, d.Analysis.CrankshaftSensorFault, is.False())

	data = MemsData{
		EngineRPM:                engineRunning,
		BatteryVoltage:           goodBattery,
		ManifoldAbsolutePressure: goodIdleMap,
		CoilTime:                 highCoilTime,
		CoolantTemp:              warmEngineTemperature,
		IdleHot:                  lowIdleHot,
		IdleSpeedOffset:          maximumIdleOffset + 1,
		IACPosition:              invalidIACPosition,
		LambdaVoltage:            goodLambdaValue,
		JackCount:                highestJackCount,
		CrankshaftPositionSensor: invalidCASPosition,
	}

	d.analyseOperationalFaults(data)
	then.AssertThat(t, d.Analysis.BatteryFault, is.False())
	then.AssertThat(t, d.Analysis.MapFault, is.False())
	then.AssertThat(t, d.Analysis.CoilFault, is.True())
	then.AssertThat(t, d.Analysis.IdleHotFault, is.True())
	then.AssertThat(t, d.Analysis.IdleAirControlFault, is.True())
	then.AssertThat(t, d.Analysis.VacuumFault, is.False())
	then.AssertThat(t, d.Analysis.LambdaRangeFault, is.False())
	then.AssertThat(t, d.Analysis.IdleAirControlJackFault, is.True())
	then.AssertThat(t, d.Analysis.CrankshaftSensorFault, is.True())

	data = MemsData{
		EngineRPM:                engineRunning,
		BatteryVoltage:           goodBattery,
		ManifoldAbsolutePressure: goodIdleMap,
		CoilTime:                 goodCoilTime,
		LambdaStatus:             activeLambdaStatus,
	}

	d.analyseOperationalFaults(data)
	then.AssertThat(t, d.Analysis.BatteryFault, is.False())
	then.AssertThat(t, d.Analysis.MapFault, is.False())
	then.AssertThat(t, d.Analysis.CoilFault, is.False())
	then.AssertThat(t, d.Analysis.O2SystemFault, is.False())
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
		EngineRPM: engineStopped,
	}

	running := d.isEngineRunning(data)
	then.AssertThat(t, running, is.False())

	data = MemsData{
		EngineRPM: engineRunning,
	}

	running = d.isEngineRunning(data)
	then.AssertThat(t, running, is.True())
}

func Test_isEngineWarming(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTemp: coldEngineTemperature,
	}

	result := d.isEngineWarming(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		CoolantTemp: warmEngineTemperature,
	}

	result = d.isEngineWarming(data)
	then.AssertThat(t, result, is.False())
}

func Test_isEngineWarm(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CoolantTemp: coldEngineTemperature,
	}

	result := d.isEngineWarm(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		CoolantTemp: warmEngineTemperature,
	}

	result = d.isEngineWarm(data)
	then.AssertThat(t, result, is.True())
}

func Test_isBatteryVoltageLow(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		BatteryVoltage: lowBattery,
	}

	result := d.isBatteryVoltageLow(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		BatteryVoltage: goodBattery,
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
		ThrottleAngle: idleThrottleAngle,
	}

	result := d.isThrottleActive(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		ThrottleAngle: activeThrottleAngle,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM: rpmIdle,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		EngineRPM: rpmCruising,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		ThrottleAngle: activeThrottleAngle,
		EngineRPM:     rpmIdle,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		ThrottleAngle: idleThrottleAngle,
		EngineRPM:     rpmCruising,
	}

	result = d.isThrottleActive(data)
	then.AssertThat(t, result, is.True())
}

func Test_isEngineIdle(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		ThrottleAngle: idleThrottleAngle,
		EngineRPM:     engineRunning,
	}

	result := d.isEngineIdle(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM:     engineRunning,
		ThrottleAngle: activeThrottleAngle,
	}

	result = d.isEngineIdle(data)
	then.AssertThat(t, result, is.False())
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

func Test_isCoilFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:      engineRunning,
		BatteryVoltage: goodBattery,
		CoilTime:       highCoilTime,
	}

	result := d.isCoilFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM:      engineRunning,
		BatteryVoltage: goodBattery,
		CoilTime:       goodCoilTime,
	}

	result = d.isCoilFaulty(data)
	then.AssertThat(t, result, is.False())

	// battery low, high coil time ignored
	data = MemsData{
		EngineRPM:      engineRunning,
		BatteryVoltage: lowBattery,
		CoilTime:       highCoilTime,
	}

	result = d.isCoilFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isMAPHigh(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:                engineRunning,
		ManifoldAbsolutePressure: goodIdleMap,
	}

	result := d.isMAPHigh(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		EngineRPM:                engineRunning,
		ManifoldAbsolutePressure: highIdleMAP,
	}

	result = d.isMAPHigh(data)
	then.AssertThat(t, result, is.True())
}

func Test_isO2SystemActive(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		LambdaStatus: inactiveLambdaStatus,
	}

	result := d.isO2SystemActive(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		LambdaStatus: activeLambdaStatus,
	}

	result = d.isO2SystemActive(data)
	then.AssertThat(t, result, is.True())
}

func Test_isEngineIdleFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	// engine off, no fault
	data := MemsData{
		EngineRPM:        engineStopped,
		CoolantTemp:      coldEngineTemperature,
		IdleBasePosition: highIdleBasePosition,
	}

	result := d.isEngineIdleFaulty(data)
	then.AssertThat(t, result, is.False())

	// idle below operating temp., no fault
	data = MemsData{
		EngineRPM:        engineRunning,
		CoolantTemp:      coldEngineTemperature,
		IdleBasePosition: highIdleBasePosition,
	}

	result = d.isEngineIdleFaulty(data)
	then.AssertThat(t, result, is.False())

	// idle below operating temp., faulty
	data = MemsData{
		EngineRPM:        engineRunning,
		CoolantTemp:      coldEngineTemperature,
		IdleBasePosition: goodIdleBasePosition,
	}

	result = d.isEngineIdleFaulty(data)
	then.AssertThat(t, result, is.True())

	// idle at operating temp., no fault
	data = MemsData{
		EngineRPM:        engineRunning,
		CoolantTemp:      warmEngineTemperature,
		IdleBasePosition: goodIdleBasePosition,
	}

	result = d.isEngineIdleFaulty(data)
	then.AssertThat(t, result, is.False())

	// idle at operating temp., fault
	data = MemsData{
		EngineRPM:        engineRunning,
		CoolantTemp:      warmEngineTemperature,
		IdleBasePosition: highIdleHot,
	}

	result = d.isEngineIdleFaulty(data)
	then.AssertThat(t, result, is.True())
}

func Test_isHotIdleFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	// engine cold, no fault
	data := MemsData{
		EngineRPM:   engineStopped,
		CoolantTemp: coldEngineTemperature,
	}

	result := d.isHotIdleFaulty(data)
	then.AssertThat(t, result, is.False())

	// engine warm, hot idle low, fault
	data = MemsData{
		EngineRPM:   1,
		CoolantTemp: 80,
		IdleHot:     5,
	}

	result = d.isHotIdleFaulty(data)
	then.AssertThat(t, result, is.True())

	// engine warm, hot idle high, fault
	data = MemsData{
		EngineRPM:   engineRunning,
		CoolantTemp: warmEngineTemperature,
		IdleHot:     highIdleHot,
	}

	result = d.isHotIdleFaulty(data)
	then.AssertThat(t, result, is.True())

	// engine warm, hot idle normal, no fault
	data = MemsData{
		EngineRPM:   engineRunning,
		CoolantTemp: warmEngineTemperature,
		IdleHot:     goodIdleHot,
	}

	result = d.isHotIdleFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isVacuumFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	// engine idle, high MAP, fault
	data := MemsData{
		EngineRPM:                engineRunning,
		ManifoldAbsolutePressure: highIdleMAP,
	}

	result := d.isVacuumFaulty(data)
	then.AssertThat(t, result, is.True())

	// engine idle, good MAP, no fault
	data = MemsData{
		EngineRPM:                engineRunning,
		ManifoldAbsolutePressure: goodIdleMap,
	}

	result = d.isVacuumFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isLambdaOutofRange(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: highestLambdaValue + 1,
	}

	result := d.isLambdaOutOfRange(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: lowestLambdaValue - 1,
	}

	result = d.isLambdaOutOfRange(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: goodLambdaValue,
	}

	result = d.isLambdaOutOfRange(data)
	then.AssertThat(t, result, is.False())
}

func Test_isJackCountHigh(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		JackCount: highestJackCount - 1,
	}

	result := d.isJackCountHigh(data)
	then.AssertThat(t, result, is.False())

	data = MemsData{
		JackCount: highestJackCount,
	}

	result = d.isJackCountHigh(data)
	then.AssertThat(t, result, is.True())
}

func Test_isCrankshaftPositionSensorFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		CrankshaftPositionSensor: invalidCASPosition,
	}

	result := d.isCrankshaftSensorFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		CrankshaftPositionSensor: goodCASPosition,
	}

	result = d.isCrankshaftSensorFaulty(data)
	then.AssertThat(t, result, is.False())
}

func Test_isLambdaOscillating(t *testing.T) {
	d := NewDataframeAnalysis(3)

	data := MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: lowestLambdaValue,
	}

	d.addToDataset(data)

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: goodLambdaValue,
	}

	d.addToDataset(data)

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: highestLambdaValue,
	}

	d.addToDataset(data)

	result := d.isLambdaOscillating(data)
	then.AssertThat(t, result, is.True())

	d = NewDataframeAnalysis(3)

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: goodLambdaValue - 50,
	}

	d.addToDataset(data)

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: goodLambdaValue,
	}

	d.addToDataset(data)

	data = MemsData{
		EngineRPM:     engineRunning,
		LambdaVoltage: goodLambdaValue + 50,
	}

	d.addToDataset(data)
	result = d.isLambdaOscillating(data)
	then.AssertThat(t, result, is.False())
}

func Test_isThermostatFaulty(t *testing.T) {
	d := NewDataframeAnalysis(2)
	d.expectedTimeEngineWarm, _ = time.Parse("15:04:05.000", "12:01:50.000")

	// not yet reached warm, no fault
	data := MemsData{
		Time:        "12:00:00.000",
		EngineRPM:   engineRunning,
		CoolantTemp: 70,
	}

	d.addToDataset(data)

	data = MemsData{
		Time:        "12:00:11.000",
		EngineRPM:   engineRunning,
		CoolantTemp: 71,
	}

	result := d.isThermostatFaulty(data)
	then.AssertThat(t, result, is.False())

	d = NewDataframeAnalysis(2)
	d.expectedTimeEngineWarm, _ = time.Parse("15:04:05.000", "12:00:11.000")

	// reached warm, no fault
	data = MemsData{
		Time:        "12:00:00.000",
		EngineRPM:   engineRunning,
		CoolantTemp: lowestEngineWarmTemperature - 1,
	}

	d.addToDataset(data)

	data = MemsData{
		Time:        "12:00:11.000",
		EngineRPM:   engineRunning,
		CoolantTemp: warmEngineTemperature,
	}

	result = d.isThermostatFaulty(data)
	then.AssertThat(t, result, is.False())

	d = NewDataframeAnalysis(2)
	d.expectedTimeEngineWarm, _ = time.Parse("15:04:05.000", "12:01:50.000")

	// not reached warm,  fault
	data = MemsData{
		Time:        "12:00:00.000",
		EngineRPM:   engineRunning,
		CoolantTemp: warmEngineTemperature - 10,
	}

	d.addToDataset(data)

	data = MemsData{
		Time:        "12:01:51.000",
		EngineRPM:   engineRunning,
		CoolantTemp: lowestEngineWarmTemperature - 1,
	}

	result = d.isThermostatFaulty(data)
	then.AssertThat(t, result, is.True())
}

func Test_isIdleSpeedFaulty(t *testing.T) {
	d := NewDataframeAnalysis(1)

	data := MemsData{
		EngineRPM:        engineRunning,
		IdleBasePosition: 200,
	}

	d.addToDataset(data)
	result := d.isIdleSpeedFaulty(data)
	then.AssertThat(t, result, is.True())

	data = MemsData{
		EngineRPM:        engineRunning,
		IdleBasePosition: 50,
	}

	d.addToDataset(data)
	result = d.isIdleSpeedFaulty(data)
	then.AssertThat(t, result, is.False())
}
