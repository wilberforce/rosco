package rosco

import (
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type LoopbackReader struct {
	connected bool
}

func NewLoopbackReader() *LoopbackReader {
	log.Infof("created loopback ecu reader")

	// initialise the responseMap
	responseMap = createResponseMap()

	return &LoopbackReader{}
}

func (r *LoopbackReader) Connect() (bool, error) {
	r.connected = true
	return r.connected, nil
}

func (r *LoopbackReader) SendAndReceive(command []byte) ([]byte, error) {
	var err error
	var response []byte

	// convert the command code to a string
	cmd := hex.EncodeToString(command)
	cmd = strings.ToUpper(cmd)

	if response, err = r.getResponse(cmd); err != nil {
		// couldn't find a response in the map for the given command
		// generate a response and clear the error
		response = command
		response = append(response, []byte{0x00}...)
		log.Warnf("manufactured response %X", response)
		err = nil
	}

	return response, err
}

func (r *LoopbackReader) Disconnect() error {
	r.connected = false
	return nil
}

func (r *LoopbackReader) getResponse(command string) ([]byte, error) {
	var err error

	response := responseMap[command]

	if response == nil {
		err = fmt.Errorf("unable to find response for command %s", command)
		log.Warnf("%s", err)
	} else {
		log.Infof("Loopback response %+v", response)
	}

	return response, err
}
