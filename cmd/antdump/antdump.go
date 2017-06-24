package main

import (
"log"
"fmt"
"github.com/half2me/antgo/driver"
"github.com/half2me/antgo/message"
	"flag"
	"os"
	"os/signal"
	"github.com/gorilla/websocket"
	"net/url"
)

// Send messages on the input to all outputs
func tee(in chan message.AntPacket, out []chan message.AntPacket) {
	defer func() {
		for _, v := range out {
			close(v)
		}
	}()

	for m := range in {
		for _, v := range out {
			v <- m
		}
	}
}

// Write ANT packets to a file
func writeToFile(in chan message.AntPacket, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	for m := range in {
		f.Write(m)
	}
}

func sendToWs(in chan message.AntPacket, host string) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer c.Close()
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	for m := range in {
		werr := c.WriteMessage(websocket.BinaryMessage, []byte(m))
		if werr != nil {
			log.Println("write:", werr)
		}
	}
}

func filter(in chan message.AntPacket, out chan message.AntPacket) {
	var prevPower message.PowerMessage = nil
	var prevSnC message.SpeedAndCadenceMessage = nil

	for e := range in {
		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			out <- e

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
	drv := flag.String("driver", "usb", "Specify the Driver to use: [usb, serial, file, debug]")
	flag.Bool("raw", false, "Do not attempt to decode ANT+ Broadcast messages")
	pid := flag.Int("pid", 0x1008, "When using the USB driver specify pid of the dongle (i.e.: 0x1008")
	inFile := flag.String("infile", "", "File to read ANT+ data from.")
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

	filtered := make(chan message.AntPacket)
	go filter(device.Read, filtered)

	go read(device.Read, *outFile)
	device.StartRxScanMode()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
