package bitarray

const maxIndex = 32

// UInt32GetFlag 获得该位置是否是1
func UInt32GetFlag(b uint32, index int) bool {
	if index < 0 || index >= maxIndex {
		panic("index should between [0,32)")
	}
	return b&(1<<uint8(index)) > 0
}

// UInt32SetFlag 设置标记位
func UInt32SetFlag(b *uint32, index int, flag bool) {
	if index < 0 || index >= maxIndex {
		panic("index should between [0,32)")
	}
	if flag {
		*b |= 1 << uint8(index)
	} else {
		*b &= ^(1 << uint8(index))
	}
}
