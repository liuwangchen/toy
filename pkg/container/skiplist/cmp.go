package skiplist

type CmpValue struct {
}

func (t *CmpValue) CmpScore(v1 interface{}, v2 interface{}) int {
	s1 := v1.(Valuer).Score()
	s2 := v2.(Valuer).Score()
	switch {
	case s1 < s2:
		return -1
	case s1 == s2:
		return 0
	default:
		return 1
	}
}

func (t *CmpValue) CmpKey(v1 interface{}, v2 interface{}) int {
	s1 := v1.(Valuer).Key()
	s2 := v2.(Valuer).Key()
	switch {
	case s1 < s2:
		return -1
	case s1 == s2:
		return 0
	default:
		return 1
	}
}
