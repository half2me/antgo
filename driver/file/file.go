package file

import (
	"os"
	"io"
	"time"
	"github.com/half2me/antgo/message"
	"encoding/gob"
	"golang.org/x/tools/go/gcimporter15/testdata"
)

type AntT struct {
	Data message.AntPacket
	Timestamp time.Time
}

type AntCaptureFile struct {
	path string
	file *os.File
	dec *gob.Decoder
	offset_t time.Time
	open_t time.Time
}

func (f *AntCaptureFile) Open() (e error) {
	f.file, e = os.Open(f.path)
	f.dec = gob.NewDecoder(f.file)
	f.open_t = time.Now()
	return
}

func (f *AntCaptureFile) Close() {
	f.file.Close()
}

func (f *AntCaptureFile) ReadLoop(out chan message.AntPacket, stop chan struct{}) {
	defer close(out)
	defer f.Close()
	var ant AntT
	for {
		if e := f.dec.Decode(&ant); e != nil {
			if e == io.EOF {
				// Loop at EOF
				f.file.Seek(0, 0)
				f.open_t = time.Now()
				f.offset_t = time.Time{}
				continue
			}
			panic(e.Error())
		}

		if f.offset_t.IsZero() {
			// We initialize the first timestamp as an offset
			f.offset_t = ant.Timestamp
		}

		openTimeDelta := time.Now().Sub(f.open_t)
		packetTimeDelta := ant.Timestamp.Sub(f.offset_t)

		waitFor := packetTimeDelta - openTimeDelta

		if waitFor <= 0 {
			// We can send the message right away
			out <- ant.Data
		} else {
			t := time.NewTimer(waitFor)
			select {
			case <-t.C:
				out <- ant.Data
			case <- stop:
				return
			}
		}
	}
}

func (f *AntCaptureFile) Write(b []byte) (int, error) {
	// We ignore output
	return len(b), nil
}

func (f *AntCaptureFile) BufferSize() int {
	return 16
}

func GetAntCaptureFile(path string) *AntCaptureFile {
	return &AntCaptureFile{
		path:path,
	}
}
