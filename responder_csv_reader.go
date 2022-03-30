package rosco

import (
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"os"
)

type ScenarioCSVReader struct {
	filepath string
	file     *os.File
	data     []*RawData
}

func NewScenarioCSVReader(filepath string) *ScenarioCSVReader {
	r := &ScenarioCSVReader{}
	r.filepath = filepath

	return r
}

func (r *ScenarioCSVReader) Load() ([]*RawData, error) {
	var err error

	if err = r.openFile(); err == nil {
		if err = gocsv.Unmarshal(r.file, &r.data); err != nil {
			log.Errorf("error parsing csv file %s (%s)", r.filepath, err)
		} else {
			log.Infof("successfully parsed %s, %d records read", r.filepath, len(r.data))
		}
	}

	return r.data, err
}

func (r *ScenarioCSVReader) openFile() error {
	var err error

	if r.file, err = os.OpenFile(r.filepath, os.O_RDWR|os.O_CREATE, os.ModePerm); err != nil {
		log.Errorf("error opening csv file %s (%s)", r.filepath, err)
	}

	return err
}
