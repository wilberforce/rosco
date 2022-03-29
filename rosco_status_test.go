package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_rosco_getECUID(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewECUReaderInstance()
	r.ecuReader = NewECUReader(virtualPort)
	connected, err := r.connectToECU()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	response, err := r.getECUID()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, response, is.EqualTo("99000303"))
}

func Test_rosco_getECUSerial(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewECUReaderInstance()
	r.ecuReader = NewECUReader(virtualPort)
	connected, err := r.connectToECU()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	response, err := r.getECUSerial()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, response, is.EqualTo("ABNMP00399000303"))
}

func Test_rosco_getIACPosition(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewECUReaderInstance()
	r.ecuReader = NewECUReader(virtualPort)
	connected, err := r.connectToECU()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	response, err := r.GetIACPosition()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, response, is.EqualTo(128))
}
