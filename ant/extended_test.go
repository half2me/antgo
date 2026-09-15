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

func channelId() []byte {
	return []byte{0x1C, 0xBE, DEVICE_TYPE_FE, 0x05}
}

// The RSSI block's width is decided by the measurement type byte leading it, so
// the timestamp behind it does not sit at a fixed offset. A dBm dongle — the
// common case — puts it two bytes earlier than an AGC one, and a dongle sending
// no RSSI at all four bytes earlier again.
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
			name: "no extended content at all",
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
}

func TestRssiInfo(t *testing.T) {
	dbm := broadcast(
		EXT_FLAG_CHANNEL_ID|EXT_FLAG_RSSI|EXT_FLAG_TIMESTAMP,
		append(channelId(), RSSI_MEASUREMENT_TYPE_DBM, 0xBA, 0x14, 0x11, 0x22)...,
	)
	rssi, ok := dbm.RssiInfo()
	if !ok {
		t.Fatal("no rssi reported for a message carrying a dBm reading")
	}
	if got := rssi.Value(); got != -70 {
		t.Errorf("rssi = %d dBm, want -70", got)
	}

	// An AGC measurement carries a threshold offset and a register, so there is
	// no dBm reading to report.
	agc := broadcast(
		EXT_FLAG_CHANNEL_ID|EXT_FLAG_RSSI,
		append(channelId(), RSSI_MEASUREMENT_TYPE_AGC, 0xBA, 0x14, 0x03)...,
	)
	if _, ok := agc.RssiInfo(); ok {
		t.Error("reported an AGC measurement as a dBm reading")
	}

	none := broadcast(EXT_FLAG_CHANNEL_ID, channelId()...)
	if _, ok := none.RssiInfo(); ok {
		t.Error("reported rssi for a message that carries none")
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
