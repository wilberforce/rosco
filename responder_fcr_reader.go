package rosco

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type ScenarioFCRReader struct {
	filepath string
	file     *os.File
	data     []*RawData
}

func NewScenarioFCRReader(filepath string) *ScenarioFCRReader {
	r := &ScenarioFCRReader{}
	r.filepath = filepath

	return r
}

func (r *ScenarioFCRReader) Load() ([]*RawData, error) {
	var err error
	var fcrData ScenarioFile

	if err = r.openFile(); err == nil {
		data, _ := ioutil.ReadAll(r.file)

		if err = json.Unmarshal(data, &fcrData); err != nil {
			log.Errorf("error parsing csv file %s (%s)", r.filepath, err)
		} else {
			r.data = fcrData.RawData
			log.Infof("successfully parsed %s, %d records read", r.filepath, len(r.data))
		}
	}

	return r.data, err
}

func (r *ScenarioFCRReader) openFile() error {
	var err error

	if r.file, err = os.OpenFile(r.filepath, os.O_RDWR|os.O_CREATE, os.ModePerm); err != nil {
		log.Errorf("error opening csv file %s (%s)", r.filepath, err)
	}

	return err
}
