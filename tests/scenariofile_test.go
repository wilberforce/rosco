package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestSaveScenarioFile(t *testing.T) {
	filepath := "test.scn"

	s := rosco.NewScenarioFile(filepath)
	s.Name = "Test"
	s.Summary = "test"
	err := s.Write()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Summary, is.EqualTo("test"))
}

func TestLoadScenarioFile(t *testing.T) {
	filepath := "test.scn"
	s := rosco.NewScenarioFile(filepath)
	err := s.Read()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Summary, is.EqualTo("test"))
}

func TestConvertLogToScenarioFile(t *testing.T) {
	s := rosco.NewScenarioFile("vacuum-fault.scn")
	err := s.ConvertLogToScenario("vacuum-fault.csv")

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Count, is.GreaterThan(1))
}

func TestConvertAndSaveScenarioFile(t *testing.T) {
	s := rosco.NewScenarioFile("vacuum-fault.scn")
	err := s.ConvertLogToScenario("vacuum-fault.csv")

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s.Count, is.GreaterThan(1))

	err = s.Write()
	then.AssertThat(t, err, is.Nil())
}
