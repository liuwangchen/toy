// Package zset implements sorted set similar to redis zset.
package zset

import (
	"errors"
	"math/rand"
	"time"
)

const (
	DefaultMaxLevel = 32   // (1/p)^MaxLevel >= maxNode
	DefaultP        = 0.25 // SkipList P = 1/4

	DefaultFreeListSize = 32
)

var ErrItemCompare = errors.New("item compare wrong")

var nilNodes = make([]skipListLevel, 16)

// Item represents a single object in the set.
type Item interface {
	Less(Item) bool
}

// ItemIterator allows callers of Range* to iterate of the zset.
// When this function returns false, iteration will stop.
type ItemIterator func(key string, i Item, rank int) bool

type skipListLevel struct {
	forward *node
	span    int
}

// node is an element of a skip list
type node struct {
	key      string
	item     Item
	backward *node
	level    []skipListLevel
}

// FreeList represents a free list of set node.
type FreeList struct {
	freelist []*node
}

// NewFreeList creates a new free list.
func NewFreeList(size int) *FreeList {
	return &FreeList{freelist: make([]*node, 0, size)}
}

func (f *FreeList) newNode(lvl int) (n *node) {
	if len(f.freelist) == 0 {
		n = new(node)
		n.level = make([]skipListLevel, lvl)
		return
	}
	index := len(f.freelist) - 1
	n = f.freelist[index]
	f.freelist[index] = nil
	f.freelist = f.freelist[:index]

	if cap(n.level) < lvl {
		n.level = make([]skipListLevel, lvl)
	} else {
		n.level = n.level[:lvl]
	}
	return
}

func (f *FreeList) freeNode(n *node) (out bool) {
	// for gc
	n.item = nil
	toClear := n.level
	for len(toClear) > 0 {
		toClear = toClear[copy(toClear, nilNodes):]
	}

	if len(f.freelist) < cap(f.freelist) {
		f.freelist = append(f.freelist, n)
		out = true
	}
	return
}

// skipList represents a skip list
type skipList struct {
	header, tail *node
	length       int
	level        int // current level count
	maxLevel     int
	freelist     *FreeList
	random       *rand.Rand
}

// newSkipList creates a skip list
func newSkipList(maxLevel int) *skipList {
	if maxLevel < DefaultMaxLevel {
		panic("maxLevel must < 32")
	}
	return &skipList{
		level: 1,
		header: &node{
			level: make([]skipListLevel, maxLevel),
		},
		maxLevel: maxLevel,
		freelist: NewFreeList(DefaultFreeListSize),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// insert an item into the SkipList.
func (sl *skipList) insert(key string, item Item) *node {
	var update [DefaultMaxLevel]*node // [0...list.maxLevel)
	var rank [DefaultMaxLevel]int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for y := x.level[i].forward; y != nil && y.item.Less(item); y = x.level[i].forward {
			rank[i] += x.level[i].span
			x = y
		}
		update[i] = x
	}

	lvl := sl.randomLevel()
	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].level[i].span = sl.length
		}
		sl.level = lvl
	}

	x = sl.freelist.newNode(lvl)
	x.item = item
	x.key = key
	for i := 0; i < lvl; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// increment span for untouched levels
	for i := lvl; i < sl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == sl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward == nil {
		sl.tail = x
	} else {
		x.level[0].forward.backward = x
	}
	sl.length++
	return x
}

// delete element
func (sl *skipList) delete(n *node) Item {
	var preAlloc [DefaultMaxLevel]*node // [0...list.maxLevel)
	update := preAlloc[:sl.maxLevel]
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.level[i].forward; y != nil && y.item.Less(n.item); y = x.level[i].forward {
			x = y
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && x.key == n.key && !n.item.Less(x.item) {
		for i := 0; i < sl.level; i++ {
			if update[i].level[i].forward == x {
				update[i].level[i].span += x.level[i].span - 1
				update[i].level[i].forward = x.level[i].forward
			} else {
				update[i].level[i].span--
			}
		}
		for sl.level > 1 && sl.header.level[sl.level-1].forward == nil {
			sl.level--
		}
		if x.level[0].forward == nil {
			sl.tail = x.backward
		} else {
			x.level[0].forward.backward = x.backward
		}
		removeItem := x.item
		sl.freelist.freeNode(x)
		sl.length--
		return removeItem
	}
	return nil
}

func (sl *skipList) updateItem(node *node, item Item) bool {
	if (node.level[0].forward == nil || !node.level[0].forward.item.Less(item)) &&
		(node.backward == nil || !item.Less(node.backward.item)) {
		node.item = item
		return true
	}
	return false
}

// GetRank find the rank for an element.
// Returns 0 when the element cannot be found, rank otherwise.
// Note that the rank is 1-based
func (sl *skipList) GetRank(item Item) int {
	var rank int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for y := x.level[i].forward; y != nil && !item.Less(y.item); y = x.level[i].forward {
			rank += x.level[i].span
			x = y
		}
		if x.item != nil && !x.item.Less(item) {
			return rank
		}
	}
	return 0
}

