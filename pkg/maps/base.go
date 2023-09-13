package maps

type Coordinate [2]int32

func (coordinate Coordinate) X() int32 {
	return coordinate[0]
}

func (coordinate Coordinate) Y() int32 {
	return coordinate[1]
}

type IMap interface {
	IsBlock(coordinate Coordinate) bool
}
