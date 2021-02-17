package rosco

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
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
	warmingFactor         = 11   // allow 11 seconds per degree to warm up to operating temperature
)

// MemsAnalysisReport is the output from running the analysis
type MemsAnalysisReport struct {
	IsEngineRunning          bool
	IsEngineWarming          bool
	IsAtOperatingTemp        bool
	IsEngineIdle             bool
	IsEngineIdleFault        bool
	IsCruising               bool
	IsClosedLoop             bool
	ClosedLoopFault          bool
	MapFault                 bool
	VacuumFault              bool
	IdleAirControlFault      bool
	IACMinFault              bool
	IACMaxFault              bool
	LambdaSensorFault        bool
	LambdaRangeFault         bool
	LambdaOscillationFault   bool
	ThermostatFault          bool
	CoolantTempSensorFault   bool
	IntakeAirTempSensorFault bool
	FuelPumpCircuitFault     bool
	ThrottlePotCircuitFault  bool
	IACPosition              int
}

// MemsDiagnostics structure
type MemsDiagnostics struct {
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

		// apply ECU detected faults
		diagnostics.Analysis.CoolantTempSensorFault = diagnostics.currentData.CoolantTempSensorFault
		diagnostics.Analysis.FuelPumpCircuitFault = diagnostics.currentData.FuelPumpCircuitFault
		diagnostics.Analysis.ThrottlePotCircuitFault = diagnostics.currentData.ThrottlePotCircuitFault
		diagnostics.Analysis.IntakeAirTempSensorFault = diagnostics.currentData.IntakeAirTempSensorFault

		// check the status of the sensors and running parameters
		diagnostics.checkIsEngineRunning()
		diagnostics.checkIsEngineWarm()
		diagnostics.checkIsEngineIdle()
		diagnostics.checkMapSensor()
		diagnostics.checkForExpectedClosedLoop()
		diagnostics.checkIdleAirControl()
		diagnostics.checkLambdaStatus()
		diagnostics.checkForVacuumFault()

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

// IsEngineWarm uses the current engine temperature and the standard deviation in the sample to determine the
// stability of the temperature. If the reading is at the designated thermostat temp (88C) and the std deviation
// is low then deem the engine to be running at operating temperature
func (diagnostics *MemsDiagnostics) checkIsEngineWarm() {
	diagnostics.Analysis.IsAtOperatingTemp = diagnostics.Stats["CoolantTemp"].Value >= engineOperatingTemp && diagnostics.Stats["CoolantTemp"].Stddev < 5

	if !diagnostics.Analysis.IsAtOperatingTemp {
		startTime, _ := time.Parse("15:04:05.000", diagnostics.initialData.Time)
		currentTime, _ := time.Parse("15:04:05.000", diagnostics.currentData.Time)
		elapsedTime := currentTime.Sub(startTime)

		// evaluate running temperature if sufficient time has passed
		degreesToWarm := engineOperatingTemp - diagnostics.initialData.CoolantTemp

		if elapsedTime.Seconds() > float64(degreesToWarm*warmingFactor) {
			if !diagnostics.Analysis.IsAtOperatingTemp {
				// set fault code if engine should be warm
				diagnostics.Analysis.ThermostatFault = true
			}
		}
	}
}

// if the last reading of engine RPM is 0 then the engine is not running
// we don't use the sample set as the engine may have recently been stopped
func (diagnostics *MemsDiagnostics) checkIsEngineRunning() {
	diagnostics.Analysis.IsEngineRunning = !(diagnostics.currentData.EngineRPM == 0)
}

// IsIdle determines the correct idle speed parameters based on whether the engine is warm or cold
// if the RPM is within the parameters for the sample period then the engine is deemed to be at Idle
func (diagnostics *MemsDiagnostics) checkIsEngineIdle() {
	// if the engine is running and the RPM is stable, then deem to be in an idle state
	// if the engine RPM is above idle but stable then we're cruising
	if diagnostics.Analysis.IsEngineRunning && diagnostics.Stats["EngineRPM"].Stddev <= 10 {
		diagnostics.Analysis.IsEngineIdle = true
		diagnostics.Analysis.IsCruising = diagnostics.Stats["EngineRPM"].Mean > maxIdleWarmRPM
	}

	// if the engine is running, and the throttle angle is low then assume we're at idle
	// check the revs are within the tolerances for warm and cold idle speeds
	if len(diagnostics.dataset) >= minReadings && diagnostics.Analysis.IsEngineRunning {
		if diagnostics.Stats["ThrottleAngle"].Mean < 5 {
			if diagnostics.Analysis.IsAtOperatingTemp {
				// use warm idle settings
				diagnostics.Analysis.IsEngineIdleFault = !(diagnostics.Stats["EngineRPM"].Mean >= minIdleWarmRPM && diagnostics.Stats["EngineRPM"].Mean <= maxIdleWarmRPM)
				diagnostics.Analysis.IsEngineWarming = false
			} else {
				// use cold idle settings
				diagnostics.Analysis.IsEngineIdleFault = !(diagnostics.Stats["EngineRPM"].Mean >= minIdleColdRPM && diagnostics.Stats["EngineRPM"].Mean <= maxIdleColdRPM)
				diagnostics.Analysis.IsEngineWarming = true
			}
		}
	}
}

// Manifold Pressure (KPa): This displays the pressure measured by the external MEMS air pressure sensor.
// Normal reading with the engine not running is approximately 100 KPa
// 30-40 KPa when the engine is idling.
// Very high values may indicate problems with the sensor or a blocked or disconnected vacuum pipe.
// Moderately raised values may indicate mechanical problems with the engine
func (diagnostics *MemsDiagnostics) checkMapSensor() {
	if diagnostics.Analysis.IsEngineRunning {
		// only check if engine is running at idle
		if diagnostics.Analysis.IsEngineIdle {
			// fault if the map readings are outside of expected when idling
			diagnostics.Analysis.MapFault = !(diagnostics.Stats["ManifoldAbsolutePressure"].Mean >= minIdleMap && diagnostics.Stats["ManifoldAbsolutePressure"].Mean <= maxIdleMap)
		}
	} else {
		// fault if the map is reading low when the engine is off
		diagnostics.Analysis.MapFault = diagnostics.Stats["ManifoldAbsolutePressure"].Mean < minMAPEngineOff
	}
}

// determines whether we're expecting the ECU to use closed loop.
// ECU will generally only use the lambda sensor’s output during two specific conditions
// (a) during idle, ie. when the engine is under no load apart from keeping itself running, and
// (b) during part-load conditions (which we usually term ‘cruising speed’) where the engine is keeping the car at a constant speed.
// Fast idle is typically 2500 - 3000 RPM
// Slow idle is typically  450 - 1500 RPM
func (diagnostics *MemsDiagnostics) checkForExpectedClosedLoop() {
	diagnostics.Analysis.IsClosedLoop = diagnostics.currentData.ClosedLoop
	// expecting ECU to switch to closed loop when at operating temperature and either idling or cruising
	diagnostics.Analysis.ClosedLoopFault = diagnostics.Analysis.IsAtOperatingTemp && (diagnostics.Analysis.IsEngineIdle || diagnostics.Analysis.IsCruising)
}

// if a hose is split the vacuum sensor in the ECU doesn't see true manifold pressure,
// but something of a slightly higher absolute pressure (a little closer to atmospheric).
// The ECU thinks then that the engine is more highly loaded, for the same RPM, than it really is and gives more fuel
func (diagnostics *MemsDiagnostics) checkForVacuumFault() {
	// wonder if this will be true if the AFR is rich and the MAP reading is high
	diagnostics.Analysis.VacuumFault = diagnostics.Stats["ManifoldAbsolutePressure"].Mean >= maxIdleMap && diagnostics.Stats["AirFuelRatio"].Mean > bestAFR
}

// Also known as stepping motor--idle air control valve (IACV)
// bolts on the side of the injection body housing to control engine idle speed
// and air flow from cold start up
// A high number of steps indicates that the ECU is attempting to close the stepper or reduce the airflow
// a low number would indicate the inability to increase airflow

func (diagnostics *MemsDiagnostics) checkIdleAirControl() {
	if diagnostics.Analysis.IsEngineRunning {
		// IAC fault if the idle offset exceeds the max error, yet the IAC Position remains at 0
		if diagnostics.currentData.IdleSpeedDeviation >= maxIdleError && diagnostics.currentData.IACPosition == 0 {
			diagnostics.Analysis.IdleAirControlFault = true
		}

		// IAC fault if the stepper motor is wide open or closed
		if diagnostics.Stats["IACPosition"].Mean <= minIAC {
			diagnostics.Analysis.IdleAirControlFault = true
			diagnostics.Analysis.IACMinFault = true
		}

		if diagnostics.Stats["IACPosition"].Mean >= maxIAC {
			diagnostics.Analysis.IdleAirControlFault = true
			diagnostics.Analysis.IACMaxFault = true
		}
	} else {
		diagnostics.Analysis.IdleAirControlFault = false
	}
}

//  At 2000 rpm it should be switching rapidly between the minimum and maximum figures as the MEMS controls the engine conditions.????
func (diagnostics *MemsDiagnostics) checkLambdaStatus() {
	// lambda sensor fault is set to 1, indicates an O2 system fault
	if diagnostics.currentData.LambdaStatus > 0 {
		diagnostics.Analysis.LambdaSensorFault = true
	}
	if diagnostics.Analysis.IsEngineRunning && diagnostics.Analysis.IsClosedLoop {
		// lambda voltages too high or too low
		if diagnostics.Stats["LambdaVoltage"].Min <= lambdaLow && diagnostics.Stats["LambdaVoltage"].Max >= lambdaHigh {
			diagnostics.Analysis.LambdaRangeFault = true
		}
	}

	// The lambda voltage should oscillate, if the lambda is static for too long, lambda sensor maybe faulty
	// sample must be a minimum of readings before evaluation
	if len(diagnostics.dataset) >= minReadings {
		if diagnostics.Stats["LambdaVoltage"].Oscillation < minLambdaOscillations {
			diagnostics.Analysis.LambdaSensorFault = true
			diagnostics.Analysis.LambdaOscillationFault = true
		}
	}
}
