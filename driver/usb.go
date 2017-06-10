package driver

import (
	"github.com/kylelemons/gousb/usb"
	"log"
	"errors"
	"github.com/half2me/antgo/message"
	"github.com/half2me/antgo/constants"
)

type UsbDevice struct {
	vid, pid int
	context  *usb.Context
	device   *usb.Device
	in, out  usb.Endpoint
	Read chan message.AntPacket
	Write chan message.AntPacket
	decode chan byte
	stopLoop chan int
}

func (dev *UsbDevice) Open() (e error) {
	log.Println("Opening device")
	dev.Read = make(chan message.AntPacket)
	dev.Write = make(chan message.AntPacket)
	dev.decode = make(chan byte)

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
	go dev.decodeLoop()

	log.Println("Device opened")

	return
}

func (dev *UsbDevice) Close() {
	log.Println("Closing device")
	dev.stopLoop <- 1

	if dev.device != nil {
		dev.device.Close()
	}

	if dev.context != nil {
		dev.context.Close()
	}
	log.Println("Device closed")
}

func (dev *UsbDevice) StartRxScanMode() {
	dev.Write <- message.SystemResetMessage()
	dev.Write <- message.SetNetworkKeyMessage(0, []byte(constants.ANTPLUS_NETWORK_KEY))
	dev.Write <- message.AssignChannelMessage(0, constants.CHANNEL_TYPE_ONEWAY_RECEIVE)
	dev.Write <- message.SetChannelIdMessage(0)
	dev.Write <- message.SetChannelRfFrequencyMessage(0, 2457)
	dev.Write <- message.EnableExtendedMessagesMessage(true)
	dev.Write <- message.LibConfigMessage(true, true, true)
	dev.Write <- message.OpenRxScanModeMessage()
}

func (dev *UsbDevice) loop() {
	log.Println("Loop started")
	defer close(dev.decode)
	defer log.Println("Stopping loop")

	for {
		select {
		case <- dev.stopLoop:
			return
		case d := <- dev.Write:
			dev.out.Write(d)
		default:
			// Read from device
			buf := make([]byte, 64)
			i, err := dev.in.Read(buf)

			if err == nil {
				for _, v := range buf[:i] {
					dev.decode <- v
				}
			}
		}
	}
}

func (dev *UsbDevice) decodeLoop() {
	defer close(dev.Read)

	for {
		// Wait for TX Sync
		sync, ok := <- dev.decode

		if !ok {
			return
		}

		if sync != constants.MESSAGE_TX_SYNC {
			continue
		}

		// Get content length (+1byte type + 1byte checksum)
		length, ok := <- dev.decode

		if !ok {
			return
		}

		buf := make([]byte, length+2)

		for i := 0; i < int(length+2); i++ {
			buf[i], ok = <- dev.decode

			if !ok {
				return
			}
		}

		// Check message integrity
		msg := message.AntPacket(append(message.AntPacket{constants.MESSAGE_TX_SYNC, length}, buf...))

		if msg.Valid() {
			dev.Read <- msg
		} else {
			// Optionally log bad msg
			log.Println("ANT+ msg with bad checksum")
		}
	}
}

func GetUsbDevice(vid, pid int) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
		stopLoop: make(chan int),
	}
}
