package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_scenarioConnectAndInitialise(t *testing.T) {
	r := NewScenarioReader()

	then.AssertThat(t, r, is.True())
}

func Test_scenario_ConnectAndInitialise(t *testing.T) {}
func Test_scenario_Disconnect(t *testing.T)           {}
func Test_scenario_GetStatus(t *testing.T)            {}
func Test_scenario_GetDataframes(t *testing.T)        {}
func Test_scenario_Reset(t *testing.T)                {}
func Test_scenario_ClearFaults(t *testing.T)          {}
func Test_scenario_Adjust(t *testing.T)               {}
func Test_scenario_Actuate(t *testing.T)              {}
