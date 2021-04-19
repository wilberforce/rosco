package rosco

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

const (
	minIdleColdRPM        = 900  // Minimum expected RPM when running at Idle when cold
	maxIdleColdRPM        = 1200 // Maximum expected RPM when running at Idle when cold
	minIdleWarmRPM        = 700  // Minimum expected RPM when running at Idle when warm
	maxIdleWarmRPM        = 900  // Maximum expected RPM when running at Idle when warm
	minIdleMap            = 30   // Minimum MAP reading when the engine is running
	maxIdleMap            = 60   // Maximum MAP reading when the engine is running
	minMAPEngineOff       = 95   // Minimum MAP reading when the engine to not running
	engineOperatingTemp   = 80   // Engine is at operating temp when coolant temp > 80C
	bestAFR               = 14.7 // Ideal Air to Fuel ratio
	lambdaLow             = 10   // Lambda minimum operating voltage
	lambdaHigh            = 900  // Lambda maximum operating voltage
	minLambdaOscillations = 2    // Minimum number of oscillations in the lambda voltage
	maxIdleError          = 50   // Max Idle Error
	maxDataset            = 30   // Max items to store in the running dataset
	minReadings           = 20   // Minimum number of readings before evaluation of data changes
	minIAC                = 30   // Minimum normal operation steps for the IAC / Stepper Motor
	maxIAC                = 160  // Maximum normal operation steps for the IAC / Stepper Motor
	maxJackCount          = 200  // IAC increment attempts
	warmingFactor         = 11   // allow 11 seconds per degree to warm up to operating temperature
)

// MemsAnalysisReport is the output from running the analysis
type MemsAnalysisReport struct {
	IsEngineRunning          bool
	IsEngineWarming          bool
	IsAtOperatingTemp        bool
	IsEngineIdle             bool
	IsEngineIdleFault        bool
	IdleSpeedFault           bool
	IdleErrorFault           bool
	IdleHotFault             bool
	IdleBaseFault            bool
	IsCruising               bool
	IsClosedLoop             bool
	IsClosedLoopExpected     bool
	ClosedLoopFault          bool
	IsThrottleActive         bool
	MapFault                 bool
	VacuumFault              bool
	IdleAirControlFault      bool
	IdleAirControlRangeFault bool
	IdleAirControlJackFault  bool
	O2SystemFault            bool
	LambdaRangeFault         bool
	LambdaOscillationFault   bool
	ThermostatFault          bool
	CoolantTempSensorFault   bool
	IntakeAirTempSensorFault bool
	FuelPumpCircuitFault     bool
	ThrottlePotCircuitFault  bool
	CrankshaftSensorFault    bool
	CoilFault                bool
	IACPosition              int
}

const DiagnosticsCSVHeader = "engine_running,warming,at_operating_temp,engine_idle,idle_fault,idle_speed_fault,idle_error_fault,idle_hot_fault," +
	"cruising,closed_loop,closed_loop_expected,closed_loop_fault,throttle_active,map_fault,vacuum_fault,iac_fault,iac_range_fault,iac_jack_fault,o2_system_fault," +
	"lambda_range_fault,lambda_oscillation_fault,thermostat_fault,crankshaft_sensor_fault,coil_fault"

// MemsDiagnostics structure
type MemsDiagnostics struct {
	// validDataSet is set once we have a full set of readings in
	// the sample dataset
	validDataSet bool
	// initialData is the first reading
	initialData MemsData
	// currentData is the latest reading
	currentData MemsData
	// dataset contains the last n readings
	dataset []MemsData
	// Stats of the sample
	Stats map[string]Stats
	// Analysis report
	Analysis MemsAnalysisReport
}

// NewMemsDiagnostics generates diagnostic reports
func NewMemsDiagnostics() *MemsDiagnostics {
	diagnostics := &MemsDiagnostics{}
	diagnostics.validDataSet = false
	diagnostics.dataset = []MemsData{}
	diagnostics.Analysis = MemsAnalysisReport{}
	diagnostics.Stats = make(map[string]Stats)
	return diagnostics
}

// Add data to the data set for diagnosis
func (diagnostics *MemsDiagnostics) Add(data MemsData) {
	// add the data to the dataset
	diagnostics.addToDataset(data)
	// update the IAC position
	diagnostics.Analysis.IACPosition = data.IACPosition
}

