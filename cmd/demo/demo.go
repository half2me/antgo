package main

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"flag"
	"time"
	"sync"
	"encoding/json"
	"math/rand"
)

type JsonMessage struct {
	Speed float32 `json:"speed"`
	SpeedStall bool `json:"speed_stall"`
	Cadence float32 `json:"cadence"`
	CadenceStall bool `json:"cadence_stall"`
	TotalDistance float32 `json:"total_distance"`
	DistanceDelta float32 `json:"distance_delta"`
	Power float32 `json:"power"`
}

func ws(rw http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// Send updates
	for {
		statMutex.Lock()
		if j, err := json.Marshal(stats); err != nil {
			log.Println(err)
		} else {
			if err := c.WriteMessage(websocket.TextMessage, j); err != nil {
				statMutex.Unlock()
				return
			}
		}
		statMutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func tickSnc() {
	for {
		statMutex.Lock()
		stats.DistanceDelta = rand.Float32() * 3
		stats.TotalDistance += stats.DistanceDelta
		stats.Speed = rand.Float32() * 40
		stats.SpeedStall = false
		stats.Cadence = rand.Float32() * 40
		stats.CadenceStall = false
		statMutex.Unlock()

		time.Sleep(3 * time.Second)
	}
}

func tickSpeedStall() {
	for {
		statMutex.Lock()
		stats.DistanceDelta = 0
		stats.Speed = 0
		stats.SpeedStall = true
		statMutex.Unlock()

		time.Sleep(1100 * time.Millisecond)
	}
}

func tickCadStall() {
	for {
		statMutex.Lock()
		stats.CadenceStall = true
		stats.Cadence = 0
		statMutex.Unlock()

		time.Sleep(1200 * time.Millisecond)
	}
}

func tickPower() {
	for {
		statMutex.Lock()
		stats.Power = rand.Float32() * 3
		statMutex.Unlock()

		time.Sleep(1 * time.Second)
	}
}



var addr = flag.String("addr", "localhost:8081", "WebSocket Server Address")
var stats = JsonMessage{}
var statMutex = sync.Mutex{}

func main() {
	flag.Parse()
	go tickSnc()
	go tickPower()
	go tickSpeedStall()
	go tickCadStall()
	http.HandleFunc("/", ws)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
