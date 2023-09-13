// Package stringx 字符串扩展库
package stringx

import (
	"fmt"
	"strconv"
	"strings"
)

func StringToUint32Slice(s string, seq string) (ret []uint32) {
	if len(s) == 0 {
		return
	}
	set := strings.Split(s, seq)
	ret = make([]uint32, len(set))
	for index, value := range set {
		if len(value) != 0 {
			tmp, _ := strconv.ParseUint(value, 10, 32)
			ret[index] = uint32(tmp)
		}
	}
	return
}

func StringToUint64Slice(s string, seq string) (ret []uint64) {
	if len(s) == 0 {
		return
	}
	set := strings.Split(s, seq)
	ret = make([]uint64, len(set))
	for index, value := range set {
		if len(value) != 0 {
			tmp, _ := strconv.ParseUint(value, 10, 64)
			ret[index] = uint64(tmp)
		}
	}
	return
}

func Uint32SliceToString(set []uint32, seq string) (ret string) {
	set_len := len(set)
	if set_len == 0 {
		return
	}
	for index, value := range set {
		ret += strconv.FormatUint(uint64(value), 10)
		if index < set_len-1 {
			ret += seq
		}
	}
	return
}

func Uint64SliceToString(set []uint64, seq string) (ret string) {
	set_len := len(set)
	if set_len == 0 {
		return
	}
	for index, value := range set {
		ret += strconv.FormatUint(value, 10)
		if index < set_len-1 {
			ret += seq
		}
	}
	return
}

func StringToMap(str, seq1, seq2 string) map[string]string {
	ss := strings.Split(str, seq1)
	if len(ss) != 0 {
		m := make(map[string]string)
		for _, pair := range ss {
			v := strings.Split(pair, seq2)
			if len(v) == 2 {
				m[v[0]] = v[1]
			}
		}
		if len(m) != 0 {
			return m
		}
	}
	return nil
}

func StringToUint32Map(str, seq1, seq2 string) map[uint32]uint32 {
	ss := strings.Split(str, seq1)
	if len(ss) != 0 {
		m := make(map[uint32]uint32, len(ss))
		for _, pair := range ss {
			v := strings.Split(pair, seq2)
			if len(v) == 2 {
				key, _ := strconv.ParseUint(v[0], 10, 32)
				value, _ := strconv.ParseUint(v[1], 10, 32)
				m[uint32(key)] = uint32(value)
			}
		}
		if len(m) != 0 {
			return m
		}
	}
	return nil
}

func TwoU32ToU64(front, back uint32) string {
	return strconv.FormatUint((uint64(front)<<32)|uint64(back), 10)
}

//SplitStr 空字符串返回空数组（strings.Split空数组会返回[""]，不符合预期，因此重写）
//example 看_test
func SplitStr(s string, sep string) []string {
	if s == "" {
		return make([]string, 0)
	} else {
		return strings.Split(s, sep)
	}

	// n := strings.Count(s, sep) + 1
	// a := make([]string, n)
	// n--
	// i := 0
	// for i < n {
	// 	m := strings.Index(s, sep)
	// 	if m < 0 {
	// 		break
	// 	}
	// 	a[i] = s[:m+0]
	// 	s = s[m+len(sep):]
	// 	i++
	// }
	// a[i] = s
	// return a[:i+1]
}

//CombineUint64ToString 把a数组合并成一个以dim分割的字符串(例如:100,200,300)
func CombineUint64ToString(dim string, a ...uint64) string {
	ret := ""
	for k, v := range a {
		if k != len(a)-1 {
			ret += fmt.Sprintf("%d%s", v, dim)
		} else {
			ret += fmt.Sprintf("%d", v)
		}

	}

	return ret
}
