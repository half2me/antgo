package driver

import (
	"github.com/google/gousb"
	"log"
	"errors"
	"github.com/half2me/antgo/message"
	"github.com/half2me/antgo/constants"
)

type UsbDevice struct {
	vid, pid 	gousb.ID
	context  	*gousb.Context
	device   	*gousb.Device
	closeIface	func()
	intf		*gousb.Interface
	in			*gousb.InEndpoint
	out			*gousb.OutEndpoint
	Read 		chan message.AntPacket
	Write 		chan message.AntPacket
	decode 		chan byte
	stopLoop 	chan int
}

func (dev *UsbDevice) Open() (e error) {
	log.Println("Opening device")
	dev.Read = make(chan message.AntPacket)
	dev.Write = make(chan message.AntPacket)
	dev.decode = make(chan byte)

	dev.context = gousb.NewContext()

	dev.device, e = dev.context.OpenDeviceWithVIDPID(dev.vid, dev.pid)

	if dev.device == nil {
		e = errors.New("Device not found!")
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
	dev.in, e = dev.intf.InEndpoint(3)
	if e != nil {
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

	if dev.closeIface != nil {
		dev.closeIface()
	}

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
	//dev.Write <- message.LibConfigMessage(true, true, true)
	dev.Write <- message.OpenRxScanModeMessage()
}

func (dev *UsbDevice) loop() {
	log.Println("Loop started")
	defer close(dev.decode)
	defer log.Println("Stopping loop")

	for {
		select {
		case <- dev.stopLoop:
			dev.out.Write(message.CloseChannelMessage(0))
			dev.out.Write(message.SystemResetMessage())
			return
		case d := <- dev.Write:
			dev.out.Write(d)
		default:
			// Read from device
			buf := make([]byte, dev.in.Desc.MaxPacketSize)
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

func GetUsbDevice(vid, pid gousb.ID) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
		stopLoop: make(chan int),
	}
}
