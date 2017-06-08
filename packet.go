package main

import "fmt"

const (
	MESSAGE_TX_SYNC = 0xA4
	MESSAGE_SYSTEM_RESET = 0x4A
	MESSAGE_STARTUP = 0x6F
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