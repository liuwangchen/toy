package array

import (
	"math/rand"
)

// 数组去重，时间O(n)，空间O(n)
func UniqueSlice(items *[]int) {
	exists := make(map[int]bool, len(*items))
	j := 0
	for i, x := range *items {
		if _, e := exists[x]; !e {
			exists[x] = true
			(*items)[j] = (*items)[i]
			j++
		}
	}
	*items = (*items)[:j]
}

func UniqueUint32Slice(items *[]uint32) {
	exists := make(map[uint32]bool, len(*items))
	j := 0
	for i, x := range *items {
		if _, e := exists[x]; !e {
			exists[x] = true
			(*items)[j] = (*items)[i]
			j++
		}
	}
	*items = (*items)[:j]
}

// UniqueInt32Slice 数组去重
func UniqueInt32Slice(items *[]int32) {
	if items == nil {
		return
	}

	exists := make(map[int32]bool, len(*items))
	j := 0
	for i, x := range *items {
		if _, e := exists[x]; !e {
			exists[x] = true
			(*items)[j] = (*items)[i]
			j++
		}
	}
	*items = (*items)[:j]
}

// IsInIntSlice item 是否在items里
func IsInIntSlice(items []int, item int) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

// IsInInt32Slice item 是否在items里
func IsInInt32Slice(items []int32, item int32) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

// IsInUint32Slice item 是否在items里
func IsInUint32Slice(items []uint32, item uint32) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

// IsInUInt64Slice item 是否在items里
func IsInUInt64Slice(items []uint64, item uint64) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

// IsInStringSlice item 是否在items里
func IsInStringSlice(items []string, item string) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

// RandInt32Slice items中随机n个不同元素
func RandInt32Slice(items []int32, n int) []int32 {
	if n <= 0 || n > len(items) {
		return nil
	}

	var cloneItem = make([]int32, len(items))
	copy(cloneItem, items)
	for i := 0; i < n; i++ {
		num := len(cloneItem) - i // n最大len(cloneItem),i < n,所以num>0
		index := rand.Intn(num)
		cloneItem[index], cloneItem[num-1] = cloneItem[num-1], cloneItem[index]
	}

	return cloneItem[len(cloneItem)-n:]
}

// id分组
func SplitArrUint64(arr []uint64, num int64) [][]uint64 {
	if num == 0 {
		return nil
	}
	max := int64(len(arr))
	if max < num {
		return nil
	}
	var segment = make([][]uint64, 0)
	quantity := max / num
	end := int64(0)
	for i := int64(1); i <= num; i++ {
		qu := i * quantity
		if i != num {
			segment = append(segment, arr[i-1+end:qu])
		} else {
			segment = append(segment, arr[i-1+end:])
		}
		end = qu - i
	}
	return segment
}

// id分组，每组加上数量限制
func SplitArrUint64WithLimit(arr []uint64, count int, limit int) [][]uint64 {
	if count == 0 {
		return nil
	}
	c := int64(count)
	max := int64(len(arr))
	if max < c {
		return nil
	}
	var segment = make([][]uint64, 0)
	quantity := max / c
	if limit > 0 {
		quantity = int64(limit)
	}
	end := int64(0)
	for i := int64(1); i <= c; i++ {
		qu := i * quantity
		if i != c {
			segment = append(segment, arr[i-1+end:qu])
		} else {
			segment = append(segment, arr[i-1+end:])
		}
		end = qu - i
	}
	return segment
}

func SplitArrStr(arr []string, num int64) [][]string {
	if num == 0 {
		return nil
	}
	max := int64(len(arr))
	if max < num {
		return nil
	}
	var segment = make([][]string, 0)
	quantity := max / num
	end := int64(0)
	for i := int64(1); i <= num; i++ {
		qu := i * quantity
		if i != num {
			segment = append(segment, arr[i-1+end:qu])
		} else {
			segment = append(segment, arr[i-1+end:])
		}
		end = qu - i
	}
	return segment
}

