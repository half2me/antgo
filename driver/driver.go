package driver

type Driver interface {
	Close()
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	BufferSize() int
}