// Analyse runs a diagnostic review of the dataset
func (diagnostics *MemsDiagnostics) Analyse() {
	if len(diagnostics.dataset) > 1 {
		// get samples and associated stats for named metrics
		diagnostics.Stats["CoolantTemp"] = diagnostics.getMetricStatistics("CoolantTemp")
		diagnostics.Stats["EngineRPM"] = diagnostics.getMetricStatistics("EngineRPM")
		diagnostics.Stats["ManifoldAbsolutePressure"] = diagnostics.getMetricStatistics("ManifoldAbsolutePressure")
		diagnostics.Stats["LambdaVoltage"] = diagnostics.getMetricStatistics("LambdaVoltage")
		diagnostics.Stats["AirFuelRatio"] = diagnostics.getMetricStatistics("AirFuelRatio")
		diagnostics.Stats["IACPosition"] = diagnostics.getMetricStatistics("IACPosition")
		diagnostics.Stats["ThrottleAngle"] = diagnostics.getMetricStatistics("ThrottleAngle")
		diagnostics.Stats["CoilTime"] = diagnostics.getMetricStatistics("CoilTime")
		diagnostics.Stats["IdleError"] = diagnostics.getMetricStatistics("IdleSpeedDeviation")
		diagnostics.Stats["IdleHot"] = diagnostics.getMetricStatistics("IdleHot")

		// apply ECU detected faults
		diagnostics.Analysis.CoolantTempSensorFault = diagnostics.currentData.CoolantTempSensorFault
		diagnostics.Analysis.FuelPumpCircuitFault = diagnostics.currentData.FuelPumpCircuitFault
		diagnostics.Analysis.ThrottlePotCircuitFault = diagnostics.currentData.ThrottlePotCircuitFault
		diagnostics.Analysis.IntakeAirTempSensorFault = diagnostics.currentData.IntakeAirTempSensorFault

		// check the status of the sensors and running parameters
		diagnostics.Analysis.IsEngineRunning = diagnostics.isEngineRunning()

		if diagnostics.Analysis.IsEngineRunning {
			diagnostics.Analysis.IsEngineWarming = diagnostics.isEngineWarming()
			diagnostics.Analysis.IsAtOperatingTemp = diagnostics.isAtOperatingTemperature()
			diagnostics.Analysis.IsEngineIdle = diagnostics.isEngineIdle()
			diagnostics.Analysis.IsCruising = diagnostics.isEngineCruising()
			diagnostics.Analysis.IsThrottleActive = diagnostics.isThrottleActive()
			diagnostics.Analysis.IsClosedLoop = diagnostics.isClosedLoop()

			// perform fault analysis once we have a sample dataset
			if diagnostics.validDataSet {
				diagnostics.Analysis.MapFault = !diagnostics.isMapValid()
				diagnostics.Analysis.O2SystemFault = !diagnostics.isO2SystemWorking()
				diagnostics.Analysis.ThermostatFault = !diagnostics.isThermostatWorking()
				diagnostics.Analysis.IsClosedLoopExpected = diagnostics.isClosedLoopExpected()
				diagnostics.Analysis.LambdaRangeFault = !diagnostics.isLambdaRangeValid()
				diagnostics.Analysis.LambdaOscillationFault = !diagnostics.isLambdaOscillating()
				diagnostics.Analysis.VacuumFault = diagnostics.isVacuumPipeFaulty()
				diagnostics.Analysis.IdleAirControlFault = diagnostics.isIACPositionValid()
				diagnostics.Analysis.IdleAirControlRangeFault = !diagnostics.isIACPositionRangeValid()
				diagnostics.Analysis.IdleAirControlJackFault = diagnostics.isIACJackCountHigh()
				diagnostics.Analysis.IdleSpeedFault = !diagnostics.isEngineIdleSpeedValid()
				diagnostics.Analysis.CrankshaftSensorFault = !diagnostics.isCrankshaftPositionWorking()
				diagnostics.Analysis.CoilFault = !diagnostics.isCoilTimeValid()
				diagnostics.Analysis.IdleErrorFault = !diagnostics.isEngineIdleErrorValid()
				diagnostics.Analysis.IdleHotFault = !diagnostics.isEngineHotIdleValid()
				diagnostics.Analysis.IdleBaseFault = !diagnostics.isIdleBaseValid()
			}
		}

		log.Infof("diagnostics %+v", diagnostics.Analysis)
		log.Infof("stats %+v", diagnostics.Stats)
	} else {
		log.Warnf("No sample data to perform diagnostics")
	}
}

