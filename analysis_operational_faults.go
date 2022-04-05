package rosco

import (
	"math"
	"time"
)

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
	df.Analysis.ThermostatFault = df.isThermostatFaulty(data)
	df.Analysis.IdleSpeedFault = df.isIdleSpeedFaulty(data)
	df.Analysis.LambdaOscillationFault = df.isLambdaFaulty(data)
}

func (df *DataframeAnalysis) isBatteryVoltageLow(data MemsData) bool {
	return data.BatteryVoltage < lowestBatteryVoltage
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

// if the engine has been started and 90 seconds have elapsed then
// check if the lambda voltage is oscillating high/low
// if we don't see the voltage changing then the lambda could be faulty
func (df *DataframeAnalysis) isLambdaFaulty(data MemsData) bool {
	if df.isEngineRunning(data) {
		if !df.engineStartedAt.IsZero() {
			currentTime, _ := time.Parse(timeFormat, data.Time)
			startOscillationsAt := df.engineStartedAt.Add(time.Second * 90)
			if currentTime.After(startOscillationsAt) {
				return !df.isLambdaOscillating(data)
			}
		}
	}

	// ignore the lambda and return no fault
	return false
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

			// expect to see oscillations of at least +/-100mV (353mv - 535mv)
			return stddev > lambdaOscillationStandardDeviation
		}
	}

	return true
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
