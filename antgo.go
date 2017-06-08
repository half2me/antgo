package main

import (
	"log"
	"fmt"
)

func main() {
	dongle := GetDevice(0x0fcf, 0x1008)
	err := dongle.Open()

	if err != nil {
		log.Fatalln(err)
		return
	}

	defer dongle.Close()

	rst := makeSystemResetMessage()
	fmt.Println(rst.String())
	dongle.Write <- rst
	c := <- dongle.Read
	fmt.Println(AntPacket(c).String())
}