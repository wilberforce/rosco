//go:build live
// +build live

package tests

import (
	"github.com/andrewdjackson/rosco"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func TestLiveConnectAndInitialise(t *testing.T) {
	logfolder := "logs"
	port := "/dev/cu.usbserial-FT94CQQS"

	mems := rosco.NewECUReaderInstance(logfolder)
	mems.ConnectAndInitialiseECU(port)

	then.AssertThat(t, mems.Status.Connected, is.True())
	then.AssertThat(t, mems.Status.Initialised, is.True())

	if mems.Status.Initialised {
		memsdata := mems.GetDataframes()

		then.AssertThat(t, memsdata.BatteryVoltage, is.GreaterThan(10.0))
	}

	if mems.Status.Connected {
		mems.Disconnect()

		then.AssertThat(t, mems.Status.Connected, is.False())
		then.AssertThat(t, mems.Status.Initialised, is.False())
	}
}
