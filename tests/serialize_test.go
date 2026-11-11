// SPDX-License-Identifier: MIT
// Copyright 2026 Alex Syrnikov <alex19srv@gmail.com>

package serialize

import (
	"testing"
)

func TestIntDataRoundTrip(t *testing.T) {
	original := IntData{
		ID:    0xffffabcd7382125,
		Id8:   127,
		Id16:  32535,
		Id32:  2147483647,
		Id64:  9223372036854775807,
		UID:   0xffffabcd7382125,
		UId8:  255,
		UId16: 65535,
		UId32: 4294967295,
		UId64: 18446744073709551615,
	}
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
