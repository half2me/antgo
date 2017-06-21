package driver

import (
	"github.com/half2me/antgo/message"
	"log"
)

type AntDriver interface {
	Open() error
	Close()
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	BufferSize() uint
}

type AntDevice struct {
	Driver AntDriver
	Read chan message.AntPacket
	Write chan message.AntPacket
	stopper chan byte
	decoder chan byte
	done chan byte
}

func (dev *AntDevice) Start() (e error) {
	log.Println("Starting Device")
	e = dev.Driver.Open()
	go dev.loop()
	go dev.decodeLoop()
	return
}

func (dev *AntDevice) Stop() {
	dev.stopper <- 1

	// Wait for loops to finish
	<- dev.done
	<- dev.done
}

func (dev *AntDevice) loop() {
	defer func() {dev.done <- 1}()
	defer dev.Driver.Close()
	defer close(dev.decoder)
	log.Println("Loop Started")

	for {
		select {
		case <- dev.stopper:
			log.Println("Loop stopped!")
			return
		case d := <- dev.Write:
			dev.Driver.Write(d)
		default:
			// Read from device
			buf := make([]byte, dev.Driver.BufferSize())
			i, err := dev.Driver.Read(buf)

			if err == nil {
				for _, v := range buf[:i] {
					dev.decoder <- v
				}
			}
		}
	}
}

func (dev *AntDevice) decodeLoop() {
	defer func() {dev.done <- 1}()
	defer close(dev.Read)

	for {
		// Wait for TX Sync
		sync, ok := <- dev.decoder

		if !ok {
			return
		}

		if sync != message.MESSAGE_TX_SYNC {
			continue
		}

		// Get content length (+1byte type + 1byte checksum)
		length, ok := <- dev.decoder

		if !ok {
			return
		}

		buf := make([]byte, length+2)

		for i := 0; i < int(length+2); i++ {
			buf[i], ok = <- dev.decoder

			if !ok {
				return
			}
		}

		// Check message integrity
		msg := message.AntPacket(append(message.AntPacket{message.MESSAGE_TX_SYNC, length}, buf...))

		if msg.Valid() {
			dev.Read <- msg
		}
	}
}

func NewDevice(driver AntDriver) *AntDevice {
	return &AntDevice {
		Driver: driver,
		Read: make(chan message.AntPacket),
		Write: make(chan message.AntPacket),
		stopper: make(chan byte),
		decoder: make(chan byte),
		done: make(chan byte),
	}
}

func (dev *AntDevice) StartRxScanMode() {
	dev.Write <- message.SystemResetMessage()
	dev.Write <- message.SetNetworkKeyMessage(0, []byte(message.ANTPLUS_NETWORK_KEY))
	dev.Write <- message.AssignChannelMessage(0, message.CHANNEL_TYPE_ONEWAY_RECEIVE)
	dev.Write <- message.SetChannelIdMessage(0)
	dev.Write <- message.SetChannelRfFrequencyMessage(0, 2457)
	dev.Write <- message.EnableExtendedMessagesMessage(true)
	//dev.Write <- message.LibConfigMessage(true, true, true)
	dev.Write <- message.OpenRxScanModeMessage()
}
