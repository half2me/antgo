package main

import (
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
)

func decode() {
	for m := range wsIn {
		log.Printf("recv: %s", m)
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
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		if mt == websocket.TextMessage {
			wsIn <- message
		}
	}
}

var wsIn chan []byte = make(chan []byte, 16)
var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	defer close(wsIn)
	go decode()
	flag.Parse()
	http.HandleFunc("/", wsFunction)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
