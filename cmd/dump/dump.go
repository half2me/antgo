package main

import (
	"context"
	"github.com/half2me/antgo/ant"
	"github.com/half2me/antgo/device"
	"github.com/half2me/antgo/driver/usb"
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
		log.Println(err.Error())
		return
	}

	// initialize device
	err = device.Reset(dev)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Start RX Scan mode
	err = device.StartRxScanMode(dev)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Start reading broadcast messages
	messages := make(chan ant.BroadcastMessage)
	go device.DumpBroadcastMessages(ctx, dev, messages)

	for msg := range messages {
		log.Println(msg)
	}
}
