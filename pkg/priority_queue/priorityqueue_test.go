package priority_queue

import (
	"strconv"
	"testing"
)

func NewNode(key string, weight int64) *Node {
	return &Node{
		key:    key,
		weight: weight,
	}
}

// Node 一个节点
type Node struct {
	key    string
	weight int64
}

func (n *Node) Key() string {
	return n.key
}

func (n *Node) Weight() int64 {
	return n.weight
}

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(func(item1, item2 interface{}) bool {
		return item1.(*Node).weight < item2.(*Node).weight
	})

	num := 100
	for i := 0; i < num; i++ {
		node := NewNode(strconv.Itoa(i), int64(i))
		pq.Push(node)
	}

	var max int64 = -1

	for i := 0; i < num; i++ {
		topNode := pq.Top().(*Node)
		popNode := pq.Pop().(*Node)
		if topNode == nil {
			t.Error()
			return
		}
		if popNode == nil {
			t.Error()
			return
		}

		if topNode.Key() != popNode.Key() {
			t.Error()
			return
		}
		if topNode.Weight() != popNode.Weight() {
			t.Error()
			return
		}

		if max > topNode.Weight() {
			t.Error()
			return
		}

		max = topNode.Weight()
	}
}

func TestPriorityQueue_Update(t *testing.T) {
	pq := NewPriorityQueue(func(item1, item2 interface{}) bool {
		return item1.(*Node).weight < item2.(*Node).weight
	})
	num := 10
	for i := 0; i < num; i++ {
		node := NewNode(strconv.Itoa(i), int64(i))
		pq.Push(node)
	}

	for i := 0; i < num; i += 2 {
		node := pq.GetByKey(strconv.Itoa(i)).(*Node)
		node.weight = int64(i - 1)
		pq.Update(strconv.Itoa(i))
	}

	var max int64 = -10
	for i := 0; i < num; i++ {
		node := pq.Pop().(*Node)
		if node.Weight() < max {
			t.Error()
			return
		}
		max = node.Weight()
	}
}

func TestPriorityQueue_Remove(t *testing.T) {
	pq := NewPriorityQueue(func(item1, item2 interface{}) bool {
		return item1.(*Node).weight < item2.(*Node).weight
	})

	num := 10
	for i := 0; i < num; i++ {
		node := NewNode(strconv.Itoa(i), int64(i))
		pq.Push(node)
	}

	for i := 0; i < num; i += 2 {
		pq.RemoveByKey(strconv.Itoa(i))
	}

	if pq.Len() != 5 {
		t.Error()
		return
	}

	var max int64 = -1
	for pq.Len() > 0 {
		node := pq.Pop().(*Node)
		if node.Weight() < max {
			t.Error()
			return
		}
		max = node.Weight()
	}
}

func TestPriorityQueue_Remove2(t *testing.T) {
	pq := NewPriorityQueue(func(item1, item2 interface{}) bool {
		return item1.(*Node).weight < item2.(*Node).weight
	})

	num := 10
	for i := 0; i < num; i++ {
		node := NewNode(strconv.Itoa(i), int64(i))
		pq.Push(node)
	}

	for i := 0; i < num; i++ {
		pq.RemoveByKey(strconv.Itoa(i))
	}
}
