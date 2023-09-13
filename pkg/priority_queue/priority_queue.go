package priority_queue

import "container/heap"

type INode interface {
	Key() string
}

// PriorityQueue 优先级队列
type priorityQueue struct {
	Nodes []INode
	Index map[string]int
	less  func(item1, item2 interface{}) bool
}

// Len 队列长度
func (pq *priorityQueue) Len() int {
	return len(pq.Nodes)
}

// Less 小堆
func (pq *priorityQueue) Less(i, j int) bool {
	return pq.less(pq.Nodes[i], pq.Nodes[j])
}

// Swap 交换位置
func (pq *priorityQueue) Swap(i, j int) {
	pq.Index[pq.Nodes[i].Key()] = j
	pq.Index[pq.Nodes[j].Key()] = i
	pq.Nodes[i], pq.Nodes[j] = pq.Nodes[j], pq.Nodes[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	node := x.(INode)
	pq.Nodes = append(pq.Nodes, node)
	pq.Index[node.Key()] = len(pq.Nodes) - 1
}

func (pq *priorityQueue) Pop() interface{} {
	if pq.Len() < 1 {
		return nil
	}
	old := pq.Nodes
	n := len(old)
	x := old[n-1]
	pq.Nodes = old[0 : n-1]
	delete(pq.Index, x.Key())
	return x
}

type option struct {
	Cap    int // 初始容量
	MaxLen int // 最大长度
}

type OptionFunc func(op *option)

// cap
func Cap(cap int) OptionFunc {
	return func(op *option) {
		op.Cap = cap
	}
}

// maxlen
func MaxLen(maxLen int) OptionFunc {
	return func(op *option) {
		op.MaxLen = maxLen
	}
}

// NewPriorityQueue 创建一个优先级队列
func NewPriorityQueue(less func(item1, item2 interface{}) bool, opts ...OptionFunc) *PriorityQueue {
	op := &option{
		Cap:    10,
		MaxLen: -1,
	}
	for _, opf := range opts {
		opf(op)
	}
	pq := &PriorityQueue{
		op: op,
		pq: &priorityQueue{
			Nodes: make([]INode, 0, op.Cap),
			Index: make(map[string]int, op.Cap),
			less:  less,
		},
	}
	return pq
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	pq *priorityQueue
	op *option
}

func (this *PriorityQueue) Push(x interface{}) {
	heap.Push(this.pq, x)
	// 超过最大长度删除最后一个
	if this.op.MaxLen > 0 && this.Len() > this.op.MaxLen {
		this.Remove(this.pq.Nodes[this.Len()-1])
	}
}

func (this *PriorityQueue) Pop() interface{} {
	return heap.Pop(this.pq)
}

// Remove remove
func (this *PriorityQueue) Remove(node INode) {
	index, ok := this.pq.Index[node.Key()]
	if !ok {
		return
	}
	heap.Remove(this.pq, index)
}

// RemoveByKey remove
func (this *PriorityQueue) RemoveByKey(key string) {
	index, ok := this.pq.Index[key]
	if !ok {
		return
	}
	heap.Remove(this.pq, index)
}

// Top top
func (this *PriorityQueue) Top() interface{} {
	if this.pq.Len() < 1 {
		return nil
	}
	return this.pq.Nodes[0]
}

// Update update
func (this *PriorityQueue) Update(key string) {
	index, ok := this.pq.Index[key]
	if !ok {
		return
	}
	heap.Fix(this.pq, index)
}

// GetByKey 通过Key获取Node
func (this *PriorityQueue) GetByKey(key string) interface{} {
	for _, node := range this.pq.Nodes {
		if node.Key() == key {
			return node
		}
	}
	return nil
}

func (this *PriorityQueue) Contains(node INode) bool {
	return this.GetByKey(node.Key()) != nil
}

// Len 队列长度
func (this *PriorityQueue) Len() int {
	return this.pq.Len()
}

func (this *PriorityQueue) GetAll() []INode {
	return this.pq.Nodes
}
