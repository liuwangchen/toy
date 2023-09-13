package maps

import (
	"github.com/liuwangchen/toy/pkg/algorithm/astar"
)

type ShapeMap struct {
	IMap
	row, col int32
}

func NewShapeMap(row, col int32, imap IMap) *ShapeMap {
	return &ShapeMap{
		IMap: imap,
		row:  row,
		col:  col,
	}
}

// IsLicit 是不是合法的坐标
func (m *ShapeMap) IsLicit(coor Coordinate) bool {
	return coor[0] >= 0 && coor[0] < m.col && coor[1] >= 0 && coor[1] < m.row
}

var (
	leftUp_Shape    = [2]int32{-1, 1}
	leftDown_Shape  = [2]int32{-1, -1}
	rightUp_Shape   = [2]int32{1, 1}
	rightDown_Shape = [2]int32{1, -1}

	dirs = [][2]int32{leftUp_Shape, leftDown_Shape, rightUp_Shape, rightDown_Shape}
)

func (m *ShapeMap) GetAround(coor Coordinate) []Coordinate {
	result := make([]Coordinate, 0, 4)
	for _, dir := range dirs {
		result = append(result, Coordinate{coor[0] + dir[0], coor[1] + dir[1]})
	}
	return result
}

// todo shape坐标系的getLen
func (m *ShapeMap) GetLen(src, dest Coordinate) int32 {
	return 0
}

// 周围有效的点
func (m *ShapeMap) Around(coor astar.ICoor) []astar.ICoor {
	coors := m.GetAround(Coordinate{coor.X(), coor.Y()})
	result := make([]astar.ICoor, 0, len(coors))
	for _, coordinate := range coors {
		// 阻挡忽略
		if m.IsBlock(coordinate) {
			continue
		}
		// 不合法忽略
		if !m.IsLicit(coordinate) {
			continue
		}
		result = append(result, coordinate)
	}
	return result
}

// 点的代价
func (m *ShapeMap) Cost(coor astar.ICoor) int32 {
	return 1
}

// 预估两点距离
func (m *ShapeMap) Estimate(src, dest astar.ICoor) int32 {
	return m.GetLen(Coordinate{src.X(), src.Y()}, Coordinate{dest.X(), dest.Y()})
}

// A*寻路算法
func (m *ShapeMap) FindPath(src Coordinate, dest Coordinate) []Coordinate {
	// 目标点和初始点相同
	if src == dest {
		return nil
	}
	path := astar.NewAstar(m).FindPath(src, dest)
	if len(path) == 0 {
		return nil
	}

	// 目标点是障碍，删掉
	if m.IsBlock(dest) {
		path = path[:len(path)-1]
	}

	result := make([]Coordinate, 0, len(path))
	for _, coor := range path {
		result = append(result, Coordinate{coor.X(), coor.Y()})
	}
	return result
}
