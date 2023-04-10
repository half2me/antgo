package main

import (
	"context"
	"fmt"
	"github.com/half2me/antgo/ant"
	"github.com/half2me/antgo/driver/usb"
	"github.com/half2me/antgo/node"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// workaround for libusb log bug
	log.SetOutput(usb.FixLibUsbLog(log.Writer()))

	dev, err := usb.AutoDetectDevice()
	defer dev.Close()
	if err != nil {
		log.Fatalf(err.Error())
	}

	//n := driver.NewNode(sniffer.Sniff(dev))
	n := node.NewNode(dev)

	// initialize node
	err = n.Reset()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start RX Scan mode
	err = n.StartRxScanMode()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start reading broadcast messages
	messages := make(chan ant.AntBroadcastMessage, 10)
	go n.DumpBroadcastMessages(ctx, messages)

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
