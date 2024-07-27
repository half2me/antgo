package ant

import (
	"math"
	"testing"
)

func TestAdjustedTime(t *testing.T) {
	var got, want float64

	// Single overflow, high estimate, low timer
	got = adjustedTime(3, 70)
	want = float64(64 + 3)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(3, 70) got: %v, want: %v", got, want)
	}

	// Single overflow, low estimate, low timer
	got = adjustedTime(3, 50)
	want = float64(64 + 3)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(3, 50) got: %v, want: %v", got, want)
	}

	// Single overflow, high estimate, high timer
	got = adjustedTime(63, 140)
	want = float64(64 + 63)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(63, 140) got: %v, want: %v", got, want)
	}

	// Single overflow, low estimate, high timer
	got = adjustedTime(63, 120)
	want = float64(64 + 63)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(63, 120) got: %v, want: %v", got, want)
	}

	// No overflow, low estimate, low timer
	got = adjustedTime(10, 5)
	want = float64(10)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(10, 5) got: %v, want: %v", got, want)
	}

	// No overflow, high estimate, low timer
	got = adjustedTime(10, 20)
	want = float64(10)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(10, 20) got: %v, want: %v", got, want)
	}

	// No overflow, low estimate, high timer
	got = adjustedTime(63, 50)
	want = float64(63)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(63, 50) got: %v, want: %v", got, want)
	}

	// No overflow, high estimate, high timer
	got = adjustedTime(63, 70)
	want = float64(63)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(63, 70) got: %v, want: %v", got, want)
	}

	// more for higher overflows
	got = adjustedTime(5, 5*64+62)
	want = float64(6*64 + 5)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(5, 5*64+62) got: %v, want: %v", got, want)
	}

	// and more for higher overflows
	got = adjustedTime(5, 5*64+1)
	want = float64(5*64 + 5)
	if !almostEqual(got, want) {
		t.Fatalf("adjustedTime(5, 5*64+1) got: %v, want: %v", got, want)
	}
}

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}