// add the data to the dataset, truncate the data if it exceeds the max values
// we want to store
func (diagnostics *MemsDiagnostics) addToDataset(data MemsData) {
	diagnostics.dataset = append(diagnostics.dataset, data)
	diagnostics.currentData = data

	if len(diagnostics.dataset) == 1 {
		// save the first entry in the dataset
		diagnostics.initialData = data
	}

	// shift the data in the buffer
	if len(diagnostics.dataset) > maxDataset {
		diagnostics.validDataSet = true
		diagnostics.dataset = diagnostics.dataset[1:]
	}
}

// getMetricStatistics takes the sample and calculates the simple average
// this is useful to detect the trend for a metric
func (diagnostics *MemsDiagnostics) getMetricStatistics(metricName string) Stats {
	// get the fields available in the sample
	sampleValues := reflect.ValueOf(diagnostics.dataset)
	// an array to hold the sample
	metricSample := []float64{}

	// iterate the fields and create an array of values for the specific metric only
	for i := 0; i < sampleValues.Len(); i++ {
		sampleValue := sampleValues.Index(i)
		if sampleValue.Kind() == reflect.Struct {
			v := reflect.Indirect(sampleValue).FieldByName(metricName)
			// don't try to create metrics for strings, bools or uints (they're bit patterns)
			switch v.Interface().(type) {
			case int:
				metricSample = append(metricSample, float64(v.Interface().(int)))
			case float32:
				metricSample = append(metricSample, float64(v.Interface().(float32)))
			}
		}
	}

	// calculate the stats for this sample
	stats := *NewStats(metricName, metricSample)
	log.Debugf("stats for %s, %+v", metricName, stats)
	return stats
}

// Given the current engine rpm
// When the rpm is > 0
// Then the engine is running
func (diagnostics *MemsDiagnostics) isEngineRunning() bool {
	return diagnostics.currentData.EngineRPM > 0
}

func (diagnostics *MemsDiagnostics) isCrankshaftPositionWorking() bool {
	count := 0

	for _, v := range diagnostics.dataset {
		if v.CrankshaftPositionSensor {
			count++
		}
	}

	// return true if we have an average of 80% of the values are true
	trend := float32(count / len(diagnostics.dataset))
	return trend > 0.8
}

// Given the engine is running
// When the throttle is not depressed
// Then the throttle is not active
func (diagnostics *MemsDiagnostics) isThrottleActive() bool {
	return diagnostics.Stats["ThrottleAngle"].Mean < 5
}

// Given the engine is running
// When the coil is charged
// Then the coil charge time should be less than 4ms
func (diagnostics *MemsDiagnostics) isCoilTimeValid() bool {
	if diagnostics.isEngineRunning() {
		return diagnostics.Stats["CoilTime"].Mean < 0.4
	}

	// ignore if the engine is not running
	return true
}

// Given the engine is running
// And the sample rpm is stable
// Then the engine is idling
func (diagnostics *MemsDiagnostics) isEngineIdle() bool {
	return diagnostics.isEngineRunning() && !diagnostics.isThrottleActive()
}

// The idle speed at cold and at operating temperature.
// Idle speed outside of this range indicates a fault condition
func (diagnostics *MemsDiagnostics) isEngineIdleSpeedValid() bool {
	if diagnostics.isEngineIdle() {
		if diagnostics.isAtOperatingTemperature() {
			return diagnostics.Stats["EngineRPM"].Mean >= minIdleWarmRPM && diagnostics.Stats["EngineRPM"].Mean <= maxIdleWarmRPM
		} else {
			return diagnostics.Stats["EngineRPM"].Mean >= minIdleColdRPM && diagnostics.Stats["EngineRPM"].Mean <= maxIdleColdRPM
		}
	}

	return true
}

// This is the current difference between the target idle speed set by the MEMS ECU and the actual engine speed.
// A value of more than 100 RPM indicates that the ECU is not in control of the idle speed. This indicates a possible fault condition.
// A quick addition of this value and the current engine RPM will also tell what the value is of the ECU's target Idle Speed.
func (diagnostics *MemsDiagnostics) isEngineIdleErrorValid() bool {
	if diagnostics.isEngineIdle() {
		return diagnostics.Stats["IdleError"].Mean < 150
	}

	return true
}

