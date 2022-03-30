package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"os"
	"testing"
)

func Test_datalogger_NewDataLogger(t *testing.T) {
	d := NewMemsDataLogger("testlogs", "TEST")
	then.AssertThat(t, d, is.Not(is.Nil()))
	then.AssertThat(t, d.IsOpen, is.True())
	d.Close()

	f, err := os.Stat(d.Filepath)
	then.AssertThat(t, f.Size(), is.GreaterThan(1360))
	then.AssertThat(t, err, is.Nil())
}

func Test_datalogger_WriteMemsDataToFile(t *testing.T) {
	d := NewMemsDataLogger("testlogs", "TEST")
	then.AssertThat(t, d, is.Not(is.Nil()))
	then.AssertThat(t, d.IsOpen, is.True())

	d.WriteMemsDataToFile(MemsData{})
	d.Close()

	f, err := os.Stat(d.Filepath)
	then.AssertThat(t, f.Size(), is.GreaterThan(1650))
	then.AssertThat(t, err, is.Nil())
}
