package main

import (
	"encoding/binary"
	"fmt"
)

// looks like c2136c55-1b7c-4ba1-8b19-71551aca4c54
func intToGuid(input int64) string {
	guid := [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // <- will be replaced with input
		0x8b, 0x19, 0x71, 0x55, 0x1a, 0xca, 0x4c, 0x54, // <- randomly generated (but now static) suffix
	}

	consumed := 0
	take := func(length int) []byte { // helper
		offset := consumed
		consumed += length // advance
		return guid[offset : offset+length]
	}

	binary.LittleEndian.PutUint64(guid[:], uint64(input))

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		take(4),
		take(2),
		take(2),
		take(2),
		take(6),
	)
}
