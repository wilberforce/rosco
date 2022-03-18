package rosco

func (df *DataframeAnalysis) analyseECUFaults(data MemsData) {
	df.Analysis.CoolantTempSensorFault = df.isCoolantSensorFaulty(data)
	df.Analysis.FuelPumpCircuitFault = df.isFuelPumpCircuitFaulty(data)
	df.Analysis.ThrottlePotCircuitFault = df.isThrottlePotCircuitFaulty(data)
	df.Analysis.IntakeAirTempSensorFault = df.isIntakeAirTempSensorFaulty(data)
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
