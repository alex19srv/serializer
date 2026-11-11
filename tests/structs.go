// SPDX-License-Identifier: MIT
// Copyright 2026 Alex Syrnikov <alex19srv@gmail.com>

package serialize

//	type User struct { // persist_it!!!
//		ID   int
//		Name string
//	}
//
// saga:serialize
type IntData struct {
	ID    int
	Id8   int8
	Id16  int16
	Id32  int32
	Id64  int64
	UID   uint
	UId8  uint8
	UId16 uint16
	UId32 uint32
	UId64 uint64
}

// saga:serialize
type Data struct {
	ID    int
	Value string
	SID   int
	// User  User
}
