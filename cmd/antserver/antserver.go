package main

import (
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/half2me/antgo/message"
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

func wsFunction(rw http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		mt, m, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		if mt == websocket.BinaryMessage {
			antIn <- message.AntPacket(m)
		}
	}
}

func pr(c chan []byte) {
	for x:= range c {
		fmt.Println(string(x))
	}
}

var antIn chan message.AntPacket = make(chan message.AntPacket, 16)
var out chan []byte = make(chan []byte)
var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	defer close(antIn)
	go decode(antIn, out, 0.98)
	go pr(out)
	flag.Parse()
	http.HandleFunc("/", wsFunction)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
