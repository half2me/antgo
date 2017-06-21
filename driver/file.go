package driver

import (
	"os"
)

type AntCaptureFile struct {
	readpath string
	writepath string
	readfile *os.File
	writefile *os.File
	bufSize uint
	buf []byte
}

func (f *AntCaptureFile) Open() (e error) {
	if len(f.readpath) > 0 {
		f.readfile, e = os.Open(f.readpath)
		if e != nil {
			return
		}
	}

	if len(f.writepath) > 0 {
		f.writefile, e = os.Open(f.writepath)
	}
	return
}

func (f *AntCaptureFile) Close() {
	if f.readfile != nil {
		f.readfile.Close()
	}

	if f.writefile != nil {
		f.writefile.Close()
	}
}

func (f *AntCaptureFile) Read(b []byte) (int, error) {
	return f.Read(f.buf)
}

func (f *AntCaptureFile) Write(b []byte) (int, error) {
	if f.writefile != nil {
		return f.writefile.Write(b)
	}

	// If there is not writeFile, we just ignore output
	return len(b), nil
}

func (f *AntCaptureFile) BufferSize() uint {
	return f.bufSize
}

func GetAntCaptureFile(readpath, writepath string) *AntCaptureFile {
	return &AntCaptureFile{
		readpath: readpath,
		writepath: writepath,
		bufSize: 512,
		buf: make([]byte, 512),
	}
}
