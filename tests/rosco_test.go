package tests

import (
	"github.com/andrewdjackson/rosco"
	"testing"
)

func TestConnectAndInitialise(t *testing.T) {
	port := "/dev/null"

	mems := rosco.NewMemsConnection()
	mems.ConnectAndInitialiseECU(port)

	if mems == nil {
		t.Errorf("failed")
	}

	if mems.Connected == true {
		t.Errorf("ECU Connected should be False")
	}

	if mems.Initialised == true {
		t.Errorf("ECU Initialised should be False")
	}
}
