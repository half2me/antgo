package main

import "fmt"

const (
	MESSAGE_TX_SYNC = 0xA4
	MESSAGE_SYSTEM_RESET = 0x4A
	MESSAGE_STARTUP = 0x6F
	MESSAGE_NETWORK_KEY = 0x46
	ANTPLUS_NETWORK_KEY = "\xB9\xA5\x21\xFB\xBD\x72\xC3\x45"
	OPEN_RX_SCAN_MODE = 0x5B

	MESSAGE_ENABLE_EXT_RX_MESSAGES = 0x66
	MESSAGE_LIB_CONFIG = 0x6E

	// Extended message flags
	EXT_FLAG_CHANNEL_ID = 0x80
	EXT_FLAG_RSSI = 0x40
	EXT_FLAG_TIMESTAMP = 0x20
)

type AntPacket []byte

func (p AntPacket) String() (s string) {
	s = fmt.Sprintf("[%02X] [", p[2])

	for _, v := range p[3:len(p)-1] {
		s += fmt.Sprintf(" %02X ", v)
	}

	s += "]"
	return
}

func makeAntPacket(messageType byte, content []byte) AntPacket {
	p := make([]byte, len(content) + 4)

	p[0] = MESSAGE_TX_SYNC
	p[1] = byte(len(content))
	p[2] = messageType
	copy(p[3:], content)

	// checksum
	for _, v := range p[:len(p)-1] {
		p[len(p)-1] ^= v
	}

	return p
}

func makeSystemResetMessage() AntPacket {
	return makeAntPacket(MESSAGE_SYSTEM_RESET, []byte{0x00})
}

func makeSetNetworkKeyMessage() AntPacket {
	return makeAntPacket(MESSAGE_NETWORK_KEY, []byte(ANTPLUS_NETWORK_KEY))
}

func makeOpenRxScanModeMessage() AntPacket {
	return makeAntPacket(OPEN_RX_SCAN_MODE, []byte{0x00})
}

func makeEnableExtendedMessagesMessage(enable bool) AntPacket {
	var opt byte = 0x00
	if enable {
		opt = 0x01
	}
	return makeAntPacket(MESSAGE_ENABLE_EXT_RX_MESSAGES, []byte{opt})
}

func makeLibConfigMessage(rxTimestamp, rssi, channelId bool) AntPacket {
	var opt byte = 0x00

	if rxTimestamp {
		opt |= EXT_FLAG_TIMESTAMP
	}
	if rssi {
		opt |= EXT_FLAG_RSSI
	}
	if channelId {
		opt |= EXT_FLAG_CHANNEL_ID
	}

	return makeAntPacket(MESSAGE_ENABLE_EXT_RX_MESSAGES, []byte{0x00, opt})
}