// This is the number of IACV steps from fully closed (0) which the ECU has learned as the correct position to
// maintain the target idle speed with a fully warmed up engine.
// If this value is outside the range 10 - 50 steps, then this is an indication of a possible fault condition or poor adjustment.
func (diagnostics *MemsDiagnostics) isEngineHotIdleValid() bool {
	if diagnostics.isEngineRunning() && diagnostics.isAtOperatingTemperature() {
		return diagnostics.Stats["IdleHot"].Mean >= 10 && diagnostics.Stats["IdleHot"].Mean <= 50
	}

	return true
}

// This is the number of steps from 0 which the ECU will use as guide for starting idle speed control during engine warm up.
// The value will start at quite a high value (>100 steps) on a very cold engine and fall to < 50 steps on a fully warm engine.
// A high value on a fully warm engine or a low value on a cold engine will cause poor idle speed control.
// Idle run line position is calculated by the ECU using the engine coolant temperature sensor.
func (diagnostics *MemsDiagnostics) isIdleBaseValid() bool {
	if diagnostics.isEngineRunning() {
		if diagnostics.isAtOperatingTemperature() {
			return diagnostics.currentData.IdleBasePosition <= 50
		} else {
			return diagnostics.currentData.IdleBasePosition >= 50
		}
	}

	return true
}

// Given the engine is running
// And the throttle is depressed
// When the sample rpm is above idle
// Then the engine is cruising
func (diagnostics *MemsDiagnostics) isEngineCruising() bool {
	return diagnostics.isThrottleActive() && diagnostics.Stats["EngineRPM"].Mean >= minIdleWarmRPM
}

// Given the engine coolant temperature is stable
// When the temperature is above the operating temperature
// Then the engine is at operating temperature
func (diagnostics *MemsDiagnostics) isAtOperatingTemperature() bool {
	return diagnostics.Stats["CoolantTemp"].Value >= engineOperatingTemp && diagnostics.Stats["CoolantTemp"].Stddev < 5
}

// Given the engine coolant temperature is increasing at > 5%
// When the engine coolant temperature is below the operating temperature
// Then the engine is warming
func (diagnostics *MemsDiagnostics) isEngineWarming() bool {
	if diagnostics.isAtOperatingTemperature() == false {
		return diagnostics.Stats["CoolantTemp"].Trend > 0.05
	}

	return false
}

// The coolant temperature as measured by the ECU and the current temperature change over time
// is used to determine if the thermostat and coolant sensor are working.
// If the sensor is open circuit, a default value of about 60C will be recorded.
// During engine warm up, the value should rise smoothly from ambient to approximately 90C.
// Sensor faults may cause several symptoms including poor starting, fast idle speed, poor fuel consumption
// and cooling fans running continuously.
func (diagnostics *MemsDiagnostics) isThermostatWorking() bool {
	if !diagnostics.isAtOperatingTemperature() {
		startTime, _ := time.Parse("15:04:05.000", diagnostics.initialData.Time)
		currentTime, _ := time.Parse("15:04:05.000", diagnostics.currentData.Time)
		elapsedTime := currentTime.Sub(startTime)

		// evaluate running temperature if sufficient time has passed
		degreesToWarm := engineOperatingTemp - diagnostics.initialData.CoolantTemp

		if elapsedTime.Seconds() > float64(degreesToWarm*warmingFactor) {
			return diagnostics.currentData.CoolantTemp >= engineOperatingTemp
		}
	}

	// could be unlucky and get a starting temp. of 60C so get a false reading
	return diagnostics.initialData.IntakeAirTemp <= 60 && diagnostics.Stats["CoolantTemp"].Mean != 60
}

// Manifold Pressure (KPa): This displays the pressure measured by the external MEMS air pressure sensor.
// Normal reading with the engine not running is approximately 100 KPa
// 30-40 KPa when the engine is idling.
// Very high values may indicate problems with the sensor or a blocked or disconnected vacuum pipe.
// Moderately raised values may indicate mechanical problems with the engine
func (diagnostics *MemsDiagnostics) isMapValid() bool {
	if diagnostics.isEngineRunning() {
		// fault if the map readings are outside of expected when idling
		return diagnostics.Stats["ManifoldAbsolutePressure"].Mean >= minIdleMap && diagnostics.Stats["ManifoldAbsolutePressure"].Mean <= maxIdleMap
	} else {
		// fault if the map is reading low when the engine is off
		return diagnostics.Stats["ManifoldAbsolutePressure"].Mean >= minMAPEngineOff
	}
}

