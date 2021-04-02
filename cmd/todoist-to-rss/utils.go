package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
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

func multiCompare(results ...int) bool {
	for _, result := range results {
		if result != 0 { // not equal
			return result < 0
		}

		// continuing only if first comparison result equal
	}

	return false
}

func intCompare(a, b int) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}

func int64Compare(a, b int64) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}

// only date component in format: 2006-01-02
type JSONPlainDate struct {
	time.Time
}

var _ json.Unmarshaler = (*JSONPlainDate)(nil)

func (b *JSONPlainDate) UnmarshalJSON(input []byte) error {
	parsed, err := time.Parse(`"2006-01-02"`, string(input))
	if err != nil {
		return fmt.Errorf("JSONPlainDate: %w", err)
	}

	*b = JSONPlainDate{parsed}

	return nil
}
