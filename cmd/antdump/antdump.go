package main

import (
"log"
"fmt"
"github.com/half2me/antgo/driver"
"github.com/half2me/antgo/message"
	"flag"
	"os"
	"os/signal"
)

func read(r chan message.AntPacket, raw bool) {
	for e := range r {
		if e.Class() == message.MESSAGE_TYPE_BROADCAST {
			msg := message.AntBroadcastMessage(e)
			if raw {
				fmt.Println(msg)
			} else {
				fmt.Println("Decoding not supported yet")
			}
		}
	}
}

func main() {
	raw := flag.Bool("raw", true, "do not attempt to decode ANT+ Broadcast messages")
	pid := flag.Int("pid", 0x1008, "Specify pid of USB Ant dongle")

	dongle := driver.GetUsbDevice(0x0fcf, *pid)
	err := dongle.Open()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer dongle.Close()

	go read(dongle.Read, *raw)

	dongle.StartRxScanMode()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
