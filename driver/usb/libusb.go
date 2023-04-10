package usb

import (
	"bytes"
	"io"
)

type SkipLibUsbLog struct {
	out io.Writer
}

func FixLibUsbLog(out io.Writer) SkipLibUsbLog {
	return SkipLibUsbLog{out: out}
}

var interruptedError = []byte("libusb: interrupted [code -10]")

func (l SkipLibUsbLog) Write(data []byte) (int, error) {
	if bytes.Contains(data, interruptedError) {
		return len(data), nil
	}
	return l.out.Write(data)
}
