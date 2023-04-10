package file

import (
	"os"
)

type AntCaptureFile struct {
	file *os.File
}

func Open(path string) (f AntCaptureFile, err error) {
	f.file, err = os.Open(path)
	return
}

func (f AntCaptureFile) Close() {
	_ = f.file.Close()
}

func (f AntCaptureFile) Read(b []byte) (int, error) {
	return f.file.Read(b)
}

func (f AntCaptureFile) Write(b []byte) (int, error) {
	// Writes are ignored
	return len(b), nil
}
