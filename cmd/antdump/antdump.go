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
	"fmt"
	"time"
)

// Write ANT packets to a file
func writeToFile(in <-chan message.AntPacket, done chan<- struct{}) {
	defer func() {done<-struct {}{}}()
	f, err := os.Create(*outfile)
	if err != nil {
		panic(err.Error())
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
		panic(errp.Error())
	}

	var c *websocket.Conn
	var err error

	for { // Connect to the websocket server
		if c, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
			if ! *persistent {
				panic(err.Error())
			}
			log.Fatalln(err.Error())
		} else {
			// Register as source
			if err := c.WriteMessage(websocket.TextMessage, []byte("source")); err != nil {
				c.Close()
				if ! *persistent {
					panic(err.Error())
				}
				log.Fatalln(err.Error())
				continue
			}

			// Send ANT+ messages
			for m := range in {
				if e := c.WriteMessage(websocket.BinaryMessage, m); e != nil {
					log.Fatalln(e.Error())
					c.Close()
					if ! *persistent {
						panic(err.Error())
					}
					continue
				}
			}
		}
		time.Sleep(time.Second) // Sleep a bit, then retry
	}

	defer c.Close()
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func filter(m message.AntPacket) (allow bool) {
	if m.Class() == message.MESSAGE_TYPE_BROADCAST {
		msg := message.AntBroadcastMessage(m)
		switch msg.DeviceType() {
		case message.DEVICE_TYPE_SPEED_AND_CADENCE:
			allow = true
		case message.DEVICE_TYPE_POWER:
			if message.PowerMessage(msg).DataPageNumber() == 0x10 {
				allow = true
			}
		}
	}
	return
}

func loop(in <-chan message.AntPacket, done chan<- struct{}) {
	defer func() {done<-struct {}{}}()

	outs := make([]chan message.AntPacket, 0, 2)

	//File
	if len(*outfile) > 0 {
		c := make(chan message.AntPacket)
		cdone := make(chan struct{})
		go writeToFile(c, cdone)
		defer func() {<-cdone}()
		outs = append(outs, c)
	}

	// Ws
	if len(*wsAddr) > 0 {
		c := make(chan message.AntPacket)
		cdone := make(chan struct{})
		go sendToWs(c, cdone)
		defer func() {<-cdone}()
		outs = append(outs, c)
	}

	defer func() {for _, c := range outs {close(c)}}()

	for m := range in {
		if filter(m) {
			if ! *silent {
				fmt.Println(message.AntBroadcastMessage(m))
			}
			for _, c := range outs {
				c <- m
			}
		}
	}
}

var drv = flag.String("driver", "usb", "Specify the Driver to use: [usb, serial, file, debug]")
var pid = flag.Int("pid", 0x1008, "When using the USB driver specify pid of the dongle (i.e.: 0x1008")
var inFile = flag.String("infile", "", "File to read ANT+ data from.")
var outfile = flag.String("outfile", "", "File to dump ANT+ data to.")
var wsAddr = flag.String("ws", "", "Upload ANT+ data to a websocket server at address:...")
var silent = flag.Bool("silent", false, "Don't show ANT+ data on terminal")
var persistent = flag.Bool("persistent", false, "Don't panic on errors, keep trying")

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

	if err := device.Start(); err != nil {
		panic(err.Error())
	}

	done := make(chan struct{})
	go loop(device.Read, done)
	defer func() {<-done}()
	defer device.Stop()

	device.StartRxScanMode()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
