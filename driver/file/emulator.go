package file

import (
	"bytes"
	"github.com/half2me/antgo/ant"
)

type Emulator struct {
	buf *bytes.Buffer
}

func NewEmulator() Emulator {
	return Emulator{buf: bytes.NewBuffer([]byte{})}
}

func (e Emulator) Close() {
	return
}

func (e Emulator) Read(b []byte) (int, error) {
	return e.Read(b)
}

func (e Emulator) Write(b []byte) (int, error) {
	msg := ant.AntPacket(b)
	if msg.Class() == ant.MESSAGE_TYPE_BROADCAST {
		// TODO: something
	}
	//e.buf.Write()

	return len(b), nil
}
