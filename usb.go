package antgo

import (
	"github.com/kylelemons/gousb/usb"
	"log"
	"errors"
)

type Device interface {
	Open(c chan byte) error
	Close()
}

type UsbDevice struct {
	vid, pid int
	context  *usb.Context
	device   *usb.Device
	in, out  usb.Endpoint
	Read chan []byte
	Write chan []byte
	stopLoop chan int
}

func (dev *UsbDevice) Open() (e error) {
	log.Println("Opening device")
	dev.context = usb.NewContext()
	dev.context.Debug(3)

	dev.device, e = dev.context.OpenDeviceWithVidPid(dev.vid, dev.pid)

	if e != nil {
		defer dev.context.Close()
		return
	}
	if dev.device == nil {
		defer dev.context.Close()
		e = errors.New("Device not found!")
		return
	}

	dev.in, e = dev.device.OpenEndpoint(1, 0, 0, uint8(1)|uint8(usb.ENDPOINT_DIR_IN))
	if e != nil {
		defer dev.context.Close()
		defer dev.Close()
		return
	}

	dev.out, e = dev.device.OpenEndpoint(1, 0, 0, uint8(1)|uint8(usb.ENDPOINT_DIR_OUT))
	if e != nil {
		defer dev.context.Close()
		defer dev.Close()
		return
	}

	dev.Read = make(chan []byte)
	dev.Write = make(chan []byte)

	go dev.loop()

	log.Println("Device opened")

	return
}

func (dev *UsbDevice) Close() {
	log.Println("Closing device")
	dev.stopLoop <- 1
	close(dev.Read)

	if dev.device != nil {
		dev.device.Close()
	}

	if dev.context != nil {
		dev.context.Close()
	}
	log.Println("Device closed")
}

func (dev *UsbDevice) loop() {
	log.Println("Loop started")
	for {
		select {
		case <- dev.stopLoop:
			break
		case d := <- dev.Write:
			dev.out.Write(d)
		default:
			// Read from device
			buf := make([]byte, 20)
			_, err := dev.in.Read(buf)

			if ! (err == nil || err == usb.ERROR_TIMEOUT) {
				log.Fatalln("Error reading from endpoint, ", err)
				close(dev.Read)
				break;
			}
		}
	}
	log.Println("Loop stopped")
}

func GetDevice(vid, pid int) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
	}
}
