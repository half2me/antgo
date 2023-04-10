package node

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/half2me/antgo/ant"
	"github.com/half2me/antgo/driver"
	"io"
)

type Node struct {
	driver driver.Driver
	reader *bufio.Reader
}

func NewNode(driver driver.Driver) Node {
	return Node{
		driver: driver,
		reader: bufio.NewReaderSize(driver, driver.BufferSize()),
	}
}

func (node Node) ReadMsg() (p ant.AntPacket, err error) {
	// 1st byte is TX SYNC
	sync, err := node.reader.ReadByte()
	if err != nil {
		return
	}
	if sync != ant.MESSAGE_TX_SYNC {
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

	p = append(ant.AntPacket{ant.MESSAGE_TX_SYNC, length}, buf...)

	// Check message integrity
	if !p.Valid() {
		err = errors.New("invalid checksum")
	}

	return
}

func (node Node) expectMessageClass(t byte) (msg ant.AntPacket, err error) {
	msg, err = node.ReadMsg()
	if err != nil {
		return msg, err
	}
	if msg.Class() != t {
		return msg, fmt.Errorf("expected %02X, got %02X", t, msg.Class())
	}
	return msg, nil
}

func (node Node) WriteMsg(msg ant.AntPacket) (err error) {
	_, err = node.driver.Write(msg)
	if err != nil {
		return err
	}
	return err
}

func (node Node) Reset() error {
	err := node.WriteMsg(ant.SystemResetMessage())
	if err != nil {
		return err
	}
	_, err = node.expectMessageClass(ant.MESSAGE_STARTUP)
	return err
}

func (node Node) StartRxScanMode() error {
	messages := []ant.AntPacket{
		ant.SetNetworkKeyMessage(0, []byte(ant.ANTPLUS_NETWORK_KEY)),
		ant.AssignChannelMessage(0, ant.CHANNEL_TYPE_ONEWAY_RECEIVE),
		ant.SetChannelIdMessage(0),
		ant.SetChannelRfFrequencyMessage(0, 2457),
		ant.EnableExtendedMessagesMessage(true),
		// message.LibConfigMessage(true, true, true),
		ant.OpenRxScanModeMessage(),
	}

	for _, msg := range messages {
		err := node.WriteMsg(msg)
		if err != nil {
			return err
		}
		_, err = node.expectMessageClass(ant.MESSAGE_CHANNEL_EVENT)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node Node) DumpBroadcastMessages(ctx context.Context, messages chan ant.AntBroadcastMessage) {
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
		if m.Class() != ant.MESSAGE_TYPE_BROADCAST {
			continue
		}

		messages <- ant.AntBroadcastMessage(m)
	}
}