func (sl *skipList) randomLevel() int {
	lvl := 1
	for lvl < sl.maxLevel && float32(sl.random.Uint32()&0xFFFF) < DefaultP*0xFFFF {
		lvl++
	}
	return lvl
}

// Finds an element by its rank. The rank argument needs to be 1-based.
func (sl *skipList) getElementByRank(rank int) *node {
	var traversed int
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && traversed+x.level[i].span <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

// ZSet set
type ZSet struct {
	dict map[string]*node
	sl   *skipList
}

// New creates a new ZSet.
func New() *ZSet {
	return &ZSet{
		dict: make(map[string]*node),
		sl:   newSkipList(DefaultMaxLevel),
	}
}

// Add a new element or update the score of an existing element. If an key already
// exist, the removed item is returned. Otherwise, nil is returned.
func (zs *ZSet) Add(key string, item Item) (removeItem Item, err error) {
	if node := zs.dict[key]; node != nil {
		// if the node after update, would be still exactly at the same position,
		// we can just update item.
		if zs.sl.updateItem(node, item) {
			return
		}
		removeItem = zs.sl.delete(node)
		if removeItem == nil {
			return removeItem, ErrItemCompare
		}
	}
	zs.dict[key] = zs.sl.insert(key, item)
	return
}

// Remove the element 'ele' from the sorted set,
// return true if the element existed and was deleted, false otherwise
func (zs *ZSet) Remove(key string) (removeItem Item, err error) {
	node := zs.dict[key]
	if node == nil {
		return
	}
	removeItem = zs.sl.delete(node)
	if removeItem == nil {
		err = ErrItemCompare
		return
	}
	delete(zs.dict, key)
	return
}

// Rank return 1-based rank or 0 if not exist
func (zs *ZSet) Rank(key string, reverse bool) int {
	node := zs.dict[key]
	if node != nil {
		rank := zs.sl.GetRank(node.item)
		if rank > 0 {
			if reverse {
				return zs.sl.length - rank + 1
			}
			return rank
		}
	}
	return 0
}

// Range return elements in [start, end].
func (zs *ZSet) Range(start, end int, reverse bool, iterator ItemIterator) {
	llen := zs.sl.length
	if start < 0 {
		start = llen + start
	}
	if end < 0 {
		end = llen + end
	}
	if start < 0 {
		start = 0
	}
	if start > end || start >= llen {
		return
	}
	if end >= llen {
		end = llen - 1
	}

	rangeLen := end - start + 1
	if reverse {
		ln := zs.sl.getElementByRank(llen - start)
		for i := 1; i <= rangeLen; i++ {
			if iterator(ln.key, ln.item, start+i) {
				ln = ln.backward
			} else {
				break
			}
		}
	} else {
		ln := zs.sl.getElementByRank(start + 1)
		for i := 1; i <= rangeLen; i++ {
			if iterator(ln.key, ln.item, start+i) {
				ln = ln.level[0].forward
			} else {
				break
			}
		}
	}
}

// RangeIterator return iterator for visit elements in [start, end].
// It is slower than Range.
func (zs *ZSet) RangeIterator(start, end int, reverse bool) RangeIterator {
	llen := zs.sl.length
	if start < 0 {
		start = llen + start
	}
	if end < 0 {
		end = llen + end
	}
	if start < 0 {
		start = 0
	}

	if start > end || start >= llen {
		return RangeIterator{end: -1}
	}

	if end >= llen {
		end = llen - 1
	}

	var n *node
	if reverse {
		n = zs.sl.getElementByRank(llen - start)
	} else {
		n = zs.sl.getElementByRank(start + 1)
	}
	return RangeIterator{
		start:   start,
		cur:     start,
		end:     end,
		node:    n,
		reverse: reverse,
	}
}

// Get return Item in dict.
func (zs *ZSet) Get(key string) Item {
	if node, ok := zs.dict[key]; ok {
		return node.item
	}
	return nil
}

// Length return the element count
func (zs *ZSet) Length() int {
	return zs.sl.length
}

type Int int

func (a Int) Less(b Item) bool {
	return a < b.(Int)
}
