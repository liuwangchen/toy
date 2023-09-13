package bitarray

// BitArray 位数组
type BitArray struct {
	values []byte
	length int
}

// NewBitArray create
func NewBitArray(length int) *BitArray {
	b := &BitArray{
		values: make([]byte, (length+7)/8),
	}
	b.length = len(b.values) * 8
	return b
}

// Set set
func (b *BitArray) Set(index int, value bool) {
	if index < 0 || index >= b.length {
		return
	}
	i := index / 8
	j := uint32(index % 8)
	b.values[i] |= 1 << j
}

// Get get
func (b *BitArray) Get(index int) bool {
	if index < 0 || index >= b.length {
		return false
	}
	i := index / 8
	j := uint32(index % 8)
	return (b.values[i] & (1 << j)) > 0
}
