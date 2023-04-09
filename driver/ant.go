package driver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/half2me/antgo/message"
	"io"
)

type Driver interface {
	Close()
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	BufferSize() int
}

type Node struct {
	driver Driver
	reader *bufio.Reader
}

func NewNode(driver Driver) Node {
	return Node{
		driver: driver,
		reader: bufio.NewReaderSize(driver, driver.BufferSize()),
	}
}

func (node Node) ReadMsg() (p message.AntPacket, err error) {
	// 1st byte is TX SYNC
	sync, err := node.reader.ReadByte()
	if err != nil {
		return
	}
	if sync != message.MESSAGE_TX_SYNC {
		return p, fmt.Errorf("expected TX SYNC, got %02X", sync)
	}

	// 2nd byte is payload length
	length, err := node.reader.ReadByte()
	if err != nil {
		return
	}

	// length +1 byte type + 1 byte checksum
	buf := make([]byte, length+2)

	// Get message content and checksum
	_, err = io.ReadFull(node.reader, buf)
	if err != nil {
		return p, err
	}

	p = append(message.AntPacket{message.MESSAGE_TX_SYNC, length}, buf...)

	// Check message integrity
	if !p.Valid() {
		err = errors.New("invalid checksum")
	}

	return
}

func (node Node) expectAck() error {
	msg, err := node.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Class() != message.MESSAGE_CHANNEL_ACK {
		return fmt.Errorf("expected ACK, got %02X", msg.Class())
	}
	return nil
}

func (node Node) WriteMsg(msg message.AntPacket) (err error) {
	_, err = node.driver.Write(msg)
	if err != nil {
		return err
	}
	return node.expectAck()
}

func (node Node) StartRxScanMode() error {
	messages := []message.AntPacket{
		message.SystemResetMessage(),
		message.SetNetworkKeyMessage(0, []byte(message.ANTPLUS_NETWORK_KEY)),
		message.AssignChannelMessage(0, message.CHANNEL_TYPE_ONEWAY_RECEIVE),
		message.SetChannelIdMessage(0),
		message.SetChannelRfFrequencyMessage(0, 2457),
		message.EnableExtendedMessagesMessage(true),
		// message.LibConfigMessage(true, true, true),
		message.OpenRxScanModeMessage(),
	}

	for _, msg := range messages {
		err := node.WriteMsg(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node Node) DumpBroadcastMessages(ctx context.Context, messages chan message.AntBroadcastMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m, err := node.ReadMsg()
		if err != nil {
			// ignore errors
			continue
		}

		// skip non-broadcast messages
		if m.Class() != message.MESSAGE_TYPE_BROADCAST {
			continue
		}

		messages <- message.AntBroadcastMessage(m)
	}
}
