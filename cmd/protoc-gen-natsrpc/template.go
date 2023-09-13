package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

var sTpl *template.Template

//go:embed tmpl.tepl
var serviceTmpl string

func init() {
	var err error
	sTpl, err = template.New("tmpl").Funcs(sprig.TxtFuncMap()).Parse(serviceTmpl)
	if err != nil {
		panic(err)
	}
}

type serviceDesc struct {
	GoPackageName      string
	ServiceType        string // Greeter
	ServiceName        string // helloworld.Greeter
	Metadata           string // api/helloworld/helloworld.proto
	Methods            []*methodDesc
	MethodSets         map[string]*methodDesc
	Comment            string
	Topic              string // topic
	Queue              string // service级别的queue
	HasServiceQueue    bool   // service级别的queue
	HasMethodQueue     bool
	HasMethodReqRespId bool
	HasMethodSequence  bool
	Async              bool
}

type methodDesc struct {
	// method
	MethodName   string
	Request      string
	Reply        string
	Comment      string
	Publish      bool   // false表示request(需要返回值)，true表示广播(不需要返回值)
	Queue        string // 方法级别的queue
	HasQueue     bool
	ReqId        int32
	RespId       int32
	HasReqRespId bool
	Async        bool
	Sequence     bool
	HasSequence  bool
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.MethodName] = m
	}
	buf := new(bytes.Buffer)
	if err := sTpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
