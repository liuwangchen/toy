package redisx

import (
	"strconv"
	"unsafe"

	"github.com/golang/protobuf/proto"
)

// stringToBytes converts string to byte slice.
func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

type StringRet struct {
	err error
	val string
}

func (s *StringRet) Err() error {
	return s.err
}

func (s *StringRet) String() (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.val, nil
}

func (s *StringRet) Int32() (int32, error) {
	if s.err != nil {
		return 0, s.err
	}
	v, err := strconv.ParseInt(s.val, 10, 64)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func (s *StringRet) Bytes() ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return stringToBytes(s.val), nil
}

func (s *StringRet) Proto(msg proto.Message) error {
	data, err := s.Bytes()
	if err != nil {
		return s.err
	}
	return proto.Unmarshal(data, msg)
}

type StringSliceRet struct {
	err error
	val []string
}

func (s *StringSliceRet) Err() error {
	return s.err
}

func (s *StringSliceRet) Strings() ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.val, nil
}

func (s *StringSliceRet) Bytes() ([][]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	ret := make([][]byte, len(s.val))
	for i, v := range s.val {
		ret[i] = stringToBytes(v)
	}
	return ret, nil
}

func (s *StringSliceRet) ForEachBytes(f func([]byte) error) error {
	if s.err != nil {
		return s.err
	}
	for _, v := range s.val {
		if err := f(stringToBytes(v)); err != nil {
			return err
		}
	}
	return nil
}

type MapRet struct {
	val map[string]string
	err error
}

func (m *MapRet) Err() error {
	return m.err
}

func (m *MapRet) Uint64Bytes() (map[uint64][]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	r := make(map[uint64][]byte)
	for k, v := range m.val {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return nil, err
		}
		r[key] = []byte(v)
	}
	return r, nil
}

func (m *MapRet) Uint64Uint32() (map[uint64]uint32, error) {
	if m.err != nil {
		return nil, m.err
	}
	r := make(map[uint64]uint32)
	for k, v := range m.val {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return nil, err
		}
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, err
		}
		r[key] = uint32(val)
	}
	return r, nil
}

func (m *MapRet) Uint64Int64() (map[uint64]int64, error) {
	if m.err != nil {
		return nil, m.err
	}
	r := make(map[uint64]int64)
	for k, v := range m.val {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return nil, err
		}
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, err
		}
		r[key] = int64(val)
	}
	return r, nil
}

func (m *MapRet) ForEachUint64Bytes(f func(uint64, []byte) error) error {
	if m.err != nil {
		return m.err
	}
	for k, v := range m.val {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return err
		}
		if err := f(key, stringToBytes(v)); err != nil {
			return err
		}
	}
	return nil
}

func (m *MapRet) ForEachStringBytes(f func(string, []byte) error) error {
	if m.err != nil {
		return m.err
	}
	for k, v := range m.val {
		if err := f(k, stringToBytes(v)); err != nil {
			return err
		}
	}
	return nil
}
