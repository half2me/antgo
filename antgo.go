package main

import (
	"log"
	"fmt"
	"time"
	"github.com/half2me/antgo/driver"
	"github.com/half2me/antgo/message"
	"github.com/half2me/antgo/constants"
)

func read(r chan message.AntPacket) {
	for e := range r {
		if e.Class() == constants.MESSAGE_TYPE_BROADCAST {
			msg := message.AntBroadcastMessage(e)
			fmt.Println(msg.ExtendedContent())
			fmt.Println(msg.RxTimestamp())
		}
	}
}

func main() {
	dongle := driver.GetUsbDevice(0x0fcf, 0x1008)
	err := dongle.Open()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer dongle.Close()

	go read(dongle.Read)

	dongle.StartRxScanMode()

	time.Sleep(time.Second * 30)
}