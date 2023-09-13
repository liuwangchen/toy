package listmap

import (
	"container/list"
	"math/rand"
	"testing"
)

func TestListMap_Init(t *testing.T) {
	l := NewListMap()
	for i := 0; i < 10; i++ {
		l.PushBack(i, i)
	}
	l.PushBack(1, 111)

	l.Foreach(func(v *list.Element) bool {
		t.Log(v.Value)
		return true
	})
	t.Log(l.GetValue(1))
}

func BenchmarkListMap_GetValue(b *testing.B) {
	l := NewListMap()
	for i := 0; i < 100000; i++ {
		l.PushBack(i, rand.Int())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.GetValue(i)
	}
}

func BenchmarkListMap_PushBack(b *testing.B) {
	l := NewListMap()
	for i := 0; i < 100000; i++ {
		l.PushBack(i, rand.Int())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushBack(i, i)
	}
}

func BenchmarkListMap_PushFront(b *testing.B) {
	l := NewListMap()
	for i := 0; i < 100000; i++ {
		l.PushBack(i, rand.Int())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.PushFront(i, i)
	}
}

func BenchmarkListMap_RemoveByKey(b *testing.B) {
	l := NewListMap()
	for i := 0; i < 1000000; i++ {
		l.PushBack(i, rand.Int())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.RemoveByKey(i % 100000)
	}
}

func TestListMap_Remove(t *testing.T) {
	l := NewListMap()
	for i := 0; i < 10; i++ {
		l.PushBack(i, i)
	}
	for e := l.Front(); e != nil; {
		v, _ := e.Value.(int)
		e = e.Next()
		if v%4 == 0 {
			l.RemoveByKey(v) //v not k !!!
		}
	}

	l.Foreach(func(v *list.Element) bool {
		t.Log(v.Value)
		return true
	})
	t.Log(l.GetValue(4) == nil)
	t.Log(l.GetValue(5))

}
