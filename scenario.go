package rosco

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"time"
)

// GetScenarios reads the directory and returns
// a list of scenario entries sorted by filename.
func GetScenarios(folder string) ([]ScenarioDescription, error) {
	logFolder := GetFullScenarioFilePath(folder)

	var scenarios []ScenarioDescription

	fileInfo, err := ioutil.ReadDir(logFolder)

	if err == nil {
		for _, file := range fileInfo {
			if isValidLogFile(file) {
				filename := fmt.Sprintf("%s%s", logFolder, file.Name())
				if scenario, err := getScenarioInfo(filename); err == nil {
					scenarios = append(scenarios, scenario)
				}
			}
		}

		scenarios = sortScenarios(scenarios)
	}

	log.Infof("sorted scenarios (%+v)", scenarios)

	return scenarios, err
}

// GetScenario returns the data for the given scenario
func GetScenario(id string) ScenarioDescription {
	file := GetFullScenarioFilePath(id)
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

func isValidLogFile(file os.FileInfo) bool {
	filename := file.Name()
	return isCSVFile(filename) || isFCRFile(filename)
}

func getScenarioInfo(filepath string) (ScenarioDescription, error) {
	var err error
	var fileReader ResponderFileReader
	var info ResponderFileInfo
	var description ScenarioDescription

	if fileReader, err = NewResponderFileReader(filepath); err == nil {
		if info, err = fileReader.Load(); err == nil {
			description = ScenarioDescription{
				Name:     info.Description.Name,
				Count:    info.Description.Count,
				Duration: info.Description.Duration,
				Position: 0,
				Date:     info.Description.Date,
				Details:  ScenarioDetails{},
				Summary:  info.Description.Summary,
				FileType: info.Description.FileType,
			}
		}
	}

	return description, err
}

func getScenarioDuration(start string, end string) (string, error) {
	var err error

	if startTime, err := ConvertTimeFieldToDate(start); err == nil {
		if endTime, err := ConvertTimeFieldToDate(end); err == nil {
			dur := endTime.Sub(startTime)
			return humanizeDuration(dur), err
		}
	}

	return "", err
}

// humanizeDuration humanizes time.Duration output to a meaningful value,
// golang's default ``time.Duration`` output is badly formatted and unreadable.
func humanizeDuration(duration time.Duration) string {
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%ds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%dm %ds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%dh %dm %ds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%d days %dh %dm %ds",
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
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
