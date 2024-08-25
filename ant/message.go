package ant

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Packet []byte
type BroadcastMessage Packet
type Rssi struct {
	measurementType, rssi, threshold byte
}

var InvalidChecksumError = errors.New("invalid checksum")

// ReadMsg reads a single ANT message
func ReadMsg(reader io.Reader) (p Packet, err error) {
	// 1st byte is TX SYNC
	buf := make([]byte, 1)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return
	}
	if buf[0] != MESSAGE_TX_SYNC {
		return p, fmt.Errorf("expected TX SYNC, got %02X", buf[0])
	}

	// 2nd byte is payload length
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return
	}
	length := buf[0]

	// length +1 byte type + 1 byte checksum
	buf = make([]byte, length+2)

	// Get message content and checksum
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return p, err
	}

	p = append(Packet{MESSAGE_TX_SYNC, length}, buf...)

	// Check message integrity
	if !p.Valid() {
		err = InvalidChecksumError
	}

	return
}

func (r Rssi) Value() (v int8) {
	_ = binary.Read(bytes.NewReader([]byte{r.rssi}), binary.LittleEndian, &v)
	return
}

func (p Packet) String() (s string) {
	s = fmt.Sprintf("[%02X] [", p.Class())

	for _, v := range p.Data() {
		s += fmt.Sprintf(" %02X ", v)
	}

	s += "]"
	return
}

func (p Packet) Class() byte {
	return p[2]
}

func (p Packet) ClassName() string {
	switch p.Class() {
	case MESSAGE_STARTUP:
		return "Startup"
	case MESSAGE_CHANNEL_ACK:
		return "ACK"
	case MESSAGE_CHANNEL_EVENT:
		return "Channel event"
	default:
		return fmt.Sprintf("Unknown: %02X", p.Class())
	}
}

func (p Packet) Data() []byte {
	return p[3 : len(p)-1]
}

func (p Packet) CalculateChecksum() (chk byte) {
	for _, v := range p[:len(p)-1] {
		chk ^= v
	}
	return
}

func (p Packet) Valid() bool {
	return p.CalculateChecksum() == p[len(p)-1]
}

func (p BroadcastMessage) String() (s string) {
	s = fmt.Sprintf("CH: %d [%d] [%s]", p.Channel(), p.DeviceNumber(), p.DeviceTypeString())

	s += "["

	for _, v := range p.Content() {
		s += fmt.Sprintf(" %02X ", v)
	}

	s += "]"
	return
}

func (p BroadcastMessage) Channel() uint8 {
	return Packet(p).Data()[0]
}

func (p BroadcastMessage) Content() []byte {
	return Packet(p).Data()[1:9]
}

func (p BroadcastMessage) ExtendedContent() []byte {
	return Packet(p).Data()[10:]
}

func (p BroadcastMessage) ExtendedFlag() byte {
	return Packet(p).Data()[9]
}

func (p BroadcastMessage) DeviceNumber() (num uint32) {
	return binary.LittleEndian.Uint32([]byte{
		p.ExtendedContent()[0], p.ExtendedContent()[1], // Device Number uint16
		p.ExtendedContent()[3] >> 4, 0x00, // Extended Device Number -> uint20
	})
}

func (p BroadcastMessage) DeviceType() byte {
	return p.ExtendedContent()[2]
}

func (p BroadcastMessage) DeviceTypeString() string {
	switch p.DeviceType() {
	case DEVICE_TYPE_SPEED_AND_CADENCE:
		return "snc"
	case DEVICE_TYPE_POWER:
		return "pwr"
	case DEVICE_TYPE_FE:
		return "fe"
	case DEVICE_TYPE_SDM:
		return "sdm"
	default:
		return "unknown"
	}
}

func (p BroadcastMessage) TransmissionType() byte {
	return p.ExtendedContent()[3]
}

func (p BroadcastMessage) RssiInfo() Rssi {
	ex := p.ExtendedContent()

	return Rssi{
		ex[4],
		ex[5],
		ex[6],
	}
}

func (p BroadcastMessage) RxTimestamp() (ts uint16) {
	_ = binary.Read(bytes.NewReader(p.ExtendedContent()[8:]), binary.LittleEndian, &ts)
	return
}

func MakeAntPacket(messageType byte, content []byte) Packet {
	p := make([]byte, len(content)+4)

	p[0] = MESSAGE_TX_SYNC
	p[1] = byte(len(content))
	p[2] = messageType
	copy(p[3:], content)

	a := Packet(p)
	a[len(a)-1] = a.CalculateChecksum()

	return a
}

func AckMessage() Packet {
	return MakeAntPacket(MESSAGE_CHANNEL_ACK, []byte{0x00})
}

func SystemResetMessage() Packet {
	return MakeAntPacket(MESSAGE_SYSTEM_RESET, []byte{0x00})
}

func SetNetworkKeyMessage(channel uint8, key []byte) Packet {
	return MakeAntPacket(MESSAGE_NETWORK_KEY, append([]byte{byte(channel)}, key...))
}

func OpenChannelMessage(channel uint8) Packet {
	return MakeAntPacket(MESSAGE_CHANNEL_OPEN, []byte{byte(channel)})
}

func CloseChannelMessage(channel uint8) Packet {
	return MakeAntPacket(MESSAGE_CHANNEL_CLOSE, []byte{byte(channel)})
}

func AssignChannelMessage(channel uint8, typ byte) Packet {
	return MakeAntPacket(MESSAGE_CHANNEL_ASSIGN, []byte{byte(channel), typ, 0x00})
}

func SetChannelIdMessage(channel uint8) Packet {
	return MakeAntPacket(MESSAGE_CHANNEL_ID, []byte{byte(channel), 0x00, 0x00, 0x00, 0x00})
}

func SetChannelRfFrequencyMessage(channel uint8, freq uint16) Packet {
	return MakeAntPacket(MESSAGE_CHANNEL_FREQUENCY, []byte{byte(channel), byte(freq - 2400)})
}

func OpenRxScanModeMessage() Packet {
	return MakeAntPacket(OPEN_RX_SCAN_MODE, []byte{0x00})
}

func EnableExtendedMessagesMessage(enable bool) Packet {
	var opt byte = 0x00
	if enable {
		opt = 0x01
	}
	return MakeAntPacket(MESSAGE_ENABLE_EXT_RX_MESSAGES, []byte{opt})
}

func LibConfigMessage(rxTimestamp, rssi, channelId bool) Packet {
	var opt byte

	if rxTimestamp {
		opt |= EXT_FLAG_TIMESTAMP
	}
	if rssi {
		opt |= EXT_FLAG_RSSI
	}
	if channelId {
		opt |= EXT_FLAG_CHANNEL_ID
	}

	return MakeAntPacket(MESSAGE_LIB_CONFIG, []byte{0x00, opt})
}
