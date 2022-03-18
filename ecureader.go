package rosco

type ECUReader interface {
	Open(connection string) (err error)
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Flush()
	Close() (err error)
}
