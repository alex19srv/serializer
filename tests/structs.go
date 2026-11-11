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
	ID int
}

// saga:serialize
type Data struct {
	ID    int
	Value string
	SID   int
	// User  User
}
