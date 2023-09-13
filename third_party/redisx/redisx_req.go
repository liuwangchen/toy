package redisx

import (
	"github.com/golang/protobuf/proto"
)

func Encode(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case proto.Message:
		val, err := proto.Marshal(v)
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return value, nil
}

type MapReq []interface{}

// NewMapReq 创建一个MapReq, size为map的len
func NewMapReq(size int) MapReq {
	return make(MapReq, 0, size*2)
}

func (m *MapReq) Add(k interface{}, v interface{}) error {
	data, err := Encode(v)
	if err != nil {
		return err
	}
	*m = append(*m, k, data)
	return nil
}

func NewMapReqStringString(m map[string]string) MapReq {
	h := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		h = append(h, k, v)
	}
	return h
}

func NewMapReqUint64Bytes(value map[uint64][]byte) MapReq {
	var values []interface{}
	for k, v := range value {
		values = append(values, k, v)
	}
	return values
}

type ZAddReq map[float64]interface{}

func NewZAddReq() ZAddReq {
	return make(ZAddReq)
}
func (z ZAddReq) Add(score float64, value interface{}) error {
	data, err := Encode(value)
	if err != nil {
		return err
	}
	z[score] = data
	return nil
}
func (z ZAddReq) Count() int {
	return len(z)
}
