package driver

import (
	"os"
	"io"
	"time"
)

type AntCaptureFile struct {
	path string
	file *os.File
}

func (f *AntCaptureFile) Open() (e error) {
	f.file, e = os.Open(f.path)
	return
}

func (f *AntCaptureFile) Close() {
	f.file.Close()
}

func (f *AntCaptureFile) Read(b []byte) (n int, e error) {
	n, e = f.file.Read(b)
	if e == io.EOF {
		f.file.Seek(0, 0)
		n, e = f.file.Read(b)
	}
	time.Sleep(100 * time.Millisecond) // Artificial delay

	return
}

func (f *AntCaptureFile) Write(b []byte) (int, error) {
	// We ignore output
	return len(b), nil
}

func (f *AntCaptureFile) BufferSize() int {
	return 512
}

func GetAntCaptureFile(path string) *AntCaptureFile {
	return &AntCaptureFile{
		path:path,
	}
}
