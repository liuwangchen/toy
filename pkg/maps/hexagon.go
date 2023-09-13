package maps

import (
	"errors"

	"github.com/liuwangchen/toy/pkg/algorithm/astar"
)

type HexagonMap struct {
	IMap
	row, col int32
}

func NewHexagonMap(row, col int32, imap IMap) *HexagonMap {
	return &HexagonMap{
		IMap: imap,
		row:  row,
		col:  col,
	}
}

// IsLicit 是不是合法的坐标
func (m *HexagonMap) IsLicit(coor Coordinate) bool {
	return coor[0] >= 0 && coor[0] < m.col && coor[1] >= 0 && coor[1] < m.row
}

// GetZonesRingBlock 一个区域一个环的坐标 只返回Block点
// @param -- coordinates 区域 -- rings 第几环
// @return rings环的坐标
func (m *HexagonMap) GetZonesRingBlock(coordinates []Coordinate, ring int32) ([]Coordinate, error) {
	return m.GetZonesRingsBlock(coordinates, []int32{ring})
}

// getMaxRing 最大还数
func (m *HexagonMap) getMaxRing(rings []int32) int {
	var result int
	for _, ring := range rings {
		if int(ring) > result {
			result = int(ring)
		}
	}
	return result
}

// GetZonesRingsBlock 一个区域几个环的坐标 只返回Block点
// @param -- coordinates 区域 -- rings 第几环
// @return rings环的坐标
func (m *HexagonMap) GetZonesRingsBlock(coordinates []Coordinate, rings []int32) ([]Coordinate, error) {
	result := make([]Coordinate, 0, m.getZonesRingResultLength(rings))
	maxRing := m.getMaxRing(rings)

	var (
		nextRingCoors = coordinates // 下一个环的坐标
	)
	findCoors := make(map[Coordinate]struct{})
	// 忽略自己
	for _, coordinate := range nextRingCoors {
		findCoors[coordinate] = struct{}{}
	}
	for i := 1; i <= maxRing; i++ {
		// 获取一个区域的一圈坐标
		coors, err := m.GetZonesOuterRing(nextRingCoors)
		if err != nil {
			return nil, err
		}
		nextRingCoors = nextRingCoors[:0]
		// 求并集
		for _, coor := range coors {
			// 已经找过
			_, ok := findCoors[coor]
			if ok {
				continue
			}
			findCoors[coor] = struct{}{}
			// 下一个环的坐标
			nextRingCoors = append(nextRingCoors, coor)
			// 存在
			if m.IsBlock(coor) && m.isInIntArray(rings, int32(i)) {
				result = append(result, coor)
			}
		}
	}

	return result, nil
}

// BeelineCoordinateBlockRange 坐标直线 只返回Block点
// @param -- coordinate 坐标 affectRange [][方向，长度]] -- resist 每个方向找几个
// @return 坐标
func (m *HexagonMap) BeelineCoordinateBlockRange(coordinate Coordinate, affectRange [][2]int32, resist int32) []Coordinate {
	result := make([]Coordinate, 0, 10)
	for _, param := range affectRange {
		num := param[1]
		// -1 是无限
		if num == -1 {
			num = m.col - coordinate.X()
		}
		var count int32
		m.ForEachN(coordinate, DirType(param[0]), int(num), func(coor Coordinate) bool {
			// 越界
			if !m.IsLicit(coor) {
				return false
			}
			if !m.IsBlock(coor) {
				return true
			}
			count++
			result = append(result, coor)
			return count < resist
		})
	}
	return result
}

// BeelineZoneBlockRange 区域直线 只返回Block点
// @param -- zones 区域 affectRange [][方向，长度]] -- resist 每个方向找几个
// @return 坐标
func (m *HexagonMap) BeelineZoneBlockRange(zones []Coordinate, affectRange [][2]int32, resist int32) []Coordinate {
	result := make([]Coordinate, 0, 10)
	findCoors := make(map[Coordinate]struct{})
	// 忽略自己
	for _, coordinate := range zones {
		findCoors[coordinate] = struct{}{}
	}
	for _, coordinate := range zones {
		retCoors := m.BeelineCoordinateBlockRange(coordinate, affectRange, resist)
		for _, coor := range retCoors {
			_, ok := findCoors[coor]
			if ok {
				continue
			}
			findCoors[coor] = struct{}{}
			result = append(result, coor)
		}
	}
	return result
}

// AllBlockRange 全图
func (m *HexagonMap) AllBlockRange() []Coordinate {
	result := make([]Coordinate, 0, 10)
	for i := 0; i < int(m.row); i++ {
		for j := 0; j < int(m.col); j++ {
			coor := Coordinate{int32(j), int32(i)}
			if !m.IsBlock(coor) {
				continue
			}
			result = append(result, coor)
		}
	}
	return result
}

// DirType 移动类型
type DirType int32

const (
	Up        DirType = 1 // 上
	RightUp   DirType = 2 // 右上
	RightDown DirType = 3 // 右下
	Down      DirType = 4 // 下
	LeftDown  DirType = 5 // 左下
	LeftUp    DirType = 6 // 左上
)

// getOffsetByRangeType 获取偏移
func getOffsetByRangeType(rangeType DirType) moveOffset {
	switch rangeType {
	default:
		return unknown
	case Up:
		return up
	case RightUp:
		return rightUp
	case RightDown:
		return rightDown
	case Down:
		return down
	case LeftDown:
		return leftDown
	case LeftUp:
		return leftUp
	}
}

// moveOffset 偏移
type moveOffset [2][2]int32

