package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/mitchellh/go-homedir"
	"path/filepath"
	"testing"
)

func getVirtualPort() string {
	homefolder, _ := homedir.Dir()
	return filepath.ToSlash(homefolder + "/ttyecu")
}

func Test_mems_Connect(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewMEMSReader(virtualPort)
	connected, err := r.Connect()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.connected, is.True())
	then.AssertThat(t, connected, is.True())

	_ = r.Disconnect()

	r = NewMEMSReader(invalidPort)
	connected, err = r.Connect()

	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, r.connected, is.False())
	then.AssertThat(t, connected, is.False())

	_ = r.Disconnect()

	r = NewMEMSReader(loopbackPort)
	connected, err = r.Connect()

	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, r.connected, is.False())
	then.AssertThat(t, connected, is.False())
}

func Test_mems_connectToSerialPort(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewMEMSReader(virtualPort)
	err := r.connectToSerialPort(virtualPort)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.serialPort, is.Not(is.Nil()))

	r = NewMEMSReader(virtualPort)
	err = r.connectToSerialPort(invalidPort)

	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, r.serialPort, is.Nil())
}

func Test_mems_Disconnect(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewMEMSReader(virtualPort)
	err := r.Disconnect()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.connected, is.False())
}

func Test_mems_SendAndReceive(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewMEMSReader(virtualPort)

	connected, err := r.Connect()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	// expect echo of command
	response, err := r.SendAndReceive([]byte{0x0A})
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, response, is.EqualTo([]byte{0x0A}))

	// expect id string response
	response, err = r.SendAndReceive([]byte{0xD0})
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, response, is.EqualTo([]byte{0xD0, 0x99, 0x00, 0x03, 0x03}))
}

func Test_mems_commandMatchesResponse(t *testing.T) {
	virtualPort := getVirtualPort()
	r := NewMEMSReader(virtualPort)
	err := r.commandMatchesResponse([]byte{0xca}, []byte{0xca})
	then.AssertThat(t, err, is.Nil())

	err = r.commandMatchesResponse([]byte{0xca}, []byte{0xca, 0x00})
	then.AssertThat(t, err, is.Nil())

	err = r.commandMatchesResponse([]byte{0xca}, []byte{0xd1, 0x00})
	then.AssertThat(t, err, is.Not(is.Nil()))
}
