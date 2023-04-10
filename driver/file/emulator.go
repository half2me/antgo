package file

import (
	"github.com/half2me/antgo/ant"
	"io"
	"os"
)

type Emulator struct {
	pr   *io.PipeReader
	pw   *io.PipeWriter
	file *os.File
}

func NewEmulator(path string) (e Emulator, err error) {
	//e.buf = bytes.NewBuffer([]byte{})
	e.pr, e.pw = io.Pipe()
	e.file, err = os.Open(path)
	return
}

func (e Emulator) Close() {
	_ = e.pw.Close()
	_ = e.file.Close()
}

func (e Emulator) Read(b []byte) (int, error) {
	return e.pr.Read(b)
}

func (e Emulator) Write(b []byte) (int, error) {
	msg := ant.AntPacket(b)
	go func() {
		switch c := msg.Class(); {
		case c == ant.MESSAGE_SYSTEM_RESET:
			_, _ = e.pw.Write(ant.MakeAntPacket(ant.MESSAGE_STARTUP, []byte{0x00}))
		case c == ant.MESSAGE_NETWORK_KEY ||
			c == ant.MESSAGE_CHANNEL_ASSIGN ||
			c == ant.MESSAGE_CHANNEL_ID ||
			c == ant.MESSAGE_CHANNEL_FREQUENCY ||
			c == ant.MESSAGE_ENABLE_EXT_RX_MESSAGES:
			// we can fine-tune this in the future, for now it's just an empty response event
			_, _ = e.pw.Write(ant.MakeAntPacket(ant.MESSAGE_CHANNEL_EVENT, []byte{0x00}))
		case c == ant.OPEN_RX_SCAN_MODE:
			_, _ = e.pw.Write(ant.MakeAntPacket(ant.MESSAGE_CHANNEL_EVENT, []byte{0x00}))

			go func() {
				_, _ = io.Copy(e.pw, e.file)
			}()
		default:
			// don't do anything
		}
	}()

	return len(b), nil
}

func (e Emulator) BufferSize() int {
	return 4096
}
