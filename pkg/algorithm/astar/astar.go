package astar

import (
	"fmt"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/liuwangchen/toy/pkg/priority_queue"
)

type ICoor interface {
	X() int32
	Y() int32
}

type IMap interface {
	Around(coor ICoor) []ICoor
	Cost(coor ICoor) int32
	Estimate(src, dest ICoor) int32
}

type node struct {
	g      int32
	h      int32
	coor   ICoor
	parent *node
}

func (n *node) Key() string {
	return fmt.Sprintf("%d,%d", n.coor.X(), n.coor.Y())
}

func (n *node) GetF() int32 {
	return n.g + n.h
}

type Astar struct {
	m       IMap
	nodeMap map[string]*node
}

func NewAstar(m IMap) *Astar {
	return &Astar{m: m, nodeMap: map[string]*node{}}
}

func (as *Astar) findAroundNode(n *node) []*node {
	coors := as.m.Around(n.coor)
	result := make([]*node, 0, len(coors))
	for _, coor := range coors {
		result = append(result, as.createNode(coor))
	}
	return result
}

func (as *Astar) createNode(coor ICoor) *node {
	key := fmt.Sprintf("%d,%d", coor.X(), coor.Y())
	n, ok := as.nodeMap[key]
	if !ok {
		n = &node{coor: coor}
		as.nodeMap[key] = n
	}
	return n
}

func (as *Astar) FindPath(src, dest ICoor) []ICoor {
	// 目标点和初始点相同
	if src.X() == dest.X() && src.Y() == dest.Y() {
		return nil
	}
	src, dest = dest, src
	as.nodeMap = make(map[string]*node)

	// 开放列表，优先级队列
	// F最小在堆顶，F值一样取G大的
	openList := priority_queue.NewPriorityQueue(func(i, j interface{}) bool {
		nodei := i.(*node)
		nodej := j.(*node)
		if nodei.GetF() == nodej.GetF() {
			return nodei.g > nodej.g
		}
		return nodei.GetF() < nodej.GetF()
	})
	// 关闭列表
	closeList := linkedhashmap.New()

	// 起始点
	currentNode := as.createNode(src)
	currentNode.g = 0
	currentNode.h = as.m.Estimate(src, dest)

	// 将开始的地格添加到开放列表中。
	openList.Push(currentNode)

	dealCount := 0
	for openList.Len() > 0 && dealCount < 10000 {
		dealCount++
		// 获得F值最低的那块地格。
		// 将当前地格从开放列表中移除，并将其添加到封闭列表中。
		currentNode = openList.Pop().(*node)
		closeList.Put(currentNode.coor, currentNode)

		// 如果在关闭的列表中有一个目标地格，我们就找到了一个路径。
		_, found := closeList.Get(dest)
		if found {
			break
		}

		// 调查当前地格的每一块相邻的地格。
		aroundNodes := as.findAroundNode(currentNode)
		for _, aroundNode := range aroundNodes {
			// 忽略已经在关闭列表中的地格。
			_, found := closeList.Get(aroundNode.coor)
			if found {
				continue
			}
			g := currentNode.g + as.m.Cost(aroundNode.coor)

			// 如果它不在开放列表中--添加它并计算G和H。
			found = openList.Contains(aroundNode)
			if !found {
				aroundNode.parent = currentNode
				aroundNode.g = g
				aroundNode.h = as.m.Estimate(aroundNode.coor, dest)
				openList.Push(aroundNode)
			} else if aroundNode.GetF() > g+aroundNode.h {
				// 检查使用当前的G是否可以得到一个更低的F值，如果可以的话，更新它的值。
				aroundNode.g = g
				openList.Update(aroundNode.Key())
			}
		}
	}

	destNode, found := closeList.Get(dest)
	// 没有路径
	if !found {
		return nil
	}

	// 回溯
	result := make([]ICoor, 0, closeList.Size())
	result = append(result, destNode.(*node).coor)
	p := destNode.(*node).parent
	for p != nil {
		result = append(result, p.coor)
		p = p.parent
	}
	return result
}
