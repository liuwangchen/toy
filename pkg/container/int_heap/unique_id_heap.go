package int_heap

import (
	"container/heap"
	"fmt"
)

// An IntHeap is a min-heap of ints.
type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

//返回[min,max]区间的数据
type UniqueId struct {
	min, max     int
	cur          int
	release_map  map[int]bool
	release_heap IntHeap
}

func NewUniqueId(min, max int) *UniqueId {
	if min > max || min < 0 || max < 0 {
		return nil
	}
	u := &UniqueId{
		min:         min,
		max:         max,
		cur:         min,
		release_map: make(map[int]bool),
	}
	heap.Init(&u.release_heap)
	return u
}

func (this *UniqueId) Get() (int, bool) {
	if this.release_heap.Len() >= 50 {
		id := heap.Pop(&this.release_heap).(int)
		delete(this.release_map, id)
		return id, true
	}
	if this.cur <= this.max {
		id := this.cur
		this.cur++
		return id, true
	}
	if this.release_heap.Len() != 0 {
		id := heap.Pop(&this.release_heap).(int)
		delete(this.release_map, id)
		return id, true
	}
	return -1, false
}

func (this *UniqueId) Put(id int) bool {
	if id < this.min || id > this.max {
		return false
	}
	if _, e := this.release_map[id]; !e {
		heap.Push(&this.release_heap, id)
		this.release_map[id] = true
		return true
	}
	return false
}

func (this *UniqueId) Reset() {
	this.cur = this.min
	this.release_map = make(map[int]bool)
}

func (this *UniqueId) String() string {
	return fmt.Sprintf("UniqueId Info: min-%d,max-%d,cur-%d,map_size-%d,heap_size-%d", this.min, this.max, this.cur, len(this.release_map), this.release_heap.Len())
}
