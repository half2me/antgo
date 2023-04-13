package device

import (
	"context"
	"fmt"
	"github.com/half2me/antgo/ant"
	"io"
)

func expectMessageClass(driver io.ReadWriter, t byte) (msg ant.Packet, err error) {
	msg, err = ant.ReadMsg(driver)
	if err != nil {
		return msg, err
	}
	if msg.Class() != t {
		return msg, fmt.Errorf("expected %02X, got %02X", t, msg.Class())
	}
	return msg, nil
}

func ResetWait(driver io.ReadWriter) error {
	_, err := driver.Write(ant.SystemResetMessage())
	if err != nil {
		return err
	}
	_, err = expectMessageClass(driver, ant.MESSAGE_STARTUP)
	return err
}

func Reset(driver io.ReadWriter) error {
	_, err := driver.Write(ant.SystemResetMessage())
	return err
}

func StartRxScanModeWait(driver io.ReadWriter) error {
	messages := []ant.Packet{
		ant.SetNetworkKeyMessage(0, []byte(ant.ANTPLUS_NETWORK_KEY)),
		ant.AssignChannelMessage(0, ant.CHANNEL_TYPE_ONEWAY_RECEIVE),
		ant.SetChannelIdMessage(0),
		ant.SetChannelRfFrequencyMessage(0, 2457),
		ant.EnableExtendedMessagesMessage(true),
		// message.LibConfigMessage(true, true, true),
		ant.OpenRxScanModeMessage(),
	}

	for _, msg := range messages {
		_, err := driver.Write(msg)
		if err != nil {
			return err
		}
		_, err = expectMessageClass(driver, ant.MESSAGE_CHANNEL_EVENT)
		if err != nil {
			return err
		}
	}
	return nil
}

func StartRxScanMode(driver io.ReadWriter) error {
	messages := []ant.Packet{
		ant.SetNetworkKeyMessage(0, []byte(ant.ANTPLUS_NETWORK_KEY)),
		ant.AssignChannelMessage(0, ant.CHANNEL_TYPE_ONEWAY_RECEIVE),
		ant.SetChannelIdMessage(0),
		ant.SetChannelRfFrequencyMessage(0, 2457),
		ant.EnableExtendedMessagesMessage(true),
		// message.LibConfigMessage(true, true, true),
		ant.OpenRxScanModeMessage(),
	}

	for _, msg := range messages {
		_, err := driver.Write(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func DumpBroadcastMessages(ctx context.Context, driver io.ReadWriter, messages chan ant.BroadcastMessage) {
	defer close(messages)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m, err := ant.ReadMsg(driver)
		if err != nil {
			// ignore errors
			continue
		}

		// skip non-broadcast messages
		if m.Class() != ant.MESSAGE_TYPE_BROADCAST {
			continue
		}

		messages <- ant.BroadcastMessage(m)
	}
}
