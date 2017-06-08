package main

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
	dev.Read = make(chan []byte, 20)
	dev.Write = make(chan []byte, 20)

	dev.context = usb.NewContext()
	dev.context.Debug(0)

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

	go dev.loop()

	log.Println("Device opened")

	return
}

func (dev *UsbDevice) Close() {
	log.Println("Closing device")
	close(dev.Read)
	dev.stopLoop <- 1

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
			log.Println("Stopping loop")
			return
		case d := <- dev.Write:
			dev.out.Write(d)
		default:
			// Read from device
			buf := make([]byte, 20)
			dev.in.Read(buf)
		}
	}
}

func GetDevice(vid, pid int) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
		stopLoop: make(chan int),
	}
}
