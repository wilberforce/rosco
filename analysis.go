package rosco

import (
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
	IACPosition              int // intend to remove this from here
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
	engineOperatingTemp                = 80
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
