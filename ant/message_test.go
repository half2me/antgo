package ant

import (
	"testing"
)

func TestFoo(t *testing.T) {
	msg := BroadcastMessage("\244\016N\000\023\377\377\377\000\333c3\200\327\344\021\365\324")
	got := msg.DeviceNumber()
	want := uint32(1041623)
	if got != want {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}

func TestFoo2(t *testing.T) {

	msg := BroadcastMessage("\244\016N\000\020\360\251>|\005\344\000\200\034\276\v\005\"")
	got := msg.DeviceNumber()
	want := uint32(48668)
	if got != want {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
