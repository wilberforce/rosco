package rosco

import (
	"time"
)

type DataframeAnalysis struct {
	datasetLength int
	dataset       []MemsData
	Analysis      MemsAnalysisReport
}

type AnalysisReport struct {
	IsEngineRunning          bool
	IsEngineWarming          bool
	IsAtOperatingTemp        bool
	IsEngineIdle             bool
	IsEngineIdleFault        bool
	IdleSpeedFault           bool
	IdleHotFault             bool
	IsClosedLoop             bool
	IsThrottleActive         bool
	BatteryFault             bool
	MapFault                 bool
	VacuumFault              bool
	IdleAirControlFault      bool
	IdleAirControlJackFault  bool
	O2SystemFault            bool
	LambdaRangeFault         bool
	LambdaOscillationFault   bool
	ThermostatFault          bool
	CoilFault                bool
	CrankshaftSensorFault    bool
	CoolantTempSensorFault   bool
	IntakeAirTempSensorFault bool
	FuelPumpCircuitFault     bool
	ThrottlePotCircuitFault  bool
	IACPosition              int
}

const (
	minimumDatasetSize          = 1
	defaultIdleThrottleAngle    = 14
	lowestBatteryVoltage        = 13
	highestIdleMAPValue         = 45
	highestIdleCoilTime         = 4
	highestIdleRPM              = 1300
	lowestEngineWarmTemperature = 78
	engineNotRunningRPM         = 0
	expectedDTC5Value           = 255
	maximumEngineRPM            = 6000
	maximumIdleBasePosition     = 250
	maximumAirIntakeTemperature = 80
	maximumCoolantTemperature   = 120
)

func NewDataframeAnalysis(datasetLength int) *DataframeAnalysis {
	df := &DataframeAnalysis{}
	df.setDatasetLength(datasetLength)

	return df
}

func (df *DataframeAnalysis) Analyse(data MemsData) {
	if df.isValid(data) {
		// analyse the current operational state
		df.analyseOperationalStatus(data)

		if df.Analysis.IsEngineRunning {
			// add data to the dataset only if the engine is running
			df.addToDataset(data)
			// detect faults from the operational data
			df.analyseOperationalFaults(data)
		}

		// decode the ecu faults
		df.analyseECUFaults(data)
	}
}

func (df *DataframeAnalysis) analyseOperationalStatus(data MemsData) {
	df.Analysis.IsEngineRunning = df.isEngineRunning(data)
	df.Analysis.IsEngineWarming = df.isEngineWarming(data)
	df.Analysis.IsAtOperatingTemp = df.isEngineWarm(data)
	df.Analysis.IsClosedLoop = df.isLoopClosed(data)
	df.Analysis.IsThrottleActive = df.isThrottleActive(data)
	df.Analysis.IsEngineIdle = df.isEngineIdle(data)
}

func (df *DataframeAnalysis) analyseECUFaults(data MemsData) {
	df.Analysis.CoolantTempSensorFault = df.isCoolantSensorFaulty(data)
	df.Analysis.FuelPumpCircuitFault = df.isFuelPumpCircuitFaulty(data)
	df.Analysis.ThrottlePotCircuitFault = df.isThrottlePotCircuitFaulty(data)
	df.Analysis.IntakeAirTempSensorFault = df.isIntakeAirTempSensorFaulty(data)
}

func (df *DataframeAnalysis) analyseOperationalFaults(data MemsData) {
	df.Analysis.BatteryFault = df.isBatteryVoltageLow(data)
	df.Analysis.CoilFault = df.isCoilFaulty(data)
}

func (df *DataframeAnalysis) addToDataset(data MemsData) {
	df.dataset = append(df.dataset, data)

	// shift the data in the buffer
	if len(df.dataset) > df.datasetLength {
		df.dataset = df.dataset[1:]
	}
}

// inspect the current dataframe and return if the frane is valid {
func (df *DataframeAnalysis) isValid(data MemsData) bool {
	return df.isEngineRPMValid(data) &&
		df.isCoolantTempValid(data) &&
		df.isIntakeAirTempValid(data) &&
		df.isIdleBasePositionValid(data) &&
		df.isDTC5Valid(data)
}

