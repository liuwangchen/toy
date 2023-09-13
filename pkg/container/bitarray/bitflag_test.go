package bitarray

import (
	"testing"
)

func TestBitFlag(t *testing.T) {
	var a uint32 = 0
	for i := 0; i < maxIndex; i++ {
		if UInt32GetFlag(a, i) {
			t.Error("error")
		}
	}
	for i := 0; i < maxIndex; i++ {
		UInt32SetFlag(&a, i, true)
		if !UInt32GetFlag(a, i) {
			t.Error("error")
		}
	}

	for i := maxIndex - 1; i >= 0; i-- {
		UInt32SetFlag(&a, i, false)
		if UInt32GetFlag(a, i) {
			t.Error("error")
		}
	}

}
