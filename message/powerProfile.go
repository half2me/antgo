package message

import (
	"bytes"
	"encoding/binary"
)

type PowerMessage AntBroadcastMessage

// The specifies the type of message sent by the power sensor
// Currently we only decode the standard Power-Only main data page (0x10)
func (p PowerMessage) DataPageNumber() uint8 {
	return uint8(AntBroadcastMessage(p).Content()[0])
}

// The update event count field is incremented each time the information in the message is updated.
// There are no invalid values for update event count.
func (p PowerMessage) EventCount() uint8 {
	return uint8(AntBroadcastMessage(p).Content()[1])
}

// Using the previous message we can see the difference of the event counts. Since sensors increment
// this value by 1 every time they generate an ANT+ message, we can use this value to get an idea of
// how many frames were dropped since the last message.
func (p PowerMessage) EventCountDiff(prev PowerMessage) uint8 {
	return p.EventCount() - prev.EventCount()
}

// The instantaneous cadence field is used to transmit the pedaling cadence recorded from the power sensor.
// This field is an instantaneous value only; it does not accumulate between messages. (RPM)
func (p PowerMessage) InstantaneousCadence() uint8 {
	return uint8(AntBroadcastMessage(p).Content()[3])
}

// Accumulated power is the running sum of the instantaneous power data and is incremented at each update
// of the update event count. (W)
func (p PowerMessage) AccumulatedPower() (num uint16) {
	binary.Read(bytes.NewReader(AntBroadcastMessage(p).Content()[4:6]), binary.LittleEndian, &num)
	return
}

// Using the previous message, get the calculated Power from the difference of the accumulated values.
// This gives a more precise measurement and should be used instead of the instantaneous values. (W)
func (p PowerMessage) AccumulatedPowerDiff(prev PowerMessage) (uint16) {
	return p.AccumulatedPower() - prev.AccumulatedPower()
}

// Instantaneous power measurement (W)
func (p PowerMessage) InstantaneousPower() (num uint16) {
	binary.Read(bytes.NewReader(AntBroadcastMessage(p).Content()[6:8]), binary.LittleEndian, &num)
	return
}

// Under normal conditions with complete RF reception, average power equals instantaneous power.
// In conditions where packets are lost, average power accurately calculates power over the interval
// between the received messages.
func (p PowerMessage) AveragePower(prev PowerMessage) uint16 {
	if prev == nil {
		return p.InstantaneousPower()
	}

	if prev.EventCount() == p.EventCount() {
		return p.InstantaneousPower()
	}

	return p.AccumulatedPowerDiff(prev) / uint16(p.EventCountDiff(prev))
}