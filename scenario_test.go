package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_scenario_GetScenarios(t *testing.T) {
	s, err := GetScenarios("./testdata/")

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(s), is.GreaterThan(0))

	then.AssertThat(t, s[0].Count, is.EqualTo(338))
	then.AssertThat(t, s[0].Duration, is.EqualTo("3m 17s"))
}

func Test_scenario_GetScenariosFromLogFolder(t *testing.T) {
	s, err := GetScenarios("")

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(s), is.GreaterThan(0))

	then.AssertThat(t, s[0].Count, is.GreaterThan(1))
	then.AssertThat(t, s[0].Duration, is.Not(is.EqualTo("")))
}

func Test_scenario_GetScenario(t *testing.T) {
	s := GetScenario("testdata/nofaults.csv")
	then.AssertThat(t, s.Name, is.EqualTo("testdata/nofaults.csv"))
	then.AssertThat(t, s.Count, is.EqualTo(338))
}

func Test_scenario_GetScenarioFCR(t *testing.T) {
	s := GetScenario("testdata/nofaults.fcr")
	then.AssertThat(t, s.Name, is.EqualTo("testdata/nofaults.fcr"))
}
