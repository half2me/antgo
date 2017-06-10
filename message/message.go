package message

import (
	"fmt"
	"github.com/half2me/antgo/constants"
	"encoding/binary"
	"bytes"
)

type AntPacket []byte
type AntBroadcastMessage AntPacket
type Rssi struct{
	measurementType, rssi, threshold byte
}

func (r Rssi) Value() (v int8) {
	binary.Read(bytes.NewReader([]byte{r.rssi}), binary.LittleEndian, &v)
	return
}
/*
func (r Rssi) String() string {
	return fmt.Sprintf("%d dBm", r.Value())
}
*/

func (p AntPacket) String() (s string) {
	s = fmt.Sprintf("[%02X] [", p.Class())

	for _, v := range p.Data() {
		s += fmt.Sprintf(" %02X ", v)
	}

	s += "]"
	return
}

func (p AntPacket) Class() byte {
	return p[2]
}

func (p AntPacket) Data() []byte {
	return p[3:len(p)-1]
}

func (p AntPacket) CalculateChecksum() (chk byte) {
	for _, v := range p[:len(p)-1] {
		chk ^= v
	}
	return
}

func (p AntPacket) Valid() bool {
	return p.CalculateChecksum() == p[len(p)-1]
}

func (p AntBroadcastMessage) String() (s string) {
	s = fmt.Sprintf("CH: %d ", p.Channel())
	s += fmt.Sprintf("[%d] ", p.DeviceNumber())
	s += fmt.Sprintf("[%s] ", constants.DeviceTypes[p.DeviceType()])

	s += "["

	for _, v := range p.Content() {
		s += fmt.Sprintf(" %02X ", v)
	}

	s += "]"
	return
}

func (p AntBroadcastMessage) Channel() uint8 {
	return uint8(AntPacket(p).Data()[0])
}

func (p AntBroadcastMessage) Content() []byte {
	return AntPacket(p).Data()[1:9]
}

func (p AntBroadcastMessage) ExtendedContent() []byte {
	return AntPacket(p).Data()[10:]
}

func (p AntBroadcastMessage) ExtendedFlag() byte {
	return AntPacket(p).Data()[9]
}

func (p AntBroadcastMessage) DeviceNumber() (num uint16) {
	binary.Read(bytes.NewReader(p.ExtendedContent()[:2]), binary.LittleEndian, &num)
	return
}

func (p AntBroadcastMessage) DeviceType() byte {
	return p.ExtendedContent()[2]
}

func (p AntBroadcastMessage) TransmissionType() byte {
	return p.ExtendedContent()[3]
}

func (p AntBroadcastMessage) RssiInfo() Rssi {
	ex := p.ExtendedContent()

	return Rssi{
		ex[4],
		ex[5],
		ex[6],
	}
}

func (p AntBroadcastMessage) RxTimestamp() (ts uint16) {
	binary.Read(bytes.NewReader(p.ExtendedContent()[8:]), binary.LittleEndian, &ts)
	return
}

func makeAntPacket(messageType byte, content []byte) AntPacket {
	p := make([]byte, len(content) + 4)

	p[0] = constants.MESSAGE_TX_SYNC
	p[1] = byte(len(content))
	p[2] = messageType
	copy(p[3:], content)

	a := AntPacket(p)
	a[len(a)-1] = a.CalculateChecksum()

	return a
}

func SystemResetMessage() AntPacket {
	return makeAntPacket(constants.MESSAGE_SYSTEM_RESET, []byte{0x00})
}

func SetNetworkKeyMessage(channel uint8, key []byte) AntPacket {
	return makeAntPacket(constants.MESSAGE_NETWORK_KEY, append([]byte{byte(channel)}, key...))
}

func OpenChannelMessage(channel uint8) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_OPEN, []byte{byte(channel)})
}

func CloseChannelMessage(channel uint8) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_CLOSE, []byte{byte(channel)})
}

func AssignChannelMessage(channel uint8, typ byte) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_ASSIGN, []byte{byte(channel), typ, 0x00})
}

func SetChannelIdMessage(channel uint8) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_ID, []byte{byte(channel), 0x00, 0x00, 0x00, 0x00})
}

func SetChannelRfFrequencyMessage(channel uint8, freq uint16) AntPacket {
	return makeAntPacket(constants.MESSAGE_CHANNEL_FREQUENCY, []byte{byte(channel), byte(freq - 2400)})
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
	var opt byte

	if rxTimestamp {
		opt |= constants.EXT_FLAG_TIMESTAMP
	}
	if rssi {
		opt |= constants.EXT_FLAG_RSSI
	}
	if channelId {
		opt |= constants.EXT_FLAG_CHANNEL_ID
	}

	return makeAntPacket(constants.MESSAGE_LIB_CONFIG, []byte{0x00, opt})
}