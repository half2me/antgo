package ant

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type PowerMessage BroadcastMessage

func (m PowerMessage) String() (s string) {
	if m.DataPageNumber() == 0x10 {
		s = fmt.Sprintf("Power #%5d %5drpm, %5dW", m.EventCount(), m.InstantaneousCadence(), m.InstantaneousPower())
	} else {
		s = "Power (service packet)"
	}

	return
}

// DataPageNumber specifies the type of message sent by the power sensor
// Currently we only decode the standard Power-Only main data page (0x10)
func (m PowerMessage) DataPageNumber() uint8 {
	return BroadcastMessage(m).Content()[0]
}

// EventCount incremented each time the information in the message is updated.
// There are no invalid values for update event count.
func (m PowerMessage) EventCount() uint8 {
	return BroadcastMessage(m).Content()[1]
}

// Using the previous message we can see the difference of the event counts. Since sensors increment
// this value by 1 every time they generate an ANT+ message, we can use this value to get an idea of
// how many frames were dropped since the last message.
func (m PowerMessage) eventCountDiff(prev PowerMessage) uint8 {
	return m.EventCount() - prev.EventCount()
}

// InstantaneousCadence (RPM) pedaling cadence recorded from the power sensor.
// This field is an instantaneous value only; it does not accumulate between messages.
func (m PowerMessage) InstantaneousCadence() uint8 {
	return BroadcastMessage(m).Content()[3]
}

// AccumulatedPower (W) running sum of the instantaneous power data, incremented at each update
// of the update event count.
func (m PowerMessage) AccumulatedPower() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[4:6]), binary.LittleEndian, &num)
	return
}

// Using the previous message, get the calculated Power from the difference of the accumulated values.
// This gives a more precise measurement and should be used instead of the instantaneous values. (W)
func (m PowerMessage) accumulatedPowerDiff(prev PowerMessage) uint16 {
	return m.AccumulatedPower() - prev.AccumulatedPower()
}

// InstantaneousPower (W) Instantaneous power
func (m PowerMessage) InstantaneousPower() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[6:8]), binary.LittleEndian, &num)
	return
}

// AveragePower (W) Under normal conditions with complete RF reception, average power equals instantaneous power.
// In conditions where packets are lost, average power accurately calculates power over the interval
// between the received messages.
func (m PowerMessage) AveragePower(prev PowerMessage) float64 {
	eventCountDiff := m.eventCountDiff(prev)

	if eventCountDiff == 0 {
		return float64(m.InstantaneousPower())
	}

	return float64(m.accumulatedPowerDiff(prev)) / float64(eventCountDiff)
}
