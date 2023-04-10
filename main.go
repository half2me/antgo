package main

import (
	"context"
	"fmt"
	"github.com/half2me/antgo/ant"
	"github.com/half2me/antgo/driver"
	"github.com/half2me/antgo/driver/usb"
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
		log.Fatalf(err.Error())
	}

	//node := driver.NewNode(sniffer.Sniff(dev))
	node := driver.NewNode(dev)

	// initialize node
	err = node.Reset()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start RX Scan mode
	err = node.StartRxScanMode()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start reading broadcast messages
	messages := make(chan ant.AntBroadcastMessage, 10)
	go node.DumpBroadcastMessages(ctx, messages)

	for msg := range messages {
		switch msg.DeviceType() {
		case ant.DEVICE_TYPE_SPEED_AND_CADENCE:
			fmt.Println(ant.SpeedAndCadenceMessage(msg))
		case ant.DEVICE_TYPE_POWER:
			fmt.Println(ant.PowerMessage(msg))
		default:
			fmt.Println(msg)
		}
	}
}
