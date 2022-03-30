package rosco

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// getScenarioPath returns the path to the scenario files
func getScenarioPath(file string) string {
	if file == "" {
		return GetLogFolder()
	}
	return GetFullScenarioFilePath(file)
}

// GetScenarios reads the directory and returns
// a list of scenario entries sorted by filename.
func GetScenarios() ([]ScenarioDescription, error) {
	logFolder := getScenarioPath("")

	var scenarios []ScenarioDescription

	fileInfo, err := ioutil.ReadDir(logFolder)

	if err == nil {
		for _, file := range fileInfo {
			if isValidLogFile(file) {
				scenario := ScenarioDescription{}
				scenario.Date = file.ModTime()
				scenario.Name = file.Name()
				scenario.Count = lineCounter(fmt.Sprintf("%s/%s", logFolder, scenario.Name))
				scenario.Status = "Ready"
				scenarios = append(scenarios, scenario)
			}
		}

		scenarios = sortScenarios(scenarios)
	}

	log.Infof("sorted scenarios (%+v)", scenarios)

	return scenarios, err
}

func isValidLogFile(file os.FileInfo) bool {
	return strings.HasSuffix(strings.ToLower(file.Name()), ".csv") || strings.HasSuffix(strings.ToLower(file.Name()), ".fcr")
}

// GetScenario returns the data for the given scenario
func GetScenario(id string) ScenarioDescription {
	file := getScenarioPath(id)
	r := NewResponder()
	err := r.LoadScenario(file)

	scenario := ScenarioDescription{}

	if err == nil {
		scenario.Count = r.Playbook.Count
		scenario.Position = r.Playbook.Position
		scenario.Name = id
		scenario.Details.First, _ = r.GetFirst()
		scenario.Details.Current, _ = r.GetCurrent()
		scenario.Details.Last, _ = r.GetLast()
	}

	return scenario
}

type timeSlice []ScenarioDescription

func (s timeSlice) Less(i, j int) bool { return s[i].Date.Before(s[j].Date) }
func (s timeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s timeSlice) Len() int           { return len(s) }

func sortScenarios(scenarios []ScenarioDescription) []ScenarioDescription {
	//Sort the map by date
	sortedScenarios := make(timeSlice, 0, len(scenarios))
	for _, scenario := range scenarios {
		sortedScenarios = append(sortedScenarios, scenario)
	}

	sort.Sort(sort.Reverse(sortedScenarios))
	return sortedScenarios
}

func lineCounter(filepath string) int {
	r, _ := os.Open(filepath)

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count

		case err != nil:
			return count
		}
	}
}
