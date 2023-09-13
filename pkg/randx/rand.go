package randx

import (
	"math/rand"
)

// RandBetween 在 [min, max] 间随机一个数
func RandBetween(min, max int32) int32 {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + rand.Int31n(max-min+1)
}
