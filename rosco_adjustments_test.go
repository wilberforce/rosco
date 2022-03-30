package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_adjustments_applyAdjustments(t *testing.T) {
	var err error
	var connected bool
	var value int

	virtualPort := getVirtualPort()

	r := NewECUReaderInstance()
	connected, err = r.ConnectAndInitialiseECU(virtualPort)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	// adjust short term fuel trim
	value, err = r.applyAdjustment(MEMSSTFTIncrement, MEMSSTFTDecrement, 138, 1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.EqualTo(139))

	value, err = r.applyAdjustment(MEMSSTFTIncrement, MEMSSTFTDecrement, 138, -1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.EqualTo(137))

	value, err = r.applyAdjustment(MEMSSTFTIncrement, MEMSSTFTDecrement, 138, 0)
	then.AssertThat(t, err, is.Not(is.Nil()))
	then.AssertThat(t, value, is.EqualTo(138))

	_ = r.Disconnect()
}

func Test_adjustments_AllAdjusters(t *testing.T) {
	var err error
	var connected bool
	var value int

	virtualPort := getVirtualPort()

	r := NewECUReaderInstance()
	connected, err = r.ConnectAndInitialiseECU(virtualPort)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, connected, is.True())

	value, err = r.AdjustShortTermFuelTrim(1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.GreaterThan(0))

	value, err = r.AdjustLongTermFuelTrim(1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.GreaterThan(0))

	value, err = r.AdjustIdleDecay(1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.GreaterThan(0))

	value, err = r.AdjustIdleSpeed(1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.GreaterThan(0))

	value, err = r.AdjustIgnitionAdvanceOffset(1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.GreaterThan(0))

	value, err = r.AdjustIACPosition(1)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, value, is.GreaterThan(0))

	_ = r.Disconnect()
}
