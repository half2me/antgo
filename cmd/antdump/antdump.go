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
	"github.com/half2me/antgo/driver/usb"
	"github.com/half2me/antgo/driver/file"
)

func sendToWs(in <-chan message.AntPacket, done chan<- struct{}) {
	defer func() {done<-struct {}{}}()
	u, errp := url.Parse(*wsAddr)
	if errp != nil {
		panic(errp.Error())
	}

	var c *websocket.Conn
	var err error
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

wsconnect: // Connect to the websocket server
	if c, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		if ! *persistent {
			log.Fatalln(err.Error())
		}
		log.Println(err.Error())
		time.Sleep(time.Second)
		goto wsconnect
	}

	// Register as source
	if err := c.WriteMessage(websocket.TextMessage, []byte("source")); err != nil {
		c.Close()
		if ! *persistent {
			log.Fatalln(err.Error())
		}
		log.Println(err.Error())
		goto wsconnect
	}

	// Setup pingpongs
	c.SetReadDeadline(time.Now().Add(time.Second * 3))
	c.SetPongHandler(func(string) error {c.SetReadDeadline(time.Now().Add(time.Second * 3)); return nil})
	go func(){for {if _, _, err := c.ReadMessage(); err != nil {c.Close(); return}}}()

	// Send ANT+ messages or pings
	log.Println("Connected to ws server!")
	for {
		select {
		case <- ticker.C:
			c.WriteMessage(websocket.PingMessage, []byte{})
		case msg, ok := <- in:
			if !ok {return}
			if e := c.WriteMessage(websocket.BinaryMessage, msg); e != nil {
				c.Close()
				if ! *persistent {
					log.Fatalln(e.Error())
				}
				log.Println(e.Error())
				goto wsconnect
			}
		}
	}

	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Antdump exiting"))
	c.Close()
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

	var f *file.AntCaptureFile

	//File
	if len(*outfile) > 0 {
		f = file.GetAntCaptureFile(*outfile)
		if e := f.Open(); e != nil {panic(e.Error())}
		defer f.Close()
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
				select {
				case c <- m:
				default:
				}
			}

			if f != nil {
				f.Write(m)
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

var stopFile = make(chan struct{})

func main() {
	flag.Parse()

	if *persistent {
		log.Println("Persistent mode actvated!")
	}

	antIn := make(chan message.AntPacket)
	antOut := make(chan message.AntPacket)
	done := make(chan struct{})
	defer func() {<-done}()

	switch *drv {
	case "usb":
		device := driver.NewDevice(usb.GetUsbDevice(0x0fcf, *pid), antIn, antOut)
		if err := device.Start(); err != nil {panic(err.Error())}
		defer device.Stop()
		device.StartRxScanMode()
	case "file":
		f := file.GetAntCaptureFile(*inFile)
		if e := f.Open(); e != nil {
			panic(e.Error())
		}
		go f.ReadLoop(antIn, stopFile)
		defer func(){stopFile <- struct{}{}}()
	default:
		panic("Unknown driver specified!")
	}

	go loop(antIn, done)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
