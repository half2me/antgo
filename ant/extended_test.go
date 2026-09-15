package ant

import (
	"testing"
)

// broadcast builds an extended broadcast message: channel, an 8-byte payload,
// the flag byte, then whatever blocks the caller passes as the extended
// content.
func broadcast(flag byte, ext ...byte) BroadcastMessage {
	data := []byte{0x00, 0x10, 0x19, 0xFF, 0x2C, 0x01, 0x64, 0x00, 0x20, flag}
	data = append(data, ext...)
	return BroadcastMessage(MakeAntPacket(MESSAGE_TYPE_BROADCAST, data))
}

// standardBroadcast is what a dongle with extended messages switched off sends:
// the channel and the payload, ending before the flag byte. Nothing may read
// one as though a flag were there.
func standardBroadcast() BroadcastMessage {
	data := []byte{0x00, 0x10, 0x19, 0xFF, 0x2C, 0x01, 0x64, 0x00, 0x20}
	return BroadcastMessage(MakeAntPacket(MESSAGE_TYPE_BROADCAST, data))
}

func channelId() []byte {
	return []byte{0x1C, 0xBE, DEVICE_TYPE_FE, 0x05}
}

// The RSSI block's width is decided by the measurement type byte leading it, so
// the timestamp behind it does not sit at a fixed offset. A dBm dongle puts it
// two bytes earlier than an AGC one, and a dongle sending no RSSI at all four
// bytes earlier again.
func TestRxTimestampFollowsTheBlocksInFrontOfIt(t *testing.T) {
	tests := []struct {
		name string
		flag byte
		ext  []byte
		want uint16
	}{
		{
			name: "channel id, dBm rssi, timestamp",
			flag: EXT_FLAG_CHANNEL_ID | EXT_FLAG_RSSI | EXT_FLAG_TIMESTAMP,
			ext:  append(channelId(), RSSI_MEASUREMENT_TYPE_DBM, 0xBA, 0x14, 0x11, 0x22),
			want: 0x2211,
		},
		{
			name: "channel id, agc rssi, timestamp",
			flag: EXT_FLAG_CHANNEL_ID | EXT_FLAG_RSSI | EXT_FLAG_TIMESTAMP,
			ext:  append(channelId(), RSSI_MEASUREMENT_TYPE_AGC, 0xBA, 0x14, 0x03, 0x11, 0x22),
			want: 0x2211,
		},
		{
			name: "channel id and timestamp, no rssi",
			flag: EXT_FLAG_CHANNEL_ID | EXT_FLAG_TIMESTAMP,
			ext:  append(channelId(), 0x11, 0x22),
			want: 0x2211,
		},
		{
			// No channel id in front, so a parser that still assumes one is
			// four bytes out. Every other rssi case here carries one.
			name: "dBm rssi and timestamp, no channel id",
			flag: EXT_FLAG_RSSI | EXT_FLAG_TIMESTAMP,
			ext:  []byte{RSSI_MEASUREMENT_TYPE_DBM, 0xBA, 0x14, 0x11, 0x22},
			want: 0x2211,
		},
		{
			name: "agc rssi and timestamp, no channel id",
			flag: EXT_FLAG_RSSI | EXT_FLAG_TIMESTAMP,
			ext:  []byte{RSSI_MEASUREMENT_TYPE_AGC, 0xBA, 0x14, 0x03, 0x11, 0x22},
			want: 0x2211,
		},
		{
			name: "timestamp alone",
			flag: EXT_FLAG_TIMESTAMP,
			ext:  []byte{0x11, 0x22},
			want: 0x2211,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, ok := broadcast(tt.flag, tt.ext...).RxTimestamp()
			if !ok {
				t.Fatal("no timestamp reported for a message carrying one")
			}
			if ts != tt.want {
				t.Errorf("timestamp = %#04X, want %#04X", ts, tt.want)
			}
		})
	}
}

// A block the flag does not announce is absent, not zero: a caller that cannot
// tell the two apart treats a dongle with no timestamps as one whose clock
// reads zero on every message.
func TestRxTimestampIsAbsentRatherThanZero(t *testing.T) {
	tests := []struct {
		name string
		flag byte
		ext  []byte
	}{
		{
			name: "not announced",
			flag: EXT_FLAG_CHANNEL_ID,
			ext:  channelId(),
		},
		{
			name: "announced but the message ends first",
			flag: EXT_FLAG_CHANNEL_ID | EXT_FLAG_TIMESTAMP,
			ext:  append(channelId(), 0x11),
		},
		{
			name: "announced behind an rssi block the message ends inside",
			flag: EXT_FLAG_CHANNEL_ID | EXT_FLAG_RSSI | EXT_FLAG_TIMESTAMP,
			ext:  append(channelId(), RSSI_MEASUREMENT_TYPE_DBM, 0xBA),
		},
		{
			// An rssi block of unknown width leaves everything behind it at an
			// unknown offset, so the two bytes there are not a timestamp.
			// Guessing a width would report them as one.
			name: "announced behind an rssi block of an unknown measurement type",
			flag: EXT_FLAG_CHANNEL_ID | EXT_FLAG_RSSI | EXT_FLAG_TIMESTAMP,
			ext:  append(channelId(), 0x40, 0xBA, 0x14, 0x11, 0x22),
		},
		{
			name: "a flag byte announcing nothing",
			flag: 0x00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ts, ok := broadcast(tt.flag, tt.ext...).RxTimestamp(); ok {
				t.Fatalf("reported a timestamp of %#04X that the message does not carry", ts)
			}
		})
	}

	t.Run("no flag byte at all", func(t *testing.T) {
		msg := standardBroadcast()
		if ts, ok := msg.RxTimestamp(); ok {
			t.Errorf("reported a timestamp of %#04X for a standard broadcast", ts)
		}
		if _, ok := msg.RssiInfo(); ok {
			t.Error("reported rssi for a standard broadcast")
		}
	})
}

