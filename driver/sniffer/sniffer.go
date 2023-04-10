package sniffer

import (
	"fmt"
	"github.com/half2me/antgo/driver"
	"log"
	"strings"
)

type Sniffer struct {
	driver driver.Driver
}

func Sniff(driver driver.Driver) Sniffer {
	return Sniffer{driver: driver}
}

func (s Sniffer) Close() {
	s.driver.Close()
}

func (s Sniffer) Write(buf []byte) (n int, err error) {
	var r []string
	for _, b := range buf {
		r = append(r, fmt.Sprintf("%02X", b))
	}
	log.Printf("-> %s", strings.Join(r, " "))

	return s.driver.Write(buf)
}

func (s Sniffer) Read(buf []byte) (n int, err error) {
	n, err = s.driver.Read(buf)
	var r []string
	for i := 0; i < n; i++ {
		r = append(r, fmt.Sprintf("%02X", buf[i]))
	}

	log.Printf("<- %s", strings.Join(r, " "))
	return
}

func (s Sniffer) BufferSize() int {
	return s.driver.BufferSize()
}
