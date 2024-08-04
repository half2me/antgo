package ant

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type FeMessage BroadcastMessage

func (m FeMessage) String() (s string) {
	if m.DataPageNumber() == 0x10 {
		s = fmt.Sprintf("FE (%s) #%5dm %5dm/s", m.EquipmentTypeString(), m.AccumulatedDistance(), m.InstantaneousSpeed())
	} else {
		s = "FE (service packet)"
	}
	return
}

// DataPageNumber specifies the type of message sent by the sensor
// Currently we only decode the main data page (0x10)
func (m FeMessage) DataPageNumber() uint8 {
	return BroadcastMessage(m).Content()[0]
}

func (m FeMessage) EquipmentType() uint8 {
	return BroadcastMessage(m).Content()[1]
}

func (m FeMessage) EquipmentTypeString() string {
	switch m.EquipmentType() {
	case 19:
		return "Treadmill"
	case 20:
		return "Elliptical"
	case 22:
		return "Rower"
	case 23:
		return "Climber"
	case 24:
		return "Nordic Skier"
	case 25:
		return "Trainer/Stationary Bike"
	default:
		return "Unknown"
	}
}

// ElapsedTime (1/4sec) Accumulated value of the elapsed time since start of workout
// rollover every 64s
func (m FeMessage) ElapsedTime() uint8 {
	return BroadcastMessage(m).Content()[2]
}

// AccumulatedDistance (m) Accumulated value of the distance traveled since start of workout
// rollover every 256m
func (m FeMessage) AccumulatedDistance() uint8 {
	return BroadcastMessage(m).Content()[3]
}

// InstantaneousSpeed (m/s) Instantaneous speed
func (m FeMessage) InstantaneousSpeed() (num uint16) {
	_ = binary.Read(bytes.NewReader(BroadcastMessage(m).Content()[4:6]), binary.LittleEndian, &num)
	if num == 0xFFFF {
		// 0xFFFF means "invalid value"
		return 0
	}
	return
}

func (m FeMessage) ellapsedTimeDiff(prev FeMessage) uint8 {
	return m.ElapsedTime() - prev.ElapsedTime()
}

// Distance travelled since the last message: (m)
func (m FeMessage) Distance(prev FeMessage) uint8 {
	return m.AccumulatedDistance() - prev.AccumulatedDistance()
}

// Speed in (m/s)
// If the "ok" parameter is false, this indicates that a complete rotation has not yet occurred.
// In this case the speed has not changed, but it is impossible to calculate from these two messages.
// We can use this to also handle cases where the pedal stops: "coasting" (EventTime counter does not change)
func (m FeMessage) Speed(prev FeMessage) (speed float64, ok bool) {
	elapsedTimeDiff := m.ellapsedTimeDiff(prev)
	if elapsedTimeDiff == 0 {
		return 0, false
	}

	return float64(m.Distance(prev)) / (float64(elapsedTimeDiff) / 4), true
}
