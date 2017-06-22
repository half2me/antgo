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

func read(r chan message.AntPacket, log string) {
	var f *os.File
	var err error
	if len(log) > 0 {
		f, err = os.Create(log)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}

	var prevPower message.PowerMessage = nil
	var prevSnC message.SpeedAndCadenceMessage = nil
	for e := range r {
		if len(log) > 0 {
			f.Write(e)
		}

		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			msg := message.AntBroadcastMessage(e)
			switch msg.DeviceType() {
			case message.DEVICE_TYPE_SPEED_AND_CADENCE:
				cad, stall := message.SpeedAndCadenceMessage(msg).Cadence(prevSnC)
				if !stall {
					fmt.Printf("(%d) %f rpm\n", msg.DeviceNumber(), cad)
				} else {
					fmt.Printf("(%d) - rpm\n", msg.DeviceNumber())
				}

				dist := message.SpeedAndCadenceMessage(msg).Distance(prevSnC, 0.98)
				if dist > 0.001 {
					fmt.Printf("(%d) %f m\n", msg.DeviceNumber(), dist)
				} else {
					fmt.Printf("(%d) - m\n", msg.DeviceNumber())
				}

				speed, stall2 := message.SpeedAndCadenceMessage(msg).Speed(prevSnC, 0.98)
				if !stall2 {
					fmt.Printf("(%d) %f m/s\n", msg.DeviceNumber(), speed)
				} else {
					fmt.Printf("(%d) - m/s\n", msg.DeviceNumber())
				}

				prevSnC = message.SpeedAndCadenceMessage(msg)
			case message.DEVICE_TYPE_POWER:
				pow := message.PowerMessage(msg).AveragePower(prevPower)
				fmt.Printf("(%d) %d W\n", msg.DeviceNumber(), int16(pow))
				prevPower = message.PowerMessage(msg)
			}
		}
	}
}

func main() {
	drv := flag.String("driver", "file", "Specify the Driver to use: [usb, serial, file, debug]")
	flag.Bool("raw", true, "Do not attempt to decode ANT+ Broadcast messages")
	pid := flag.Int("pid", 0x1008, "When using the USB driver specify pid of the dongle (i.e.: 0x1008")
	inFile := flag.String("infile", "capture/123.cap", "File to read ANT+ data from.")
	outFile := flag.String("outfile", "", "File to dump ANT+ data to.")
	flag.Parse()

	var device *driver.AntDevice

	switch *drv {
	case "usb":
		device = driver.NewDevice(driver.GetUsbDevice(0x0fcf, *pid))
	case "file":
		device = driver.NewDevice(driver.GetAntCaptureFile(*inFile))
	default:
		log.Fatalln("Unknown driver specified!")
	}

	err := device.Start()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer device.Stop()

	go read(device.Read, *outFile)
	device.StartRxScanMode()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
