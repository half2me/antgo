package driver

import (
	"github.com/kylelemons/gousb/usb"
	"log"
	"errors"
	"github.com/half2me/antgo/message"
)

type UsbDevice struct {
	vid, pid int
	context  *usb.Context
	device   *usb.Device
	in, out  usb.Endpoint
}

func (dev *UsbDevice) Open() (e error) {
	log.Println("Opening USB device")

	dev.context = usb.NewContext()

	dev.device, e = dev.context.OpenDeviceWithVidPid(dev.vid, dev.pid)

	if e != nil {
		return
	}

	if dev.device == nil {
		e = errors.New("USB Device not found!")
		return
	}

	dev.in, e = dev.device.OpenEndpoint(1, 0, 0, uint8(1)|uint8(usb.ENDPOINT_DIR_IN))
	if e != nil {
		return
	}

	dev.out, e = dev.device.OpenEndpoint(1, 0, 0, uint8(1)|uint8(usb.ENDPOINT_DIR_OUT))
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

func GetUsbDevice(vid, pid int) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
	}
}
