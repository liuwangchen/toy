package stringx

import (
	"reflect"
	"strings"
	"unsafe"
)

// USE AT YOUR OWN RISK

// String force casts a []byte to a string.
func String(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

//Bytes force casts a string to a []byte
//禁止对[]byte赋值
func Bytes(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

// 字符串拼接
func StringSplicing(option ...string) string {
	var b strings.Builder
	b.Grow(len(option))
	for _, str := range option {
		b.WriteString(str)
	}
	return b.String()
}
