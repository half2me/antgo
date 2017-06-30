package main

import (
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/half2me/antgo/message"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Dup struct {
	data map[message.AntPacket]struct{}
	mtx sync.Mutex
}

func (d *Dup) Lock() {
	d.mtx.Lock()
}

func (d *Dup) UnLock() {
	d.mtx.Unlock()
}

// Returns true if message is new, and adds it to the cache. If false, message is already listed
func (d *Dup) Test(m message.AntPacket) (ok bool) {
	d.Lock()
	defer d.UnLock()
	if _, exists := d.data[m]; !exists {
		ok = true
		d.data[m] = struct{}{}
		go func(){
			// Clean out of cache in 10 sec
			time.Sleep(time.Second * 10)
			d.Lock()
			defer d.UnLock()
			delete(d.data, m)
		}()
	}

	return
}

type JsonMessage struct {
	Speed float32 `json:"speed,omitempty"`
	SpeedStall bool `json:"speed_stall,omitempty"`
	Cadence float32 `json:"cadence,omitempty"`
	CadenceStall bool `json:"cadence_stall,omitempty"`
	TotalDistance float32 `json:"total_distance,omitempty"`
	Power float32 `json:"power,omitempty"`
}

type statT struct {
	lastSncMessage message.SpeedAndCadenceMessage
	lastPowMessage message.PowerMessage
	json JsonMessage
}

func decode(in <-chan message.AntPacket, out chan []byte, wheel float32) {
	defer close(out)

	for e := range in {
		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			if dup.Test(e) {
				msg := message.AntBroadcastMessage(e)
				if s, ok := stat[msg.DeviceNumber()]; ok {
					switch msg.DeviceType() {
					case message.DEVICE_TYPE_SPEED_AND_CADENCE:
						prev := s.lastSncMessage
						snc := message.SpeedAndCadenceMessage(msg)
						s.json.TotalDistance += snc.Distance(prev, wheel)
						s.json.Cadence, s.json.CadenceStall = snc.Cadence(prev)
						s.json.Speed, s.json.SpeedStall = snc.Speed(prev, wheel)
						s.lastSncMessage = snc
					case message.DEVICE_TYPE_POWER:
						prev := stat[msg.DeviceNumber()].lastPowMessage
						pm := message.PowerMessage(msg)
						if pm.DataPageNumber() == 0x10 {
							s.json.Power= pm.AveragePower(prev)
							s.lastPowMessage = pm
						} else {
							continue
						}
					default:
						continue
					}

					if j, err := json.Marshal(s.json); err != nil {
						log.Println(err)
					} else {
						out <- j
					}
				}
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

var dup = Dup {
	data: make(map[message.AntPacket]struct{}),
}

var stat = map[uint16]*statT{
	123: {}, // only allow packets from these IDs
	124: {},
}

func main() {
	defer close(antIn)
	go decode(antIn, out, 0.98)
	go pr(out)
	flag.Parse()
	http.HandleFunc("/", wsFunction)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
