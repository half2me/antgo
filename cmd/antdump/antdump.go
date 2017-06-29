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

type JsonPowerMessage struct {
	Power float32 `json:"power"`
}

type JsonSnCMessage struct {
	Speed float32 `json:"speed"`
	SpeedStall bool `json:"speed_stall"`
	Cadence float32 `json:"cadence"`
	CadenceStall bool `json:"cadence_stall"`
	Distance float32 `json:"distance"`
}

// Send messages on the input to all outputs
func tee(in <-chan []byte, out []chan<- []byte) {
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
func writeToFile(in <-chan message.AntPacket, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	for m := range in {
		f.Write(m)
	}
}

func sendToWs(in <-chan []byte, host string) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer c.Close()
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	for m := range in {
		werr := c.WriteMessage(websocket.TextMessage, m)
		if werr != nil {
			log.Println("write:", werr)
		}
	}
}

func decode(in <-chan message.AntPacket, out chan []byte, wheel float32) {
	var prevPower message.PowerMessage = nil
	var prevSnC message.SpeedAndCadenceMessage = nil

	defer close(out)

	for e := range in {
		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			msg := message.AntBroadcastMessage(e)
			dec := make(map[uint16]interface{})

			switch msg.DeviceType() {
			case message.DEVICE_TYPE_SPEED_AND_CADENCE:
				cad, cad_stall := message.SpeedAndCadenceMessage(msg).Cadence(prevSnC)
				speed, speed_stall := message.SpeedAndCadenceMessage(msg).Speed(prevSnC, wheel)
				dist := message.SpeedAndCadenceMessage(msg).Distance(prevSnC, wheel)
				dec[msg.DeviceNumber()] = JsonSnCMessage{
					speed,
					speed_stall,
					cad,
					cad_stall,
					dist,
				}
				prevSnC = message.SpeedAndCadenceMessage(msg)
			case message.DEVICE_TYPE_POWER:
				if message.PowerMessage(msg).DataPageNumber() == 0x10 {
					pow := message.PowerMessage(msg).AveragePower(prevPower)
					dec[msg.DeviceNumber()] = JsonPowerMessage{
						pow,
					}
					prevPower = message.PowerMessage(msg)
				} else {
					continue
				}
			default:
				continue
			}

			if j, err := json.Marshal(dec); err != nil {
				log.Println(err)
			} else {
				out <- j
			}
		}
	}
}

func show(in <-chan []byte) {
	for m := range in {
		fmt.Println(string(m))
	}
}

func main() {
	drv := flag.String("driver", "file", "Specify the Driver to use: [usb, serial, file, debug]")
	pid := flag.Int("pid", 0x1008, "When using the USB driver specify pid of the dongle (i.e.: 0x1008")
	inFile := flag.String("infile", "capture/123.cap", "File to read ANT+ data from.")
	wheel := flag.Int("wheel", 98, "Wheel circumference in mm")
	flag.String("outfile", "", "File to dump ANT+ data to.")
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

	err := device.Start();

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer device.Stop()

	p := make(chan []byte)
	go decode(device.Read, p, float32(*wheel)/100)
	go show(p)

	device.StartRxScanMode()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
