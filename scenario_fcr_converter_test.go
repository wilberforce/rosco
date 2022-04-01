package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_scenario_NewScenarioFile(t *testing.T) {
	s := NewScenarioFile("testdata/nofaults.csv")
	then.AssertThat(t, s.filePath, is.EqualTo("testdata/nofaults.csv"))
}

func Test_scenario_ConvertLogToScenario(t *testing.T) {
	// create a new scenario file
	s := NewScenarioFile("testdata/nofaults.fcr")
	then.AssertThat(t, s.filePath, is.EqualTo("testdata/nofaults.fcr"))

	// convert the log (csv)
	err := s.ConvertLogToScenario("testdata/nofaults.csv")
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Count, is.EqualTo(338))
}

func Test_scenario_ReadScenarioFile(t *testing.T) {
	s := NewScenarioFile("testdata/nofaults.fcr")
	then.AssertThat(t, s.filePath, is.EqualTo("testdata/nofaults.fcr"))

	err := s.Read()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Name, is.EqualTo("testdata/nofaults.fcr"))
}

func Test_scenario_ConvertAndSaveScenarioFile(t *testing.T) {
	// create a new scenario file
	s := NewScenarioFile("testdata/nofaults.fcr")
	then.AssertThat(t, s.filePath, is.EqualTo("testdata/nofaults.fcr"))

	// convert the log (csv)
	err := s.ConvertLogToScenario("testdata/nofaults.csv")
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Count, is.EqualTo(338))

	err = s.Write()
	then.AssertThat(t, err, is.Nil())
}
