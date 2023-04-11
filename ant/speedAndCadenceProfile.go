package ant

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type SpeedAndCadenceMessage BroadcastMessage

func (m SpeedAndCadenceMessage) String() string {
	return fmt.Sprintf("Cadence: #%5d %5drev, Speed: #%5d %5drev", m.CadenceEventTime(), m.CumulativeCadenceRevolutionCount(), m.SpeedEventTime(), m.CumulativeSpeedRevolutionCount())
}

// CadenceEventTime represents the time of the last valid bike cadence event (1/1024 sec)
func (m SpeedAndCadenceMessage) CadenceEventTime() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[0:2]), binary.LittleEndian, &num)
	return
}

// CumulativeCadenceRevolutionCount represents the total number of pedal revolutions
func (m SpeedAndCadenceMessage) CumulativeCadenceRevolutionCount() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[2:4]), binary.LittleEndian, &num)
	return
}

// SpeedEventTime represents the time of the last valid bike speed event (1/1024 sec)
func (m SpeedAndCadenceMessage) SpeedEventTime() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[4:6]), binary.LittleEndian, &num)
	return
}

// CumulativeSpeedRevolutionCount represents the total number of wheel revolutions
func (m SpeedAndCadenceMessage) CumulativeSpeedRevolutionCount() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[6:8]), binary.LittleEndian, &num)
	return
}

func (m SpeedAndCadenceMessage) SpeedEventTimeDiff(prev SpeedAndCadenceMessage) uint16 {
	if prev == nil {
		return 0
	}
	return m.SpeedEventTime() - prev.SpeedEventTime()
}

func (m SpeedAndCadenceMessage) CadenceEventTimeDiff(prev SpeedAndCadenceMessage) uint16 {
	if prev == nil {
		return 0
	}
	return m.CadenceEventTime() - prev.CadenceEventTime()
}

func (m SpeedAndCadenceMessage) speedRevolutionCountDiff(prev SpeedAndCadenceMessage) uint16 {
	return m.CumulativeSpeedRevolutionCount() - prev.CumulativeSpeedRevolutionCount()
}

func (m SpeedAndCadenceMessage) CadenceRevolutionCountDiff(prev SpeedAndCadenceMessage) uint16 {
	return m.CumulativeCadenceRevolutionCount() - prev.CumulativeCadenceRevolutionCount()
}

// Returns the cadence calculated from the previous message (0 if previous message is nil)
// The second parameter "nochange" indicates that the EventTime counter has not changed since the
// previous message, which means a complete rotation has not yet occurred. You should use this value
// to handle cases where the pedal stops: "coasting" (EventTime counter does not change)
// Cadence: (RPM)
func (m SpeedAndCadenceMessage) Cadence(prev SpeedAndCadenceMessage) (cadence float32, nochange bool) {
	if prev == nil {
		return 0, false
	}

	if m.CadenceEventTimeDiff(prev) == 0 {
		return 0, true
	}

	return float32(m.CadenceRevolutionCountDiff(prev)) * 1024 * 60 / float32(m.CadenceEventTimeDiff(prev)), false
}

// Distance travelled since the last message: (m)
// circumference: Circumference of the wheel (m)
func (m SpeedAndCadenceMessage) Distance(prev SpeedAndCadenceMessage, circumference float32) float32 {
	if prev == nil {
		return 0
	}
	return float32(m.speedRevolutionCountDiff(prev)) * circumference
}

// Speed: (m/s)
// circumference: Circumference of the wheel (m)
// The second parameter "nochange" indicates that the EventTime counter has not changed since the
// previous message, which means a complete rotation has not yet occurred. You should use this value
// to handle cases where the pedal stops: "coasting" (EventTime counter does not change)
func (m SpeedAndCadenceMessage) Speed(prev SpeedAndCadenceMessage, circumference float32) (speed float32, nochange bool) {
	if prev == nil {
		return 0, false
	}

	if m.SpeedEventTimeDiff(prev) == 0 {
		return 0, true
	}

	return m.Distance(prev, circumference) * 1024 / float32(m.SpeedEventTimeDiff(prev)), false
}
