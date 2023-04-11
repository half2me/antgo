package main

import (
	"bufio"
	"github.com/half2me/antgo/ant"
	"log"
	"net"
)

const (
	HOST = "localhost"
	PORT = "9999"
	TYPE = "tcp"
)

func main() {
	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	// close listener
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	log.Println("New connection!")
	defer conn.Close()
	r := bufio.NewReader(conn)

	for {
		msg, err := ant.ReadMsg(r)
		if err != nil {
			break
		}
		// skip non-broadcast messages
		if msg.Class() != ant.MESSAGE_TYPE_BROADCAST {
			continue
		}
		b := ant.BroadcastMessage(msg)
		log.Println(b)
	}
	log.Println("Client disconnected!")
}
