package ant

const (
	MESSAGE_RF           = 0x01
	MESSAGE_TX_SYNC      = 0xA4
	MESSAGE_SYSTEM_RESET = 0x4A
	MESSAGE_STARTUP      = 0x6F
	MESSAGE_NETWORK_KEY  = 0x46
	ANTPLUS_NETWORK_KEY  = "\xB9\xA5\x21\xFB\xBD\x72\xC3\x45"
	OPEN_RX_SCAN_MODE    = 0x5B

	MESSAGE_CHANNEL_EVENT     = 0x40
	MESSAGE_CHANNEL_ACK       = 0x4F
	MESSAGE_CHANNEL_ASSIGN    = 0x42
	MESSAGE_CHANNEL_ID        = 0x51
	MESSAGE_CHANNEL_FREQUENCY = 0x45
	MESSAGE_CHANNEL_OPEN      = 0x4B
	MESSAGE_CHANNEL_CLOSE     = 0x4C

	MESSAGE_ENABLE_EXT_RX_MESSAGES = 0x66
	MESSAGE_LIB_CONFIG             = 0x6E

	// Extended message flags
	EXT_FLAG_CHANNEL_ID = 0x80
	EXT_FLAG_RSSI       = 0x40
	EXT_FLAG_TIMESTAMP  = 0x20

	CHANNEL_TYPE_TWOWAY_RECEIVE  = 0x00
	CHANNEL_TYPE_TWOWAY_TRANSMIT = 0x10
	CHANNEL_TYPE_SHARED_RECEIVE  = 0x20
	CHANNEL_TYPE_SHARED_TRANSMIT = 0x30
	CHANNEL_TYPE_ONEWAY_RECEIVE  = 0x40
	CHANNEL_TYPE_ONEWAY_TRANSMIT = 0x50

	MESSAGE_TYPE_BROADCAST = 0x4E

	DEVICE_TYPE_SPEED_AND_CADENCE = 0x79
	DEVICE_TYPE_POWER             = 0x0B
	DEVICE_TYPE_FE                = 0x11
	DEVICE_TYPE_SDM               = 0x7C
)