func TestRssiInfoDbm(t *testing.T) {
	dbm := broadcast(
		EXT_FLAG_CHANNEL_ID|EXT_FLAG_RSSI|EXT_FLAG_TIMESTAMP,
		append(channelId(), RSSI_MEASUREMENT_TYPE_DBM, 0xBA, 0x14, 0x11, 0x22)...,
	)
	rssi, ok := dbm.RssiInfo()
	if !ok {
		t.Fatal("no rssi reported for a message carrying a dBm reading")
	}
	value, threshold, ok := rssi.Dbm()
	if !ok {
		t.Fatal("a dBm block did not report a dBm reading")
	}
	if value != -70 || threshold != 20 {
		t.Errorf("rssi = %d dBm, threshold %d, want -70 and 20", value, threshold)
	}
	// The gain state of a receiver that reports strength directly is not a
	// thing to hand back, however convenient the bytes would be.
	if _, _, ok := rssi.Agc(); ok {
		t.Error("a dBm block reported an AGC reading")
	}

	// Without a channel id block the reading sits where the old fixed offsets
	// never looked, which is the other half of the same regression.
	bare := broadcast(EXT_FLAG_RSSI, RSSI_MEASUREMENT_TYPE_DBM, 0xBA, 0x14)
	bareRssi, ok := bare.RssiInfo()
	if !ok {
		t.Fatal("no rssi reported for a message whose rssi block leads")
	}
	if value, _, _ := bareRssi.Dbm(); value != -70 {
		t.Errorf("rssi = %d dBm, want -70", value)
	}
}

// Both genuine Dynastream sticks on the bench reported AGC, so this is the
// reading real hardware actually produces; the bytes here are theirs.
func TestRssiInfoAgc(t *testing.T) {
	agc := broadcast(
		EXT_FLAG_CHANNEL_ID|EXT_FLAG_RSSI|EXT_FLAG_TIMESTAMP,
		append(channelId(), RSSI_MEASUREMENT_TYPE_AGC, 0x00, 0x68, 0x00, 0x1B, 0x2A)...,
	)
	rssi, ok := agc.RssiInfo()
	if !ok {
		t.Fatal("no rssi reported for a message carrying an AGC measurement")
	}
	offset, register, ok := rssi.Agc()
	if !ok {
		t.Fatal("an AGC block did not report an AGC reading")
	}
	if offset != 0 || register != 104 {
		t.Errorf("agc offset = %d, register = %d, want 0 and 104", offset, register)
	}
	// An AGC block's bytes are a threshold offset and a register, so reading
	// them as a signal strength would produce a plausible number that is not one.
	if _, _, ok := rssi.Dbm(); ok {
		t.Error("an AGC block reported a dBm reading")
	}

	// The block behind it is still located correctly, which is what the width
	// is for.
	if ts, ok := agc.RxTimestamp(); !ok || ts != 0x2A1B {
		t.Errorf("timestamp = %#04X (ok=%v), want 0x2A1B", ts, ok)
	}
}

func TestRssiInfoAbsent(t *testing.T) {
	// A measurement type we do not know has an unknown width, so there is
	// nothing to report and nothing behind it can be located either.
	unknown := broadcast(EXT_FLAG_CHANNEL_ID|EXT_FLAG_RSSI, append(channelId(), 0x40, 0xBA, 0x14)...)
	if _, ok := unknown.RssiInfo(); ok {
		t.Error("reported a reading from an rssi block of an unknown measurement type")
	}

	none := broadcast(EXT_FLAG_CHANNEL_ID, channelId()...)
	if _, ok := none.RssiInfo(); ok {
		t.Error("reported rssi for a message that carries none")
	}

	// The zero value answers rather than panicking, since that is what a caller
	// who ignored the second return gets.
	var zero Rssi
	if _, _, ok := zero.Dbm(); ok {
		t.Error("the zero value reported a dBm reading")
	}
	if _, _, ok := zero.Agc(); ok {
		t.Error("the zero value reported an AGC reading")
	}
}

// The channel id sits in front of the optional blocks, so the fields the rest
// of the library routes on must not move when they are present.
func TestChannelIdSurvivesTheBlocksBehindIt(t *testing.T) {
	bare := broadcast(EXT_FLAG_CHANNEL_ID, channelId()...)
	full := broadcast(
		EXT_FLAG_CHANNEL_ID|EXT_FLAG_RSSI|EXT_FLAG_TIMESTAMP,
		append(channelId(), RSSI_MEASUREMENT_TYPE_DBM, 0xBA, 0x14, 0x11, 0x22)...,
	)

	if bare.DeviceNumber() != full.DeviceNumber() {
		t.Errorf("device number %d with the extended blocks, %d without", full.DeviceNumber(), bare.DeviceNumber())
	}
	if full.DeviceType() != DEVICE_TYPE_FE {
		t.Errorf("device type = %#02X, want %#02X", full.DeviceType(), DEVICE_TYPE_FE)
	}
}
