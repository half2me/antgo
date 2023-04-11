package main

import (
	"context"
	"github.com/half2me/antgo/ant"
	"github.com/half2me/antgo/device"
	"github.com/half2me/antgo/driver/file"
	"github.com/half2me/antgo/driver/usb"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// UDP Client
	udp, err := net.ResolveUDPAddr("udp", ":9999")

	if err != nil {
		log.Fatal(err.Error())
	}

	conn, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		log.Fatal(err.Error())
	}

	//close the connection
	defer conn.Close()

	// workaround for libusb log bug
	log.SetOutput(usb.FixLibUsbLog(log.Writer()))

	//dev, err := usb.AutoDetectDevice()
	dev, err := file.NewEmulator("examples/123.cap")
	defer dev.Close()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// initialize device
	err = device.Reset(dev)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start RX Scan mode
	err = device.StartRxScanMode(dev)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start reading broadcast messages
	messages := make(chan ant.BroadcastMessage)
	go device.DumpBroadcastMessages(ctx, dev, messages)

	for msg := range messages {
		<-time.After(500 * time.Millisecond)
		log.Println(msg)
		// dump to UDP
		_, err = conn.Write(msg)
		if err != nil {
			log.Printf("Write failed: %s\n", err.Error())
		}
	}
}
