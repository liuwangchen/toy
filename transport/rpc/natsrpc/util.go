package natsrpc

import (
	"go/ast"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
)

// NewEnc 创建enc
func NewEnc(url string, encType string, option ...nats.Option) (*nats.EncodedConn, error) {
	nc, err := nats.Connect(url, option...)
	if err != nil {
		return nil, err
	}
	enc, err1 := nats.NewEncodedConn(nc, encType)
	if nil != err1 {
		return nil, err1
	}
	return enc, nil
}

// NewPBEnc 创建 protobuf enc
func NewPBEnc(url string, option ...nats.Option) (*nats.EncodedConn, error) {
	return NewEnc(url, protobuf.PROTOBUF_ENCODER, option...)
}

// NewJSONEnc 创建 json enc
func NewJSONEnc(url string, option ...nats.Option) (*nats.EncodedConn, error) {
	return NewEnc(url, nats.JSON_ENCODER, option...)
}

// NewGobEnc 创建 json enc
func NewGobEnc(url string, option ...nats.Option) (*nats.EncodedConn, error) {
	return NewEnc(url, nats.GOB_ENCODER, option...)
}

// CombineStr 组合字符串成subject
func CombineStr(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	src := make([]string, 0, len(s))
	for _, v := range s {
		if len(v) == 0 {
			continue
		}
		src = append(src, v)
	}
	return strings.Join(src, ".")
}

// isExportedOrBuiltinType 是导出或内置类型
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// IsProtoPtrType 是否是proto指针类型
func IsProtoPtrType(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return false
	}
	_, ok := reflect.New(t.Elem()).Interface().(proto.Message)
	return ok
}

// IsErrorType 是否是error类型
func IsErrorType(t reflect.Type) bool {
	return t == reflect.TypeOf((*error)(nil)).Elem()
}

// IsContextType 是否是context类型
func IsContextType(t reflect.Type) bool {
	if t.Kind() != reflect.Interface {
		return false
	}
	if t.String() != "context.Context" {
		return false
	}
	return true
}

func IfNotNilPanic(err error) {
	if err != nil {
		panic(err)
	}
}
