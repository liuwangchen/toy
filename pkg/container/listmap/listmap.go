package listmap

import "container/list"

type ListMap struct {
	l *list.List
	m map[interface{}]*list.Element
}

func NewListMap() *ListMap {
	return &ListMap{
		l: list.New(),
		m: make(map[interface{}]*list.Element),
	}
}

func (l *ListMap) PushBack(k, v interface{}) {
	if e, ok := l.m[k]; ok {
		e.Value = v
		l.l.MoveToBack(e)
		return
	}

	e := l.l.PushBack(v)
	l.m[k] = e
}

func (l *ListMap) PushFront(k, v interface{}) {
	if e, ok := l.m[k]; ok {
		e.Value = v
		l.l.MoveToFront(e)
		return
	}
	e := l.l.PushFront(v)
	l.m[k] = e
}

func (l *ListMap) Front() *list.Element {
	return l.l.Front()
}

func (l *ListMap) Back() *list.Element {
	return l.l.Back()
}

func (l *ListMap) RemoveByKey(k interface{}) interface{} {
	e := l.m[k]
	if e == nil {
		return nil
	}
	delete(l.m, k)
	return l.l.Remove(e)
}

func (l *ListMap) Len() int {
	return l.l.Len()
}

func (l *ListMap) Init() {
	l.l.Init()
	l.m = make(map[interface{}]*list.Element)
}

func (l *ListMap) GetValue(k interface{}) interface{} {
	e := l.m[k]
	if e == nil {
		return nil
	}
	return e.Value
}

func (l *ListMap) GetNode(k interface{}) interface{} {
	return l.m[k]
}

func (l *ListMap) Foreach(f func(element *list.Element) bool) {
	for e := l.Front(); e != nil; {
		ce := e
		e = e.Next()
		if !f(ce) {
			break
		}
	}
}
