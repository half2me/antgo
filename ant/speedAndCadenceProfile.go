package ant

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
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

func (m SpeedAndCadenceMessage) speedEventTimeDiff(prev SpeedAndCadenceMessage) uint16 {
	return m.SpeedEventTime() - prev.SpeedEventTime()
}

func (m SpeedAndCadenceMessage) cadenceEventTimeDiff(prev SpeedAndCadenceMessage) uint16 {
	return m.CadenceEventTime() - prev.CadenceEventTime()
}

func (m SpeedAndCadenceMessage) speedRevolutionCountDiff(prev SpeedAndCadenceMessage) uint16 {
	return m.CumulativeSpeedRevolutionCount() - prev.CumulativeSpeedRevolutionCount()
}

func (m SpeedAndCadenceMessage) cadenceRevolutionCountDiff(prev SpeedAndCadenceMessage) uint16 {
	return m.CumulativeCadenceRevolutionCount() - prev.CumulativeCadenceRevolutionCount()
}

// Cadence Returns the cadence calculated from the previous message
// If the "ok" parameter is false, this indicates that a complete rotation has not yet occurred.
// In this case the cadence has not changed, but it is impossible to calculate from these two messages.
// We can use this to also handle cases where the pedal stops: "coasting" (EventTime counter does not change)
// Cadence: (RPM)
func (m SpeedAndCadenceMessage) Cadence(prev SpeedAndCadenceMessage) (cadence float64, ok bool) {
	eventCountDiff := m.cadenceEventTimeDiff(prev)
	if eventCountDiff == 0 {
		return 0, false
	}

	return float64(m.cadenceRevolutionCountDiff(prev)) * 1024 * 60 / float64(eventCountDiff), true
}

// Distance travelled since the last message: (m)
// circumference: Circumference of the wheel (m)
func (m SpeedAndCadenceMessage) Distance(prev SpeedAndCadenceMessage, circumference float64) float64 {
	return float64(m.speedRevolutionCountDiff(prev)) * circumference
}

// Speed in (m/s)
// circumference: Circumference of the wheel (m)
// If the "ok" parameter is false, this indicates that a complete rotation has not yet occurred.
// In this case the speed has not changed, but it is impossible to calculate from these two messages.
// We can use this to also handle cases where the pedal stops: "coasting" (EventTime counter does not change)
func (m SpeedAndCadenceMessage) Speed(prev SpeedAndCadenceMessage, circumference float64) (speed float64, ok bool) {
	eventCountDiff := m.speedEventTimeDiff(prev)
	if eventCountDiff == 0 {
		return 0, false
	}

	return m.Distance(prev, circumference) * 1024 / float64(eventCountDiff), true
}

// Speed2 calculates speed in (m/s)
// circumference: Circumference of the wheel (m)
// timeSincePrev
// If the "ok" parameter is false, this indicates that a complete rotation has not yet occurred.
func (m SpeedAndCadenceMessage) Speed2(prev SpeedAndCadenceMessage, timeSincePrev time.Duration, circumference float64) (speed float64, ok bool) {
	eventCountDiff := m.speedEventTimeDiff(prev)
	adjustedTimeDiff := adjustedTime(float64(eventCountDiff)/1024, timeSincePrev.Seconds())

	if eventCountDiff == 0 && adjustedTimeDiff < .01 {
		return 0, false
	}

	return m.Distance(prev, circumference) / adjustedTimeDiff, true
}

// adjustedTime adjusts the precise time by adding any number of overflows to the clock that may have happened.
// the estimated parameter should specify roughly how much time has passed since the previous packet.
// This allows us to make a good guess about how many overflows happened.
// The counter can only count up to 64sec before overflowing, so if a packet takes longer than that to arrive,
// the timer will be off by -64 sec.
// parameters should be time in seconds
func adjustedTime(precise, estimated float64) float64 {
	estimatedOverflows := int(estimated / 64)
	estimatedRemainder := math.Mod(estimated, 64)

	overflows := estimatedOverflows
	left, right := moduloDistance(precise, estimatedRemainder)

	if left < right && left > precise {
		overflows += 1
	} else if right < left && precise > left {
		overflows -= 1
	}

	return float64(overflows)*64 + precise
}

func moduloDistance(crt, next float64) (left, right float64) {
	right = math.Mod(next-crt+64, 64)
	left = math.Mod(crt-next+64, 64)
	return left, right
}
