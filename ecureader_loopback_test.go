package rosco

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"testing"
)

func Test_loopback_Open(t *testing.T) {
	r := NewLoopbackReader()
	err := r.Open("loopback")

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, len(r.responseMap), is.GreaterThan(0))

	err = r.Open("invalid")
	then.AssertThat(t, err, is.Not(is.Nil()))
}

func Test_loopback_Read(t *testing.T) {
	r := NewLoopbackReader()
	n, err := r.Read([]byte{0x0A})

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, n, is.EqualTo(1))
}
