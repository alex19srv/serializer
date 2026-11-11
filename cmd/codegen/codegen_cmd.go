// SPDX-License-Identifier: MIT
// Copyright 2026 Alex Syrnikov <alex19srv@gmail.com>

package main

import (
	"fmt"
	"os"
	"persistent_value/internal/serializer"
)

func main() {
	if len(os.Args) != 3 {
		fail("usage: codegen <input.go> <output.go>")
	}

	if err := serializer.GenerateSerializationsForFile(os.Args[1], os.Args[2]); err != nil {
		fail("%v", err)
	}
}

func fail(formatString string, args ...any) {
	fmt.Fprintf(os.Stderr, formatString+"\n", args...)
	os.Exit(1)
}
