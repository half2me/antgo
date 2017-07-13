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

// A simple cache for deduplicating incoming packets
type Dup struct {
	data map[string]struct{}
	mtx sync.Mutex
	timeout uint
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
	if _, exists := d.data[string(m)]; !exists {
		ok = true
		d.data[string(m)] = struct{}{}
		go func(){
			// Clean out of cache in 5 sec
			time.Sleep(time.Second * time.Duration(d.timeout))
			d.Lock()
			defer d.UnLock()
			delete(d.data, string(m))
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
	DistanceDelta float32 `json:"distance_delta,omitempty"`
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
						s.json.DistanceDelta = snc.Distance(prev, wheel)
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

					if j, err := json.Marshal(map[uint16]JsonMessage{
						msg.DeviceNumber(): s.json,
					}); err != nil {
						log.Println(err)
					} else {
						out <- j
					}
				}
			}
		}
	}
}

func wsHandler(rw http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	defer func(){log.Println("Closing ws connection")}()

	// Decide if source or sink
	if mt, m, err := c.ReadMessage(); err == nil && mt == websocket.TextMessage {
		switch string(m) {
		case "source":
			log.Println("Source connected!")
			for {
				if mt, m, err := c.ReadMessage(); err == nil && mt == websocket.BinaryMessage {
					antIn <- message.AntPacket(m)
				} else {
					log.Println("read:", err)
					break
				}
			}
		case "sink":
			log.Println("Sink connected!")
			ch := make(chan []byte, 4)
			sinks.RegisterSink(c.RemoteAddr().String(), ch)
			defer func(){sinks.UnregisterSink(c.RemoteAddr().String())}()

			for dat := range ch {
				if err := c.WriteMessage(websocket.TextMessage, dat); err != nil {
					return
				}
			}
		default:
			c.WriteMessage(websocket.TextMessage, []byte("Bad initial message, should be sink or source"))
			log.Fatalln("Bad initial message, should be sink or source")
		}
	} else {
		log.Fatalln("Initial message not received or incorrect format!")
	}
}

func pr(c chan []byte) {
	for x:= range c {
		sinks.Broadcast(x)
		fmt.Println(string(x))
	}
}

var antIn chan message.AntPacket = make(chan message.AntPacket, 16)
var out chan []byte = make(chan []byte)
var addr = flag.String("addr", "localhost:8080", "http service address")

type sinkStore struct {
	sinks map[string]chan []byte
	mutex sync.Mutex
}

func (s *sinkStore) RegisterSink(addr string, c chan []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.sinks[addr]; ok == true {
		// This sink is already registered
		delete(s.sinks, addr)
	}

	s.sinks[addr] = c
}

func (s *sinkStore) UnregisterSink(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.sinks[addr]; ok == true {
		delete(s.sinks, addr)
	}
}

func (s *sinkStore) Broadcast(msg []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, v := range s.sinks {
		v <- msg
	}
}

func (s *sinkStore) CloseAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, v := range s.sinks {
		close(v)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var dup = Dup {
	data: make(map[string]struct{}),
	timeout: 5,
}

var stat = map[uint16]*statT{
	123: {}, // only allow packets from these IDs
	456: {},
	789: {},
}

var sinks = sinkStore{
	sinks: make(map[string]chan []byte),
}

func main() {
	defer close(antIn)
	go decode(antIn, out, 0.98)
	go pr(out)
	flag.Parse()
	http.HandleFunc("/", wsHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
