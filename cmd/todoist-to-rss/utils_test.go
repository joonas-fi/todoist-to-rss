package main

import (
	"math"
	"testing"

	"github.com/function61/gokit/testing/assert"
)

func TestIntToGuid(t *testing.T) {
	assert.Equal(t, intToGuid(0), "00000000-0000-0000-8b19-71551aca4c54")
	assert.Equal(t, intToGuid(128), "80000000-0000-0000-8b19-71551aca4c54")
	assert.Equal(t, intToGuid(0b11111010<<32), "00000000-fa00-0000-8b19-71551aca4c54")
	assert.Equal(t, intToGuid(math.MaxInt64), "ffffffff-ffff-ff7f-8b19-71551aca4c54")
}
