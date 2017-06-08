package main

import (
	"log"
	"fmt"
	"time"
)

func read(r chan []byte) {
	for e := range r {
		fmt.Println(AntPacket(e).String())
	}
}

func main() {
	dongle := GetDevice(0x0fcf, 0x1008)
	err := dongle.Open()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer dongle.Close()

	go read(dongle.Read)

	dongle.StartRxScanMode()

	time.Sleep(time.Second * 20)
}