package main

import (
	"context"
	"github.com/half2me/antgo/ant"
	"github.com/half2me/antgo/device"
	"github.com/half2me/antgo/driver/emulator"
	"github.com/half2me/antgo/driver/usb"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"
)

const (
	HOST = "localhost"
	PORT = "9999"
	TYPE = "tcp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// TCP Client
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)

	if err != nil {
		log.Println(err.Error())
		return
	}

	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//close the connection
	defer conn.Close()

	// workaround for libusb log bug
	log.SetOutput(usb.FixLibUsbLog(log.Writer()))

	//dev, err := usb.AutoDetectDevice()
	dev, err := emulator.NewEmulator("examples/123.cap")
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
		<-time.After(500 * time.Millisecond) // Artificial delay for
		log.Println(msg)

		// we only need s&c or power messages
		if !(msg.DeviceType() == ant.DEVICE_TYPE_SPEED_AND_CADENCE ||
			msg.DeviceType() == ant.DEVICE_TYPE_POWER) {
			continue
		}

		// ignore power sensor service packets
		if msg.DeviceType() == ant.DEVICE_TYPE_POWER && ant.PowerMessage(msg).DataPageNumber() != 0x10 {
			continue
		}

		_, err = conn.Write(msg)
		if err != nil {
			log.Printf("Write failed: %s\n", err.Error())
			stop()
		}
	}
}
