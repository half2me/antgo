package message

import (
	"fmt"
	"github.com/half2me/antgo/constants"
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

	p[0] = constants.MESSAGE_TX_SYNC
	p[1] = byte(len(content))
	p[2] = messageType
	copy(p[3:], content)

	// checksum
	for _, v := range p[:len(p)-1] {
		p[len(p)-1] ^= v
	}

	return p
}

func SystemResetMessage() AntPacket {
	return makeAntPacket(constants.MESSAGE_SYSTEM_RESET, []byte{0x00})
}

func SetNetworkKeyMessage(channel byte, key []byte) AntPacket {
	return makeAntPacket(constants.MESSAGE_NETWORK_KEY, append([]byte{channel}, key...))
}

func AssignChannelMessage(channel, typ byte) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_ASSIGN, []byte{channel, typ, 0x00})
}

func SetChannelIdMessage(channel byte) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_ID, []byte{channel, 0x00, 0x00, 0x00, 0x00})
}

func SetChannelRfFrequencyMessage(channel byte, freq uint16) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_FREQUENCY, []byte{channel, byte(freq - 2400)})
}

func OpenRxScanModeMessage() AntPacket {
	return makeAntPacket(constants.OPEN_RX_SCAN_MODE, []byte{0x00})
}

func EnableExtendedMessagesMessage(enable bool) AntPacket {
	var opt byte = 0x00
	if enable {
		opt = 0x01
	}
	return makeAntPacket(constants.MESSAGE_ENABLE_EXT_RX_MESSAGES, []byte{opt})
}

func LibConfigMessage(rxTimestamp, rssi, channelId bool) AntPacket {
	var opt byte = 0x00

	if rxTimestamp {
		opt |= constants.EXT_FLAG_TIMESTAMP
	}
	if rssi {
		opt |= constants.EXT_FLAG_RSSI
	}
	if channelId {
		opt |= constants.EXT_FLAG_CHANNEL_ID
	}

	return makeAntPacket(constants.MESSAGE_ENABLE_EXT_RX_MESSAGES, []byte{0x00, opt})
}