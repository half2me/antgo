package main

import (
"log"
"github.com/half2me/antgo/driver"
"github.com/half2me/antgo/message"
	"flag"
	"os"
	"os/signal"
	"github.com/gorilla/websocket"
	"net/url"
	"encoding/json"
	"fmt"
)

// Write ANT packets to a file
func writeToFile(in <-chan message.AntPacket, done chan<- struct{}) {
	defer func() {done<-struct {}{}}()
	f, err := os.Create(*outfile)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	for m := range in {
		f.Write(m)
	}
}

func sendToWs(in <-chan message.AntPacket, done chan<- struct{}) {
	defer func() {done<-struct {}{}}()
	u, errp := url.Parse(*wsAddr)
	if errp != nil {
		panic(errp)
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}

	defer c.Close()
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	for m := range in {
		if e := c.WriteMessage(websocket.BinaryMessage, m); e != nil {
			log.Println("write:", e)
		}
	}
}

func printToTerminal(in <-chan message.AntPacket) {
	for m := range in {
		fmt.Println(m)
	}
}

func filter(in <-chan message.AntPacket, out chan<- message.AntPacket) {
	defer close(out)

	for e := range in {
		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			msg := message.AntBroadcastMessage(e)

			switch msg.DeviceType() {
			case message.DEVICE_TYPE_SPEED_AND_CADENCE:
				out <- e
			case message.DEVICE_TYPE_POWER:
				if message.PowerMessage(msg).DataPageNumber() == 0x10 {
					out <- e
				}
			}
		}
	}
}

func loop(in <-chan message.AntPacket, done chan<- struct{}) {
	defer func() {done<-struct {}{}}()

}

var drv = flag.String("driver", "file", "Specify the Driver to use: [usb, serial, file, debug]")
var pid = flag.Int("pid", 0x1008, "When using the USB driver specify pid of the dongle (i.e.: 0x1008")
var inFile = flag.String("infile", "", "File to read ANT+ data from.")
var outfile = flag.String("outfile", "", "File to dump ANT+ data to.")
var wsAddr = flag.String("uploadwsaddr", "", "Upload ANT+ data to a websocket server at address:...")
var silent = flag.Bool("silent", false, "Don't show ANT+ data on terminal")

func main() {
	flag.Parse()

	var device *driver.AntDevice

	switch *drv {
	case "usb":
		device = driver.NewDevice(driver.GetUsbDevice(0x0fcf, *pid))
	case "file":
		device = driver.NewDevice(driver.GetAntCaptureFile(*inFile))
	default:
		panic("Unknown driver specified!")
	}

	err := device.Start();

	if err != nil {
		panic(err)
	}

	done := make(chan struct{})
	go loop(device.Read, done)

	defer device.Stop()

	device.StartRxScanMode()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
