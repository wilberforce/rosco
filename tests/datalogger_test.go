package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestOpenLogfile(t *testing.T) {
	log := rosco.NewMemsDataLogger(".", "TEST")

	then.AssertThat(t, log.IsOpen, is.True())
}

func TestWriteLogfile(t *testing.T) {
	log := rosco.NewMemsDataLogger(".", "TEST")
	then.AssertThat(t, log.IsOpen, is.True())

	memsdata := rosco.MemsData{}
	log.WriteMemsDataToFile(memsdata)

	log.Close()
	then.AssertThat(t, log.IsOpen, is.False())
}

func TestCloseExceptionLogfile(t *testing.T) {
	log := rosco.MemsDataLogger{}
	log.Close()
	then.AssertThat(t, log.IsOpen, is.False())
}
