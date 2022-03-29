package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"reflect"
	"testing"
)

func Test_rosco_ConnectAndInitialiseECU(t *testing.T) {
	virtualPort := getVirtualPort()
	rosco_ConnectAndInitialiseECU(t, virtualPort)

	rosco_ConnectAndInitialiseECU(t, loopbackPort)

	rosco_ConnectAndInitialiseECU(t, scenarioPort)

	r := NewECUReaderInstance()
	connected, err := r.ConnectAndInitialiseECU(invalidPort)

	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, r.Status.Connected, is.False())
	then.AssertThat(t, connected, is.False())
}

func rosco_ConnectAndInitialiseECU(t *testing.T, port string) {
	r := NewECUReaderInstance()
	connected, err := r.ConnectAndInitialiseECU(port)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.Status.Connected, is.True())
	then.AssertThat(t, connected, is.True())

	if reflect.TypeOf(r.ecuReader) == reflect.TypeOf(&ScenarioReader{}) {
		then.AssertThat(t, r.ecuReader.(*ScenarioReader).Responder, is.Not(is.Nil()))
	}

	_ = r.Disconnect()
}

func Test_rosco_Disconnect(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewECUReaderInstance()
	connected, err := r.ConnectAndInitialiseECU(virtualPort)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.Status.Connected, is.True())
	then.AssertThat(t, connected, is.True())

	err = r.Disconnect()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.Status.Connected, is.False())
}

func Test_rosco_connectToeECU(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewECUReaderInstance()
	r.ecuReader = NewECUReader(virtualPort)
	connected, err := r.connectToECU()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())
}

func Test_rosco_ResetDiagnostics(t *testing.T) {
	r := NewECUReaderInstance()
	r.ResetDiagnostics()
	then.AssertThat(t, r.Diagnostics.datasetLength, is.EqualTo(20))
}

func Test_rosco_GetDataframes(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewECUReaderInstance()
	r.ecuReader = NewECUReader(virtualPort)
	connected, err := r.connectToECU()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	df := r.GetDataframes()
	then.AssertThat(t, df.CoolantTemp, is.GreaterThan(1))
}