func SplitArrStrWithLimit(arr []string, count int, limit int) [][]string {
	if count == 0 {
		return nil
	}
	c := int64(count)
	max := int64(len(arr))
	if max < c {
		return nil
	}
	var segment = make([][]string, 0)
	quantity := max / c
	if limit > 0 {
		quantity = int64(limit)
	}
	end := int64(0)
	for i := int64(1); i <= c; i++ {
		qu := i * quantity
		if i != c {
			segment = append(segment, arr[i-1+end:qu])
		} else {
			segment = append(segment, arr[i-1+end:])
		}
		end = qu - i
	}
	return segment
}

// 生成[start,end)中count个不重复的数
func GenerateRandomNumber(start int, end int, count int) []int {
	// 范围检查
	if end < start || (end-start) < count {
		return nil
	}

	// 存放结果的slice
	nums := make([]int, 0)
	for len(nums) < count {
		// 生成随机数
		num := rand.Intn(end-start) + start

		// 查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

// 生成[start,end)中count个不重复的数
func GenerateRandomUint32Number(start int, end int, count int) []uint32 {
	// 范围检查
	if count <= 0 || end < start || (end-start) < count {
		return nil
	}

	// 存放结果的slice
	nums := make([]uint32, 0, count)
	for len(nums) < count {
		// 生成随机数
		num := uint32(rand.Intn(end-start) + start)

		// 查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

// 两个数组求并集
func ArrayUnion(arr1 []int32, arr2 []int32) []int32 {
	if len(arr1) == 0 {
		return arr2
	}
	if len(arr2) == 0 {
		return arr1
	}
	m := make(map[int32]struct{}, len(arr1))
	res := make([]int32, 0, len(arr1)+len(arr2))
	for _, v := range arr1 {
		m[v] = struct{}{}
		res = append(res, v)
	}
	for _, v := range arr2 {
		if _, ok := m[v]; !ok {
			res = append(res, v)
		}
	}
	return res
}

type RandArray []uint64

func (r *RandArray) Rand(remove bool) (uint64, bool) {
	if r.Len() == 0 {
		return 0, false
	}
	var index int
	if r.Len() == 1 {
		index = 0
	} else {
		index = rand.Intn(r.Len())
	}

	p := (*r)[index]
	if remove {
		*r = append((*r)[:index], (*r)[index+1:]...)
	}
	return p, true
}

func (r RandArray) Len() int {
	return len(r)
}

func (r *RandArray) Pop() (uint64, bool) {
	if r.Len() == 0 {
		return 0, false
	}
	t := (*r)[0]
	copy(*r, (*r)[1:])
	*r = (*r)[:r.Len()-1]
	return t, true
}

// 数组往右移动n个元素，末尾的元素前移，返回移动后的array，不是原地操作
func ForwardArray(count int, src []uint64) []uint64 {
	if len(src) == 0 {
		return nil
	}
	if count > len(src) {
		return nil
	}
	tt := make([]uint64, len(src))
	copy(tt, src)

	return append(tt[len(tt)-count:], tt[:len(tt)-count]...)
}

// 两个数组求交集
func ArrayStringInter(arr1 []string, arr2 []string) []string {
	if len(arr1) == 0 || len(arr2) == 0 {
		return nil
	}

	m := make(map[string]struct{}, len(arr1))
	for _, v := range arr1 {
		m[v] = struct{}{}

	}
	var res = make([]string, 0, 8)
	for _, v := range arr2 {
		if _, ok := m[v]; ok {
			res = append(res, v)
		}
	}
	return res
}

// IShuffle 随机打乱顺序
type IShuffle interface {
	Swap(i, j int)
	Len() int
}

// Shuffle 随机打乱顺序
func Shuffle(arr IShuffle) {
	length := arr.Len()
	for i := 0; i < length; i++ {
		randNum := rand.Intn(length)
		arr.Swap(randNum, i)
	}
}
