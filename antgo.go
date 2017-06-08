package main

import (
	"log"
	"time"
)

func main() {
	dongle := GetDevice(0x0fcf, 0x1008)
	err := dongle.Open()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer dongle.Close()

	time.Sleep(time.Second * 2)
}