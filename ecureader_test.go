package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"reflect"
	"testing"
)

const (
	invalidPort  = ""
	loopbackPort = "loopback"
	scenarioPort = "scenario.csv"
)

func Test_ecureader_NewECUReader(t *testing.T) {
	r := NewECUReader("loopback")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&LoopbackReader{})))

	r = NewECUReader("/loopback")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&LoopbackReader{})))

	// test the FCR and CSV file extensions
	r = NewECUReader("filename.csv")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&ScenarioReader{})))

	r = NewECUReader("filename.fcr")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&ScenarioReader{})))

	// ensure only the extension determines the reader is a file reader
	r = NewECUReader("filenamefcr.file")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&MEMSReader{})))

	r = NewECUReader("filenamecsv.file")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&MEMSReader{})))

	// MEMSReader for serial ports
	r = NewECUReader("COM5")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&MEMSReader{})))

	r = NewECUReader("/dev/tty.Serial")
	then.AssertThat(t, reflect.TypeOf(r), is.EqualTo(reflect.TypeOf(&MEMSReader{})))
}

func Test_ecureader_createResponseMap(t *testing.T) {
	m := createResponseMap()
	then.AssertThat(t, len(m), is.GreaterThan(0))
	then.AssertThat(t, m["79"], is.EqualTo([]byte{0x79, 0x8b}))
	// unmapped command
	then.AssertThat(t, m["01"], is.EqualTo([]byte{0x01, 0x00}))
}

func Test_ecureader_getResponseSize(t *testing.T) {
	s, err := getResponseSize([]byte{0x79})
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, s, is.EqualTo(2))

	// unmapped command, expect a default response sze of 2 bytes
	s, err = getResponseSize([]byte{0x20})
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, s, is.EqualTo(2))
}
