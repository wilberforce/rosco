package rosco

import (
	"math"
	"time"
)

type DataframeAnalysis struct {
	datasetLength          int
	dataset                []MemsData
	expectedTimeEngineWarm time.Time
	Analysis               AnalysisReport
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
}

const (
	timeFormat                         = "15:04:05.000"
	secondsPerDegree                   = 11
	minimumDatasetSize                 = 1
	defaultIdleThrottleAngle           = 14
	lowestBatteryVoltage               = 13
	highestIdleMAPValue                = 45
	highestIdleCoilTime                = 4
	highestIdleRPM                     = 1300
	lowestEngineWarmTemperature        = 78
	engineNotRunningRPM                = 0
	expectedDTC5Value                  = 255
	maximumEngineRPM                   = 6000
	maximumIdleBasePosition            = 250
	maximumAirIntakeTemperature        = 80
	maximumCoolantTemperature          = 120
	maximumIdleOffset                  = 50
	minimumIdleHot                     = 10
	maximumIdleHot                     = 55
	lowestIdleBasePosition             = 45
	highestIdleBasePosition            = 55
	invalidIACPosition                 = 0
	invalidCASPosition                 = 0
	lowestLambdaValue                  = 10
	highestLambdaValue                 = 900
	highestJackCount                   = 50
	lambdaOscillationStandardDeviation = 100
	highestIdleSpeedDeviation          = 150
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
			// set the expected time for the engine to reach operating temperature
			if df.expectedTimeEngineWarm.IsZero() {
				df.expectedTimeEngineWarm = df.getExpectedEngineWarmTime(data)
			}
		}

		// decode the ecu faults
		df.analyseECUFaults(data)
	}
}

func (df *DataframeAnalysis) analyseOperationalStatus(data MemsData) {
	df.Analysis.IsEngineRunning = df.isEngineRunning(data)
	df.Analysis.IsEngineWarming = df.isEngineWarming(data)
	df.Analysis.IsAtOperatingTemp = df.isEngineWarm(data)
	df.Analysis.IsEngineIdle = df.isEngineIdle(data)
	df.Analysis.IsClosedLoop = df.isLoopClosed(data)
	df.Analysis.IsThrottleActive = df.isThrottleActive(data)
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
	df.Analysis.MapFault = df.isMAPHigh(data)
	df.Analysis.O2SystemFault = !df.isO2SystemActive(data)
	df.Analysis.IsEngineIdleFault = df.isEngineIdleFaulty(data)
	df.Analysis.IdleHotFault = df.isHotIdleFaulty(data)
	df.Analysis.IdleAirControlFault = df.isIACFaulty(data)
	df.Analysis.VacuumFault = df.isVacuumFaulty(data)
	df.Analysis.LambdaRangeFault = df.isLambdaOutOfRange(data)
	df.Analysis.IdleAirControlJackFault = df.isJackCountHigh(data)
	df.Analysis.CrankshaftSensorFault = df.isCrankshaftSensorFaulty(data)
	df.Analysis.LambdaOscillationFault = !df.isLambdaOscillating(data)
	df.Analysis.ThermostatFault = df.isThermostatFaulty(data)
	df.Analysis.IdleSpeedFault = df.isIdleSpeedFaulty(data)
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
	} else {
		df.datasetLength = datasetLength
	}
}

func (df *DataframeAnalysis) isThermostatFaulty(data MemsData) bool {
	if df.isEngineRunning(data) {
		currentTime, _ := time.Parse(timeFormat, data.Time)

		return currentTime.After(df.expectedTimeEngineWarm) &&
			data.CoolantTemp < lowestEngineWarmTemperature
	} else {
		return false
	}
}

func (df *DataframeAnalysis) getExpectedEngineWarmTime(data MemsData) time.Time {
	// the engine warms at around 11 seconds per degree
	// given the current time and coolant temp, the estimated warm time to 80C can be calculated
	currentTime, _ := time.Parse(timeFormat, data.Time)
	degreesToWarm := engineOperatingTemp - data.CoolantTemp
	secondsToWarm := time.Duration(degreesToWarm * secondsPerDegree)
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
	if df.isEngineRunning(data) {
		return !df.isBatteryVoltageLow(data) &&
			data.CoilTime > highestIdleCoilTime
	} else {
		return false
	}
}

func (df *DataframeAnalysis) isMAPHigh(data MemsData) bool {
	// MAP value should be less than 45kPa when the engine is at idle
	return df.isEngineIdle(data) &&
		data.ManifoldAbsolutePressure > highestIdleMAPValue
}

