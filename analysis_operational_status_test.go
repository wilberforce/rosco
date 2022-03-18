package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
)

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
