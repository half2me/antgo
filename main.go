package main

import (
	"context"
	"fmt"
	"github.com/half2me/antgo/driver"
	"github.com/half2me/antgo/driver/usb"
	"github.com/half2me/antgo/message"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dev, err := usb.AutoDetectDevice()
	defer dev.Close()
	if err != nil {
		log.Fatalf("Can't open USB device")
	}

	node := driver.NewNode(dev)
	err = node.StartRxScanMode()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start reading broadcast messages
	messages := make(chan message.AntBroadcastMessage)
	go node.DumpBroadcastMessages(ctx, messages)

	for msg := range messages {
		fmt.Println(msg)
	}
}