var (
	up        = moveOffset{{0, 1}, {0, 1}}
	down      = moveOffset{{0, -1}, {0, -1}}
	leftUp    = moveOffset{{-1, 1}, {-1, 0}}
	leftDown  = moveOffset{{-1, 0}, {-1, -1}}
	rightUp   = moveOffset{{1, 1}, {1, 0}}
	rightDown = moveOffset{{1, 0}, {1, -1}}
	unknown   = moveOffset{{0, 0}, {0, 0}}
)

// GetCoorByDir 根据坐标获取某个方向的坐标
func (m *HexagonMap) GetCoorByDir(coordinate Coordinate, dirType DirType) Coordinate {
	offset := getOffsetByRangeType(dirType)
	addOffset := offset[0]
	if coordinate.X()&1 == 1 {
		addOffset = offset[1]
	}
	return Coordinate{
		coordinate.X() + addOffset[0],
		coordinate.Y() + addOffset[1],
	}
}

// GetAllCoordinateByDir 获取某方向距离为num的所有点
func (m *HexagonMap) GetAllCoordinateByDir(coordinate Coordinate, rangeType DirType, num int) (coors []Coordinate) {
	coor := coordinate
	for i := 0; i < num; i++ {
		coor = m.GetCoorByDir(coor, rangeType)
		coors = append(coors, coor)
	}
	return
}

// ForEachN
func (m *HexagonMap) ForEachN(coordinate Coordinate, rangeType DirType, num int, eachFunc func(coor Coordinate) bool) {
	coor := coordinate
	for i := 0; i < num; i++ {
		coor = m.GetCoorByDir(coor, rangeType)
		if !eachFunc(coor) {
			return
		}
	}
}

// GetCoordinateOuterRing 获取一个点的一圈坐标
// @param -- coordinate 坐标
// @return 一个点的一圈坐标
func (m *HexagonMap) GetCoordinateOuterRing(coordinate Coordinate) ([]Coordinate, error) {
	rangeTypes := []DirType{
		RightDown, Down, LeftDown, LeftUp, Up, RightUp,
	}
	// 向上移动
	coors := m.GetAllCoordinateByDir(coordinate, Up, 1)
	result := make([]Coordinate, 0, len(rangeTypes))
	// 向六个方向移动
	for _, rangeType := range rangeTypes {
		if len(coors) < 1 {
			return nil, errors.New("coordinate find error")
		}
		coors = m.GetAllCoordinateByDir(coors[len(coors)-1], rangeType, 1)
		result = append(result, coors...)
	}
	return result, nil
}

// getOuterRingResultLength
func (m *HexagonMap) getOuterRingResultLength(length int) int {
	if length < 2 {
		return 6
	}
	return length*2 + 2*4
}

// GetZonesOuterRing 获取一个区域的一圈坐标
// @param -- zones 区域
// @return 一个区域的一圈坐标
func (m *HexagonMap) GetZonesOuterRing(zones []Coordinate) ([]Coordinate, error) {
	result := make([]Coordinate, 0, m.getOuterRingResultLength(len(zones)))
	findCoors := make(map[Coordinate]struct{})
	// 忽略自己
	for _, coordinate := range zones {
		findCoors[coordinate] = struct{}{}
	}
	for _, coordinate := range zones {
		// 获取一个点的一圈坐标
		coors, err := m.GetCoordinateOuterRing(coordinate)
		if err != nil {
			return nil, err
		}
		// 求并集
		for _, coor := range coors {
			_, ok := findCoors[coor]
			if ok {
				continue
			}
			findCoors[coor] = struct{}{}
			result = append(result, coor)
		}
	}
	return result, nil
}

// getZonesRingResultLength getZonesRingResultLength坐标数量
func (m *HexagonMap) getZonesRingResultLength(rings []int32) int {
	var result int
	for _, ring := range rings {
		result += int(ring) * 6
	}
	return result
}

func (m *HexagonMap) isInIntArray(arr []int32, target int32) bool {
	for _, num := range arr {
		if num == target {
			return true
		}
	}
	return false
}

// 获取六边形两点间距离
func (m *HexagonMap) GetLength(src Coordinate, dest Coordinate) int32 {
	if src.X() > dest.X() {
		src, dest = dest, src
	}
	var yup, ydown int32
	if src.X()%2 == 0 {
		yup = src.Y() + (dest.X()-src.X()+1)/2
		ydown = src.Y() - (dest.X()-src.X())/2
	} else {
		yup = src.Y() + (dest.X()-src.X())/2
		ydown = src.Y() - (dest.X()-src.X()+1)/2
	}
	if dest.Y() > yup {
		return dest.X() - src.X() + dest.Y() - yup
	}
	if dest.Y() < ydown {
		return dest.X() - src.X() + ydown - dest.Y()
	}
	return dest.X() - src.X()
}

func (m *HexagonMap) GetMinLengthCoordinate(coor Coordinate, coors []Coordinate) Coordinate {
	if len(coors) == 0 {
		return coor
	}
	d := coors[0]
	minLength := m.GetLength(coor, d)
	for _, coordinate := range coors[1:] {
		length := m.GetLength(coor, coordinate)
		if length < minLength {
			d = coordinate
			minLength = length
		}
	}
	return d
}

// 周围有效的点
func (m *HexagonMap) Around(coor astar.ICoor) []astar.ICoor {
	ring, _ := m.GetCoordinateOuterRing(Coordinate{coor.X(), coor.Y()})
	result := make([]astar.ICoor, 0, len(ring))
	for _, coordinate := range ring {
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
func (m *HexagonMap) Cost(coor astar.ICoor) int32 {
	return 1
}

// 预估两点距离
func (m *HexagonMap) Estimate(src, dest astar.ICoor) int32 {
	return m.GetLength(Coordinate{src.X(), src.Y()}, Coordinate{dest.X(), dest.Y()})
}

// A*寻路算法
func (m *HexagonMap) FindPath(src Coordinate, dest Coordinate) []Coordinate {
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
