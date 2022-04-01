package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_scenarioReader_Connect(t *testing.T) {
	var err error
	var connected bool

	r := NewScenarioReader("testdata/nofaults.csv")
	connected, err = r.Connect()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.connected, is.True())
	then.AssertThat(t, connected, is.True())
}

func Test_scenarioReader_SendAndReceive(t *testing.T) {
	var err error
	var connected bool
	var response []byte

	r := NewScenarioReader("testdata/nofaults.csv")
	connected, err = r.Connect()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	// test dataframes
	response, err = r.SendAndReceive([]byte{0x7D})
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(response), is.EqualTo(33))
	then.AssertThat(t, response[0:1], is.EqualTo([]byte{0x7d}))

	response, err = r.SendAndReceive([]byte{0x80})
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(response), is.EqualTo(29))
	then.AssertThat(t, response[0:1], is.EqualTo([]byte{0x80}))

	// test general command
	response, err = r.SendAndReceive([]byte{0x75})
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(response), is.GreaterThanOrEqualTo(1))
	then.AssertThat(t, response[0:1], is.EqualTo([]byte{0x75}))
}

func Test_scenarioReader_Disconnect(t *testing.T) {
	var err error
	var connected bool

	r := NewScenarioReader("testdata/nofaults.csv")
	connected, err = r.Connect()

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.connected, is.True())
	then.AssertThat(t, connected, is.True())

	err = r.Disconnect()
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, r.connected, is.False())
}