func (df *DataframeAnalysis) setDatasetLength(datasetLength int) {
	if datasetLength < minimumDatasetSize {
		df.datasetLength = minimumDatasetSize
	}

	df.datasetLength = datasetLength
}

func (df *DataframeAnalysis) getExpectedEngineWarmTime(data MemsData) time.Time {
	// the engine warms at around 11 seconds per degree
	// given the current time and coolant temp, the estimated warm time to 80C can be calculated
	currentTime, _ := time.Parse("15:04:05.000", data.Time)
	degreesToWarm := engineOperatingTemp - data.CoolantTemp
	secondsToWarm := time.Duration(degreesToWarm * warmingFactor)
	warmAt := currentTime.Add(time.Second * secondsToWarm)

	return warmAt
}

func (df *DataframeAnalysis) isEngineRPMValid(data MemsData) bool {
	return data.EngineRPM < maximumEngineRPM
}

func (df *DataframeAnalysis) isCoolantTempValid(data MemsData) bool {
	return data.CoolantTemp < maximumCoolantTemperature
}

func (df *DataframeAnalysis) isIntakeAirTempValid(data MemsData) bool {
	return data.IntakeAirTemp < maximumAirIntakeTemperature
}

func (df *DataframeAnalysis) isIdleBasePositionValid(data MemsData) bool {
	return data.IdleBasePosition < maximumIdleBasePosition
}

func (df *DataframeAnalysis) isDTC5Valid(data MemsData) bool {
	// observed behaviour is that when dtc5 changes from 255, a number of parameters change that can alter
	// the df, for example jack count leaps to 125
	// since this behaviour is not yet understood, we'll remove these entries
	return data.DTC5 == expectedDTC5Value
}

func (df *DataframeAnalysis) isEngineRunning(data MemsData) bool {
	return data.EngineRPM > engineNotRunningRPM
}

func (df *DataframeAnalysis) isEngineWarming(data MemsData) bool {
	return data.CoolantTemp < lowestEngineWarmTemperature
}

func (df *DataframeAnalysis) isEngineIdle(data MemsData) bool {
	// engine is deemed to be at idle if the engine is running
	// and the angle of the throttle pot indicates the throttle is off
	// later MEMS ECUs use the throttle pot to determine the idle position
	return df.isEngineRunning(data) &&
		data.ThrottleAngle <= defaultIdleThrottleAngle
}

func (df *DataframeAnalysis) isEngineWarm(data MemsData) bool {
	return data.CoolantTemp >= lowestEngineWarmTemperature
}

func (df *DataframeAnalysis) isBatteryVoltageLow(data MemsData) bool {
	return data.BatteryVoltage < lowestBatteryVoltage
}

func (df *DataframeAnalysis) isLoopClosed(data MemsData) bool {
	return data.ClosedLoop
}

func (df *DataframeAnalysis) isThrottleActive(data MemsData) bool {
	return data.ThrottleAngle > defaultIdleThrottleAngle || data.EngineRPM > highestIdleRPM
}

func (df *DataframeAnalysis) isIntakeAirTempSensorFaulty(data MemsData) bool {
	return data.IntakeAirTempSensorFault
}

func (df *DataframeAnalysis) isThrottlePotCircuitFaulty(data MemsData) bool {
	return data.ThrottlePotCircuitFault
}

func (df *DataframeAnalysis) isFuelPumpCircuitFaulty(data MemsData) bool {
	return data.FuelPumpCircuitFault
}

func (df *DataframeAnalysis) isCoolantSensorFaulty(data MemsData) bool {
	return data.CoolantTempSensorFault
}

func (df *DataframeAnalysis) isCoilFaulty(data MemsData) bool {
	// battery must not be low as this will affect the coil timing
	return !df.isBatteryVoltageLow(data) &&
		data.CoilTime > highestIdleCoilTime
}

func (df *DataframeAnalysis) isMAPHigh(data MemsData) bool {
	// MAP value should be less than 45kPa when the engine is at idle
	return df.isEngineIdle(data) &&
		data.ManifoldAbsolutePressure > highestIdleMAPValue
}
