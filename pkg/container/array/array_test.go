package array

import (
	"testing"
)

func TestRandInt32Slice(t *testing.T) {
	RandInt32Slice(nil, 3)
	var a = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if RandInt32Slice(a, 100) != nil {
		t.Error("RandInt32Slice")
	}
	if RandInt32Slice(a, 0) != nil {
		t.Error("RandInt32Slice")
	}

	if RandInt32Slice(a, -1) != nil {
		t.Error("RandInt32Slice")
	}

	if r := RandInt32Slice(a, len(a)); len(r) != len(a) {
		t.Error("RandInt32Slice")
	}
}

func BenchmarkRandInt32Slice(b *testing.B) {
	var a = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < b.N; i++ {
		RandInt32Slice(a, 3)
	}
}
