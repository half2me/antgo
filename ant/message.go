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

// extendedBlocks splits the extended content into the blocks the flag byte
// announces. They appear in this order and each is optional, so a block's
// offset depends on which ones precede it — and the RSSI block's width is not
// fixed either: the measurement type byte leading it decides whether it carries
// a dBm reading and a threshold, or a threshold offset and a 16-bit AGC
// register. A block the flag announces but the message is too short to hold
// comes back nil, along with everything after it.
func (p BroadcastMessage) extendedBlocks() (channelId, rssi, timestamp []byte) {
	data := Packet(p).Data()
	// A standard broadcast is the channel and the payload and stops there, with
	// no flag byte to read: nothing is announced, so nothing is present.
	if len(data) <= extFlagIndex {
		return
	}
	flag := data[extFlagIndex]
	rest := data[extFlagIndex+1:]

	take := func(n int) []byte {
		if len(rest) < n {
			rest = nil
			return nil
		}
		block := rest[:n]
		rest = rest[n:]
		return block
	}

	if flag&EXT_FLAG_CHANNEL_ID != 0 {
		channelId = take(extChannelIdSize)
	}
	if flag&EXT_FLAG_RSSI != 0 {
		if len(rest) == 0 {
			return
		}
		rssi = take(rssiBlockSize(rest[0]))
	}
	if flag&EXT_FLAG_TIMESTAMP != 0 {
		timestamp = take(extTimestampSize)
	}
	return
}

func rssiBlockSize(measurementType byte) int {
	if measurementType == RSSI_MEASUREMENT_TYPE_AGC {
		return extRssiAgcSize
	}
	return extRssiDbmSize
}

// RssiInfo reports the dBm reading, if the dongle sent one. An AGC measurement
// is not one: its bytes are a threshold offset and a register, so reading them
// as a dBm value would produce a plausible-looking number that is not a signal
// strength.
func (p BroadcastMessage) RssiInfo() (Rssi, bool) {
	_, rssi, _ := p.extendedBlocks()
	if rssi == nil || rssi[0] != RSSI_MEASUREMENT_TYPE_DBM {
		return Rssi{}, false
	}
	return Rssi{rssi[0], rssi[1], rssi[2]}, true
}

// RxTimestamp reports when the dongle's own 32768 Hz clock heard the message,
// if it sent one. The counter is 16 bits, so it wraps every two seconds.
func (p BroadcastMessage) RxTimestamp() (uint16, bool) {
	_, _, timestamp := p.extendedBlocks()
	if timestamp == nil {
		return 0, false
	}
	return binary.LittleEndian.Uint16(timestamp), true
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
