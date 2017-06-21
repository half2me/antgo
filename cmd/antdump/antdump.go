package main

import (
"log"
"fmt"
"github.com/half2me/antgo/driver"
"github.com/half2me/antgo/message"
	"flag"
	"os"
	"os/signal"
)

func read(r chan message.AntPacket) {
	for e := range r {
		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			msg := message.AntBroadcastMessage(e)
			switch msg.DeviceType() {
			case message.DEVICE_TYPE_SPEED_AND_CADENCE:
				fmt.Println(message.SpeedAndCadenceMessage(msg))
			case message.DEVICE_TYPE_POWER:
				fmt.Println(message.PowerMessage(msg))
			}
		}
	}
}

func main() {
	drv := flag.String("driver", "usb", "Specify the Driver to use: [usb, serial, file, debug]")
	flag.Bool("raw", true, "Do not attempt to decode ANT+ Broadcast messages")
	pid := flag.Int("pid", 0x1008, "When using the USB driver specify pid of the dongle (i.e.: 0x1008")
	inFile := flag.String("", "in.cap", "File to read ANT+ data from.")
	outFile := flag.String("", "out.cap", "File to write ANT+ data to.")
	flag.Bool("dump", false, "Dump all raw ANT+ data to capture file")
	flag.Parse()

	var device *driver.AntDevice

	switch drv {
	case "usb":
		device = driver.NewDevice(driver.GetUsbDevice(0x0fcf, *pid))
	case "file":
		device = driver.NewDevice(driver.GetAntCaptureFile(*inFile, *outFile))
	}

	err := device.Start()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer device.Stop()

	go read(device.Read)
	device.StartRxScanMode()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