func (df *DataframeAnalysis) isO2SystemActive(data MemsData) bool {
	return data.LambdaStatus == 1
}

func (df *DataframeAnalysis) isEngineIdleFaulty(data MemsData) bool {
	// This is the number of steps from 0 which the ECU will use as guide for starting idle speed control during engine warm up.
	// The value will start at quite a high value (>100 steps) on a very cold engine and fall to < 50 steps on a fully warm engine.
	// A high value on a fully warm engine or a low value on a cold engine will cause poor idle speed control.
	// Idle run line position is calculated by the ECU using the engine coolant temperature sensor.

	if df.isEngineIdle(data) {
		if df.isEngineWarm(data) {
			// fault if > 50 when engine is warm
			return data.IdleBasePosition > highestIdleBasePosition
		} else {
			// fault if < 50 when engine is cold
			return data.IdleBasePosition < lowestIdleBasePosition
		}
	}

	return false
}

func (df *DataframeAnalysis) isHotIdleFaulty(data MemsData) bool {
	// fault if idle hot is outside the range of 10 - 50
	if df.isEngineIdle(data) {
		if df.isEngineWarm(data) {
			return data.IdleHot < minimumIdleHot || data.IdleHot > maximumIdleHot
		}
	}

	return false
}

// Also known as stepper motor idle air control valve (IACV) to control engine idle speed and air flow from cold start up
// A high number of steps indicates that the ECU is attempting to close the stepper or reduce the airflow
// a low number would indicate the inability to increase airflow
// IAC position invalid if the idle offset exceeds the max error, yet the IAC Position remains at 0
func (df *DataframeAnalysis) isIACFaulty(data MemsData) bool {
	return df.isEngineIdle(data) &&
		data.IdleSpeedOffset > maximumIdleOffset && data.IACPosition == invalidIACPosition
}

func (df *DataframeAnalysis) isVacuumFaulty(data MemsData) bool {
	return df.isEngineIdle(data) &&
		data.ManifoldAbsolutePressure > highestIdleMAPValue
}

func (df *DataframeAnalysis) isLambdaOutOfRange(data MemsData) bool {
	return df.isEngineRunning(data) &&
		data.LambdaVoltage < lowestLambdaValue || data.LambdaVoltage > highestLambdaValue
}

// the jack count indicates the number of times the ECU has had to re-learn
// the relationship between the stepper position and the throttle position.
// If this count is high or increments each time the ignition is turned off,
// then there may be a problem with the stepper motor, throttle cable adjustment or the throttle pot.
// The count is increased for each journey with no closed throttle, indicating a throttle adjustment problem.
func (df *DataframeAnalysis) isJackCountHigh(data MemsData) bool {
	return data.JackCount >= highestJackCount
}

func (df *DataframeAnalysis) isCrankshaftSensorFaulty(data MemsData) bool {
	return data.CrankshaftPositionSensor == invalidCASPosition
}

// check if the lambda voltage is oscillating (std dev > 100)
// need a full dataset for this analysis
func (df *DataframeAnalysis) isLambdaOscillating(data MemsData) bool {
	var sum, mean, stddev, count float64

	if df.isEngineRunning(data) {
		if len(df.dataset) == df.datasetLength {
			count = float64(df.datasetLength)

			for i := 0; i < df.datasetLength; i++ {
				sum += float64(df.dataset[i].LambdaVoltage)
			}

			mean = sum / count

			for i := 0; i < df.datasetLength; i++ {
				stddev += math.Pow(float64(df.dataset[i].LambdaVoltage)-mean, 2)
			}

			stddev = math.Sqrt(stddev / (count - 1))

			return stddev > lambdaOscillationStandardDeviation
		}
	}

	return true
}

// a mean value of more than 100 RPM indicates that the ECU is not in control of the idle speed.
// This indicates a possible fault condition.
// need a full dataset for this analysis
func (df *DataframeAnalysis) isIdleSpeedFaulty(data MemsData) bool {
	var sum, mean, count float64

	if df.isEngineRunning(data) {
		if len(df.dataset) == df.datasetLength {
			count = float64(df.datasetLength)

			for i := 0; i < df.datasetLength; i++ {
				sum += float64(df.dataset[i].IdleBasePosition)
			}

			mean = sum / count

			return mean > highestIdleSpeedDeviation
		}
	}

	return false
}
