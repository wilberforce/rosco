package rosco

import (
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MemsDataLogger logs the mems data to a CSV file
type MemsDataLogger struct {
	file   *os.File
	writer *csv.Writer
	IsOpen bool
}

const MemsDataHeader = "#time," +
	"80x01-02_engine-rpm,80x03_coolant_temp,80x04_ambient_temp,80x05_intake_air_temp,80x06_fuel_temp,80x07_map_kpa,80x08_battery_voltage,80x09_throttle_pot,80x0A_idle_switch,80x0B_uk1," +
	"80x0C_park_neutral_switch,80x0D-0E_fault_codes,80x0F_idle_set_point,80x10_idle_hot,80x11_uk2,80x12_iac_position,80x13-14_idle_error,80x15_ignition_advance_offset,80x16_ignition_advance,80x17-18_coil_time," +
	"80x19_crankshaft_position_sensor,80x1A_uk4,80x1B_uk5," +
	"7dx01_ignition_switch,7dx02_throttle_angle,7dx03_uk6,7dx04_air_fuel_ratio,7dx05_dtc2,7dx06_lambda_voltage,7dx07_lambda_sensor_frequency,7dx08_lambda_sensor_dutycycle,7dx09_lambda_sensor_status,7dx0A_closed_loop," +
	"7dx0B_long_term_fuel_trim,7dx0C_short_term_fuel_trim,7dx0D_carbon_canister_dutycycle,7dx0E_dtc3,7dx0F_idle_base_pos,7dx10_uk7,7dx11_dtc4,7dx12_ignition_advance2,7dx13_idle_speed_offset,7dx14_idle_error2," +
	"7dx14-15_uk10,7dx16_dtc5,7dx17_uk11,7dx18_uk12,7dx19_uk13,7dx1A_uk14,7dx1B_uk15,7dx1C_uk16,7dx1D_uk17,7dx1E_uk18,7dx1F_uk19,0x7d_raw,0x80_raw"

// NewMemsDataLogger logs the mems data to a CSV file
func NewMemsDataLogger(folder string, prefix string) *MemsDataLogger {
	var err error

	datalogger := &MemsDataLogger{}
	filename := getFilename(folder, prefix)

	err = openFile(datalogger, filename)

	if err == nil {
		datalogger.IsOpen = true

		if err = writeFileHeader(datalogger); err != nil {
			log.Errorf("Unable to create header in log file %s (%S)", filename, err)
		}
	} else {
		log.Errorf("Unable to create log file %s (%s)", filename, err)
	}

	return datalogger
}

func writeFileHeader(datalogger *MemsDataLogger) error {
	var err error

	if datalogger.IsOpen {
		// create the header
		header := MemsDataHeader + "," + DiagnosticsCSVHeader
		// write the header to the file
		err = datalogger.writer.Write(strings.Split(header, ","))
	}

	return err
}

func openFile(datalogger *MemsDataLogger, filename string) error {
	var err error

	// create the file
	datalogger.file, err = os.Create(filename)

	if err != nil {
		log.Errorf("unable to create log file %s (%s)", filename, err)
	} else {
		// create a file write for the new file
		createFileWriter(datalogger)
		datalogger.IsOpen = true
	}

	return err
}

func createFileWriter(datalogger *MemsDataLogger) {
	datalogger.writer = csv.NewWriter(datalogger.file)
	defer datalogger.writer.Flush()
}

func (datalogger *MemsDataLogger) WriteMemsDataToFile(memsdata MemsData) {
	if datalogger.IsOpen {
		// convert the memdata into csv fields
		data := convertMemsDataToCSVData(memsdata)

		// write the data
		datalogger.writeMemsDataToLogfile(data)
	}
}

func (datalogger *MemsDataLogger) writeMemsDataToLogfile(data []string) {
	var err error

	err = datalogger.writer.Write(data)
	defer datalogger.writer.Flush()

	if err != nil {
		log.Errorf("Unable to write data to logfile (%s)", err)
	}
}

func (datalogger *MemsDataLogger) Close() {
	var err error

	if datalogger.IsOpen {
		datalogger.IsOpen = false

		err = datalogger.file.Close()

		if err != nil {
			log.Errorf("Error closing logfile (%s)", err)
		}
	}
}

func getFilename(folder string, prefix string) string {
	currentTime := time.Now()
	dateTime := currentTime.Format("2006-01-02 15:04:05")
	dateTime = strings.ReplaceAll(dateTime, ":", "")
	dateTime = strings.ReplaceAll(dateTime, " ", "-")
	dateTime = dateTime[:len(dateTime)-2] + "00"

	filename := fmt.Sprintf("%s/%s-%s.csv", folder, prefix, dateTime)
	return filepath.FromSlash(filename)
}

func convertMemsDataToCSVData(data MemsData) []string {
	s := fmt.Sprintf("%s,"+
		"%d,%d,%d,%d,%d,%.2f,%.2f,%.2f,%t,%t,"+
		"%t,%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f,"+
		"%d,%d,%d,"+
		"%t,%d,%d,%.2f,%d,%d,%d,%d,%d,%t,"+
		"%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,"+
		"%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,"+
		"%s,%s,"+
		"%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t",
		data.Time,
		data.EngineRPM,
		data.CoolantTemp,
		data.AmbientTemp,
		data.IntakeAirTemp,
		data.FuelTemp,
		data.ManifoldAbsolutePressure,
		data.BatteryVoltage,
		data.ThrottlePotSensor,
		data.IdleSwitch,
		data.AirconSwitch,
		data.ParkNeutralSwitch,
		data.DTC0,
		data.IdleSetPoint,
		data.IdleHot,
		data.Uk8011,
		data.IACPosition,
		data.IdleSpeedDeviation,
		data.IgnitionAdvanceOffset80,
		data.IgnitionAdvance,
		data.CoilTime,
		data.CrankshaftPositionSensor,
		data.Uk801a,
		data.Uk801b,
		data.IgnitionSwitch,
		data.ThrottleAngle,
		data.Uk7d03,
		data.AirFuelRatio,
		data.DTC2,
		data.LambdaVoltage,
		data.LambdaFrequency,
		data.LambdaDutycycle,
		data.LambdaStatus,
		data.ClosedLoop,
		data.LongTermFuelTrim,
		data.ShortTermFuelTrim,
		data.CarbonCanisterPurgeValve,
		data.DTC3,
		data.IdleBasePosition,
		data.Uk7d10,
		data.DTC4,
		data.IgnitionAdvanceOffset7d,
		data.IdleSpeedOffset,
		data.Uk7d14,
		data.Uk7d15,
		data.DTC5,
		data.Uk7d17,
		data.Uk7d18,
		data.Uk7d19,
		data.Uk7d1a,
		data.Uk7d1b,
		data.Uk7d1c,
		data.Uk7d1d,
		data.Uk7d1e,
		data.JackCount,
		strings.ToUpper(data.Dataframe7d),
		strings.ToUpper(data.Dataframe80),
		data.Analytics.IsEngineRunning,
		data.Analytics.IsEngineWarming,
		data.Analytics.IsAtOperatingTemp,
		data.Analytics.IsEngineIdle,
		data.Analytics.IsEngineIdleFault,
		data.Analytics.IdleSpeedFault,
		data.Analytics.IdleErrorFault,
		data.Analytics.IdleHotFault,
		data.Analytics.IsCruising,
		data.Analytics.IsClosedLoop,
		data.Analytics.IsClosedLoopExpected,
		data.Analytics.ClosedLoopFault,
		data.Analytics.IsThrottleActive,
		data.Analytics.MapFault,
		data.Analytics.VacuumFault,
		data.Analytics.IdleAirControlFault,
		data.Analytics.IdleAirControlRangeFault,
		data.Analytics.IdleAirControlJackFault,
		data.Analytics.O2SystemFault,
		data.Analytics.LambdaRangeFault,
		data.Analytics.LambdaOscillationFault,
		data.Analytics.ThermostatFault,
		data.Analytics.CrankshaftSensorFault,
		data.Analytics.CoilFault,
	)

	return strings.Split(s, ",")
}
