package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
	"time"
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

func Test_scenario_filterScenarios(t *testing.T) {
	var s []ScenarioDescription

	s = append(s, ScenarioDescription{
		Name:     "file1.csv",
		FileType: "CSV",
		Date:     time.Now(),
	})

	s = append(s, ScenarioDescription{
		Name:     "file2.csv",
		FileType: "CSV",
		Date:     time.Now(),
	})

	s = append(s, ScenarioDescription{
		Name:     "file1.fcr",
		FileType: "FCR",
		Date:     time.Now(),
	})

	filtered := filterScenarios(s)
	then.AssertThat(t, len(filtered), is.EqualTo(2))
	then.AssertThat(t, filtered[0].Name, is.EqualTo("file1.fcr"))
	then.AssertThat(t, filtered[1].Name, is.EqualTo("file2.csv"))

	sorted := sortScenarios(filtered)
	then.AssertThat(t, len(sorted), is.EqualTo(2))
	then.AssertThat(t, sorted[0].Name, is.EqualTo("file1.fcr"))
	then.AssertThat(t, sorted[1].Name, is.EqualTo("file2.csv"))
}
