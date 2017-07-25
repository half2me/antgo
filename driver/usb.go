package driver

import (
	"github.com/google/gousb"
	"log"
	"errors"
	"github.com/half2me/antgo/message"
)

type UsbDevice struct {
	vid, pid 	gousb.ID
	context  	*gousb.Context
	device   	*gousb.Device
	closeIface	func()
	intf		*gousb.Interface
	in			*gousb.InEndpoint
	out			*gousb.OutEndpoint
}

func (dev *UsbDevice) Open() (e error) {
	log.Println("Opening USB device")

	dev.context = gousb.NewContext()

	dev.device, e = dev.context.OpenDeviceWithVIDPID(dev.vid, dev.pid)

	if e != nil {
		return
	}

	if dev.device == nil {
		e = errors.New("USB Device not found!")
		return
	}

	// Claim default interface
	dev.intf, dev.closeIface, e = dev.device.DefaultInterface()
	if e != nil {
		return
	}

	// Open an OUT endpoint.
	dev.out, e = dev.intf.OutEndpoint(1)
	if e != nil {
		return
	}

	// Open an IN endpoint.
	dev.in, e = dev.intf.InEndpoint(1)
	if e != nil {
		return
	}

	log.Println("USB Device opened")

	return
}

func (dev *UsbDevice) Close() {
	log.Println("Closing USB device")

	dev.out.Write(message.CloseChannelMessage(0))
	dev.out.Write(message.SystemResetMessage())

	if dev.closeIface != nil {
		dev.closeIface()
	}

	if dev.device != nil {
		dev.device.Close()
	}

	if dev.context != nil {
		dev.context.Close()
	}
	log.Println("USB Device closed")
}

func (dev *UsbDevice) Read(b []byte) (int, error) {
	return dev.in.Read(b)
}

func (dev *UsbDevice) Write(b []byte) (int, error) {
	return dev.out.Write(b)
}

func (dev *UsbDevice) BufferSize() int {
	return 64 // replace with maxBufferSize query in google's gousb
}

func GetUsbDevice(vid, pid gousb.ID) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
	}
}
