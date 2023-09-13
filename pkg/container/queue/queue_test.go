package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	queue := &Queue{}
	for i := 0; i < 8; i++ {
		queue.Push(i)
	}
	for i := 0; i < 7; i++ {
		queue.Pop()
	}
	queue.Push(1)
	v := queue.Get(0)
	if v != 7 {
		t.Error()
	}
	v = queue.Get(1)
	if v != 1 {
		t.Error()
	}
	queue.Pop()
	queue.Pop()
	queue.Push(0)
	v = queue.Pop()
	if v != 0 {
		t.Error()
	}
	queue.Push(1)
	v = queue.Get(0)
	if v != 1 {
		t.Error()
	}
	queue.Push(2)
	v = queue.Get(1)
	if v != 2 {
		t.Error()
	}
	v = queue.Get(0)
	if v != 1 {
		t.Error()
	}
	v = queue.Pop()
	if v != 1 {
		t.Error()
	}
	v = queue.Get(0)
	if v != 2 {
		t.Error()
	}

	v = queue.Get(1)
	if v != nil {
		t.Error()
	}
}

func TestQueue1(t *testing.T) {
	queue := &Queue{}
	for i := 0; i < 6; i++ {
		queue.Push(i)
	}
	for i := 0; i < 6; i++ {
		queue.Pop()
	}
	for i := 0; i < 7; i++ {
		queue.Push(i)
	}
	v := queue.Get(1)
	if v != 1 {
		t.Error()
	}
}
