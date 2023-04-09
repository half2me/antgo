package usb

import (
	"github.com/google/gousb"
)

type Driver struct {
	ctx  *gousb.Context
	dev  *gousb.Device
	in   *gousb.InEndpoint
	out  *gousb.OutEndpoint
	done *func()
}

func GetDevice(vid, pid gousb.ID) (Driver, error) {
	ctx := gousb.NewContext()
	dev, err := ctx.OpenDeviceWithVIDPID(vid, pid)
	if err != nil {
		return Driver{}, err
	}

	intf, done, err := dev.DefaultInterface()
	if err != nil {
		return Driver{}, err
	}

	d := Driver{
		ctx:  ctx,
		done: &done,
	}

	d.out, err = intf.OutEndpoint(1)
	if err != nil {
		return Driver{}, err
	}

	d.in, err = intf.InEndpoint(1)
	if err != nil {
		return Driver{}, err
	}

	return d, nil
}

func AutoDetectDevice() (Driver, error) {
	return GetDevice(0x0fcf, 0x1009)
}

func (d Driver) Close() {
	if d.done != nil {
		(*d.done)()
	}
	if d.dev != nil {
		_ = d.dev.Close()
	}
	d.ctx.Close()
}

func (d Driver) Write(p []byte) (n int, err error) {
	return d.out.Write(p)
}

func (d Driver) Read(buf []byte) (n int, err error) {
	n, err = d.in.Read(buf)
	return
}

func (d Driver) BufferSize() int {
	return 4096
}
