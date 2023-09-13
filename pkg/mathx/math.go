package mathx

import "math"

const (
	EPSILON32 float32 = 0.00000001
	EPSILON64 float64 = 0.00000001
)

// Float32Equals 比较float32等于
func Float32Equals(a, b float32) bool {
	if (a-b) < EPSILON32 && (b-a) < EPSILON32 {
		return true
	}
	return false
}

// Float64Equals 比较float64等于
func Float64Equals(a, b float64) bool {
	if (a-b) < EPSILON64 && (b-a) < EPSILON64 {
		return true
	}
	return false
}

func SafeSubUint32(x, y uint32) uint32 {
	if x > y {
		return x - y
	} else {
		return 0
	}
}

func If(condition bool, trueVal, falseVal uint32) uint32 {
	if condition {
		return trueVal
	}
	return falseVal
}

func If64(condition bool, trueVal, falseVal uint64) uint64 {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfInt64(condition bool, trueVal, falseVal int64) int64 {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfInt32(condition bool, trueVal, falseVal int32) int32 {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfString(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfInt(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// IsMultyOverFlowUint32 x*y是否溢出
func IsMultyOverFlowUint32(x, y uint32) bool {
	if x == 0 || y == 0 {
		return false
	}
	maxValue := uint32(math.MaxUint32)
	t := maxValue / x // x*t <= maxValue < x*(t+1)
	return y > t
}

// Uint64ToUint32 uint64->uint32
func Uint64ToUint32(n uint64) uint32 {
	if n >= math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(n)
}