func (diagnostics *MemsDiagnostics) isLambdaOscillating() bool {
	// At 2000 rpm lambda values should oscillate
	if len(diagnostics.dataset) >= minReadings {
		return !(diagnostics.Stats["LambdaVoltage"].Oscillation < minLambdaOscillations)
	}

	return false
}

// Shows the state of MEMS internal diagnostics on the oxygen sensor and its associated wiring.
// A displayed value of 1 indicates no fault. A displayed value of 0 indicates a possible problem.
func (diagnostics *MemsDiagnostics) isO2SystemWorking() bool {
	return diagnostics.currentData.LambdaStatus >= 1
}

// This shows whether the fuelling is being controlled using feedback from the oxygen sensors.
// A value of True indicates that closed loop fuelling is active, a value of False indicates fuelling open loop.
// On a fully warm vehicle, Loop Status should indicate closed loop under most driving and idling conditions.
func (diagnostics *MemsDiagnostics) isClosedLoop() bool {
	return diagnostics.currentData.ClosedLoop
}

// determines whether we're expecting the ECU to use closed loop.
// ECU will generally only use the lambda sensor’s output during two specific conditions
// (a) during idle, ie. when the engine is under no load apart from keeping itself running, and
// (b) during part-load conditions (which we usually term ‘cruising speed’) where the engine is keeping the car at a constant speed.
// Fast idle is typically 2500 - 3000 RPM
// Slow idle is typically  450 - 1500 RPM
// summary : expecting ECU to switch to closed loop when at operating temperature and either idling or cruising
func (diagnostics *MemsDiagnostics) isClosedLoopExpected() bool {
	return diagnostics.isAtOperatingTemperature() && (diagnostics.isEngineIdle() || diagnostics.isEngineCruising())
}

func (diagnostics *MemsDiagnostics) isLambdaRangeValid() bool {
	// lambda voltages too high or too low
	if len(diagnostics.dataset) >= minReadings {
		return diagnostics.Stats["LambdaVoltage"].Min >= lambdaLow && diagnostics.Stats["LambdaVoltage"].Max <= lambdaHigh
	}

	return true
}

// Also known as stepper motor idle air control valve (IACV)
// bolts on the side of the injection body housing to control engine idle speed and air flow from cold start up
// A high number of steps indicates that the ECU is attempting to close the stepper or reduce the airflow
// a low number would indicate the inability to increase airflow
// IAC position invalid if the idle offset exceeds the max error, yet the IAC Position remains at 0
func (diagnostics *MemsDiagnostics) isIACPositionValid() bool {
	return diagnostics.currentData.IdleSpeedDeviation < maxIdleError && diagnostics.currentData.IACPosition > 0
}

func (diagnostics *MemsDiagnostics) isIACPositionRangeValid() bool {
	return diagnostics.Stats["IACPosition"].Mean >= minIAC && diagnostics.Stats["IACPosition"].Mean <= maxIAC
}

func (diagnostics *MemsDiagnostics) isVacuumPipeFaulty() bool {
	// if a hose is split the vacuum sensor in the ECU doesn't see true manifold pressure,
	// but something of a slightly higher absolute pressure (a little closer to atmospheric).
	// The ECU thinks then that the engine is more highly loaded, for the same RPM, than it really is and gives more fuel

	return diagnostics.Stats["ManifoldAbsolutePressure"].Mean >= maxIdleMap && diagnostics.Stats["AirFuelRatio"].Mean > bestAFR
}

func (diagnostics *MemsDiagnostics) isIACJackCountHigh() bool {
	// On systems using a throttle body where the idle air is controlled by a stepper motor which directly acts on the
	// throttle disk (normally metal inlet manifold), the count indicates the number of times the ECU has had to re-learn
	// the relationship between the stepper position and the throttle position.
	// If this count is high or increments each time the ignition is turned off,
	//then there may be a problem with the stepper motor, throttle cable adjustment or the throttle pot.
	// On systems using a plastic throttle body/manifold, the count is a warning that the MEMS ECU has never seen the
	// throttle fully closed.
	// The count is increased for each journey with no closed throttle, indicating a throttle adjustment problem.
	return diagnostics.currentData.JackCount >= maxJackCount
}
