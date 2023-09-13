package mathx

import (
	"math"
	"testing"
)

func TestIsMulUint32OverFlow(t *testing.T) {
	maxValue := uint32(math.MaxUint32)
	if IsMultyOverFlowUint32(1, maxValue) == true {
		t.Error("TestIsMulUint32OverFlow 1")
	}

	if IsMultyOverFlowUint32(2, maxValue) == false {
		t.Error("TestIsMulUint32OverFlow 2")
	}

	if IsMultyOverFlowUint32(0, maxValue) == true {
		t.Error("TestIsMulUint32OverFlow")
	}

	if IsMultyOverFlowUint32(maxValue, 1) == true {
		t.Error("TestIsMulUint32OverFlow")
	}
}

func Benchmark_IsMulOverFlowUint32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsMultyOverFlowUint32(435345, 334527)
	}
}

func Test_Uint64ToUint32(t *testing.T) {
	if Uint64ToUint32(math.MaxUint32-1) != math.MaxUint32-1 {
		t.Error()
	}
	if Uint64ToUint32(math.MaxUint32) != math.MaxUint32 {
		t.Error()
	}
	if Uint64ToUint32(math.MaxUint32+1) != math.MaxUint32 {
		t.Error()
	}
	if Uint64ToUint32(math.MaxUint64) != math.MaxUint32 {
		t.Error()
	}
}
