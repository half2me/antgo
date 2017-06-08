package antgo

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

	c:= <- dongle.Read

	fmt.Println(c)
}