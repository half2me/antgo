package main

import (
	"testing"
	"github.com/half2me/antgo/message"
	"time"
)

func TestDeDup(t *testing.T) {
	dup := Dup {
		data: make(map[string]struct{}),
		timeout: 1,
	}

	m := message.AntPacket{}

	if dup.Test(m) != true {
		t.Error("Dedup falsely reporting packet as duplicate!")
	}

	if dup.Test(m) == true {
		t.Error("Dedup missed a duplicate!")
	}

	time.Sleep(time.Second * 2)

	if dup.Test(m) != true {
		t.Error("Dedup cache stuck!")
	}
}
