// SPDX-License-Identifier: MIT
// Copyright 2026 Alex Syrnikov <alex19srv@gmail.com>

package serialize

import (
	"testing"
)

func TestIntDataRoundTrip(t *testing.T) {
	original := IntData{ID: 0xffffabcd7382125}
	buffer := make([]byte, original.SerializedSize())
	size := original.Serialize(buffer)

	got := NewIntDataFromBytes(buffer[:size])
	if got != original {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, original)
	}
}

func TestDataRoundTrip(t *testing.T) {
	original := Data{
		ID:    0xffffabcd7382125,
		Value: "persistent value",
		SID:   0xffff00007777fff,
	}
	buffer := make([]byte, original.SerializedSize())
	size := original.Serialize(buffer)

	got := NewDataFromBytes(buffer[:size])
	if got != original {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, original)
	}
}
func TestDataRoundTripMisaligned(t *testing.T) {
	// misaligned buffer for int boundary
	original := Data{
		ID:    0xffffabcd7382125,
		Value: "persistent value",
		SID:   0xffff00007777fff,
	}
	bufferWhal := make([]byte, 1+original.SerializedSize())
	buffer := bufferWhal[1:]
	size := original.Serialize(buffer)

	got := NewDataFromBytes(buffer[:size])
	if got != original {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, original)
	}
}
