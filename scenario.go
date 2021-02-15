package rosco

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// getScenarioPath returns the path to the scenario files
func getScenarioPath(file string) string {
	var filename string

	homeFolder, _ := homedir.Dir()
	appFolder := fmt.Sprintf("%s/memsfcr", homeFolder)

	if file == "" {
		filename = fmt.Sprintf("%s/logs", appFolder)
	} else {
		filename = fmt.Sprintf("%s/logs/%s", appFolder, file)
	}

	return filepath.FromSlash(filename)
}

// GetScenarios reads the directory and returns
// a list of scenario entries sorted by filename.
func GetScenarios() ([]ScenarioDescription, error) {
	logFolder := getScenarioPath("")

	var scenarios []ScenarioDescription

	fileInfo, err := ioutil.ReadDir(logFolder)

	if err == nil {
		for _, file := range fileInfo {
			if strings.HasSuffix(file.Name(), ".csv") {
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

// GetScenario returns the data for the given scenario
func GetScenario(id string) ScenarioDescription {
	file := getScenarioPath(id)
	r := NewResponder()
	_ = r.LoadScenario(file)

	scenario := ScenarioDescription{}
	scenario.Count = r.Playbook.Count
	scenario.Position = r.Playbook.Position
	scenario.Name = id

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
