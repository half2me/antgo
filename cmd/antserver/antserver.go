package main

import (
	"net/http"
	"log"
	"flag"
	"github.com/gorilla/websocket"
)

func wsFunction(rw http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		c.WriteMessage(websocket.TextMessage, []byte("Hey there :)"))
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
	flag.Parse()
	http.HandleFunc("/", wsFunction)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
