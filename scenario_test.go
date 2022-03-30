package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_scenario_GetScenarios(t *testing.T) {
	s, err := GetScenarios()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(s), is.GreaterThan(0))
}

func Test_scenario_GetScenario(t *testing.T) {
	s := GetScenario("testdata/nofaults.csv")
	then.AssertThat(t, s.Name, is.EqualTo("testdata/nofaults.csv"))
}

func Test_scenario_GetScenarioFCR(t *testing.T) {
	s := GetScenario("testdata/nofaults.fcr")
	then.AssertThat(t, s.Name, is.EqualTo("testdata/nofaults.fcr"))
}
