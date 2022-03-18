package rosco

type ScenarioReader struct {
	scenarioFile string
}

func NewScenarioReader() *ScenarioReader {
	r := &ScenarioReader{}
	return r
}

func (r *ScenarioReader) Open(connection string) error {
	var err error
	return err
}

func (r *ScenarioReader) Read(b []byte) (int, error) {
	var err error
	var n int

	return n, err
}

func (r *ScenarioReader) Write(b []byte) (int, error) {
	var err error
	var n int

	return n, err
}

func (r *ScenarioReader) Flush() {
}

func (r *ScenarioReader) Close() error {
	var err error
	return err
}
