package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

var sTpl *template.Template

//go:embed tmpl.gohtml
var serviceTmpl string

func init() {
	var err error
	sTpl, err = template.New("tmpl").Funcs(sprig.TxtFuncMap()).Parse(serviceTmpl)
	if err != nil {
		panic(err)
	}
}

type serviceDesc struct {
	ServiceType  string // Greeter
	ServiceName  string // helloworld.Greeter
	Metadata     string // api/helloworld/helloworld.proto
	Methods      []*methodDesc
	MethodSets   map[string]*methodDesc
	DynamicTopic string // 动态topic
}

type methodDesc struct {
	// method
	Name    string
	Request string
	Reply   string
	AutoAck bool
	Queue   string
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	if err := sTpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
