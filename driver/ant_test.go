package driver

import (
	"testing"
	"bytes"
	"github.com/half2me/antgo/message"
	"time"
)

type DummyDriver struct {
	b bytes.Buffer
}

func (*DummyDriver) Open() error {return nil}

func (*DummyDriver) Close() {}

func (d *DummyDriver) Read(b []byte) (int, error) {
	return d.b.Read(b)
}

func (d *DummyDriver) Write(b []byte) (int, error) {
	return d.b.Write(b)
}

func (d *DummyDriver) BufferSize() int {
	return 16
}

func TestAntDevice_Read(t *testing.T) {
	testPack := []byte{0xA4, 0x0E, 0x4E, 0x00, 0x78, 0x5B, 0xD5, 0x07, 0xEC, 0x1F, 0xF4, 0x1A, 0x80, 0x84, 0x69, 0x79, 0x01, 0x1D}
	read := make(chan message.AntPacket)
	write := make(chan message.AntPacket)

	drv := DummyDriver{}
	drv.Write(testPack)
	dev := NewDevice(&drv, read, write)
	if e := dev.Start(); e != nil {
		t.Error(e.Error())
	}
	defer dev.Stop()

	p := <- read
	if bytes.Compare(p, testPack) != 0 {
		t.Error("Packet got corrupted while reading!")
	}
}

func TestAntDevice_ChecksumCheck(t *testing.T) {
	testPack := []byte{0xA4, 0x0E, 0x4E, 0x00, 0x78, 0x5B, 0xD5, 0x07, 0xEC, 0x1F, 0xF4, 0x1A, 0x80, 0x84, 0x69, 0x79, 0x02, 0x1D}
	read := make(chan message.AntPacket)
	write := make(chan message.AntPacket)

	drv := DummyDriver{}
	drv.Write(testPack)
	dev := NewDevice(&drv, read, write)
	if e := dev.Start(); e != nil {
		t.Error(e.Error())
	}
	defer dev.Stop()

	time.Sleep(time.Millisecond * 10)

	select {
		case <- read:
			t.Error("Bad checksum got through")
	default:
	}
}
