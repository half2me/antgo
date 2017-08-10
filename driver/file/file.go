package file

import (
	"os"
	"io"
	"time"
	"github.com/half2me/antgo/message"
	"encoding/gob"
	"fmt"
)

type AntT struct {
	Data message.AntPacket
	Timestamp time.Time
}

func (a *AntT) String() string {
	t := a.Timestamp
	return fmt.Sprintf("[%02d:%02d:%02d:%03d] %s", t.Hour(), t.Minute(), t.Second(), t.Nanosecond() / 1000000, a.Data.String())
}

type AntCaptureFile struct {
	path string
	file *os.File
	enc *gob.Encoder
	dec *gob.Decoder
	offset_t time.Time
	open_t time.Time
}

func (f *AntCaptureFile) Open() (e error) {
	f.file, e = os.Open(f.path)
	f.dec = gob.NewDecoder(f.file)
	f.enc = gob.NewEncoder(f.file)
	f.open_t = time.Now()
	return
}

func (f *AntCaptureFile) Create() (e error) {
	f.file, e = os.Create(f.path)
	f.dec = gob.NewDecoder(f.file)
	f.enc = gob.NewEncoder(f.file)
	f.open_t = time.Now()
	return
}

func (f *AntCaptureFile) Close() {
	f.file.Close()
}

func (f *AntCaptureFile) Read() (a AntT, e error) {
	e = f.dec.Decode(&a)
	return
}

func (f *AntCaptureFile) ReadLoop(out chan message.AntPacket, stop chan struct{}) {
	defer close(out)
	defer f.Close()
	for {
		if ant, e := f.Read(); e != nil {
			if e == io.EOF {
				// Loop at EOF
				f.file.Seek(0, 0)
				f.open_t = time.Now()
				f.offset_t = time.Time{}
				continue
			}
			panic(e.Error())
		} else {
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
}

func (f *AntCaptureFile) Write(packet message.AntPacket) error {
	return f.enc.Encode(AntT {
		Data: packet,
		Timestamp: time.Now(),
	})
}

func (f *AntCaptureFile) BufferSize() int {
	return 16
}

func GetAntCaptureFile(path string) *AntCaptureFile {
	return &AntCaptureFile{
		path:path,
	}
}
