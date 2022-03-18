package rosco

import "github.com/tarm/serial"

type MEMSReader struct {
	port *serial.Port
}

func NewMEMSReader() *MEMSReader {
	m := &MEMSReader{}
	return m
}

func (r *MEMSReader) Open(connection string) error {
	var err error
	return err
}

func (r *MEMSReader) Read(b []byte) (int, error) {
	var err error
	var n int

	return n, err
}

func (r *MEMSReader) Write(b []byte) (int, error) {
	var err error
	var n int

	return n, err
}

func (r *MEMSReader) Flush() {
}

func (r *MEMSReader) Close() error {
	var err error
	return err
}
