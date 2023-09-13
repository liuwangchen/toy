package main

import (
	"bytes"
	"strings"
	"text/template"
)

var httpTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
{{$serviceAsync := .Async}}
{{$serviceWrapperName := print .ServiceType "HttpServiceWrapper"}}
type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{- if .IsStreamingServer}}
	{{.Name}}Stream(context.Context, *{{.Request}}, *{{.Name}}ServerStream) error
	{{- else }}
		{{- if .IsPublish}}
	{{.Name}}(context.Context, *{{.Request}}) error
		{{- else }}
			{{- if .Async }}
	{{ .Name }}(ctx context.Context, req *{{ .Request }}, cb func(*{{ .Reply }}, error))
			{{- else }}
	{{ .Name }}(ctx context.Context, req *{{ .Request }}) (*{{ .Reply }}, error)
			{{- end}}
		{{- end}}
	{{- end}}
{{- end}}
}

{{- if .Async}}
// {{ $serviceWrapperName }} DO NOT USE
type {{ $serviceWrapperName }} struct {
	doer async.IAsync
	s    {{.ServiceType}}HTTPServer
}
{{- range .Methods }}
// {{ .Name }} DO NOT USE
	{{- if eq .IsPublish true }}
func (s *{{ $serviceWrapperName }}){{ .Name }}(ctx context.Context, req *{{ .Request }}) error {
	_, err := s.doer.Do(ctx, func() (interface{}, error) {
	err := s.s.{{ .Name }}(ctx , req)
		return nil, err
	})
	return err
}
	{{- else }}
func (s *{{ $serviceWrapperName }}){{ .Name }}(ctx context.Context, req *{{ .Request }})(*{{ .Reply }}, error) {
	{{- if .Async }}
	f := func(cb func(interface{}, error)) {
		s.s.{{ .Name }}(ctx, req, func(r *{{ .Reply }}, e error) {
			cb(r,e)
		})
	}
	temp, err := s.doer.AsyncDo(ctx, f)
	if temp == nil {
		return nil, err
	}
	return temp.(*{{ .Reply }}), err
	{{- else }}
	temp, err := s.doer.Do(ctx, func() (interface{}, error) {
		return s.s.{{ .Name }}(ctx, req)
	})
	if temp == nil {
		return nil, err
	}
	return temp.(*{{ .Reply }}), err
	{{- end}}
}
	{{- end }}
{{- end }}

func RegisterAsync{{.ServiceType}}HTTPServer(conn *httprpc.ServerConn, doer async.IAsync, srv {{.ServiceType}}HTTPServer, opts ...httprpc.ServiceOption) {
	r := conn.Route("/")
	opt := new(httprpc.ServiceOpt)
	for _, o := range opts {
		o(opt)
	}
	ss := &{{ $serviceWrapperName }}{
		doer: doer,
		s:    srv,
	}
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(ss, opt))
	{{- end}}
}
{{- else }}
func Register{{.ServiceType}}HTTPServer(conn *httprpc.ServerConn, srv {{.ServiceType}}HTTPServer, opts ...httprpc.ServiceOption) {
	r := conn.Route("/")
	opt := new(httprpc.ServiceOpt)
	for _, o := range opts {
		o(opt)
	}
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv, opt))
	{{- end}}
}
{{- end}}

{{range .Methods}}
	{{- if .IsStreamingServer}}
type {{.Name}}ServerStream struct {
	httprpc.IServerStream
}

func New{{.Name}}ServerStream(stream httprpc.IServerStream) *{{.Name}}ServerStream {
	return &{{.Name}}ServerStream{IServerStream: stream}
}

func (st *{{.Name}}ServerStream) Send(v *{{.Reply}}) error {
	return st.SendMsg(v)
}
	{{- end}}

func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler({{- if $serviceAsync}}srv *{{ $serviceWrapperName }}{{- else}}srv {{$svrType}}HTTPServer{{- end}}, opt *httprpc.ServiceOpt) func(ctx httprpc.Context) error {
	return func(ctx httprpc.Context) error {
		var in {{.Request}}
		{{- if .HasBody}}
		if err := ctx.Bind(&in{{.Body}}); err != nil {
			return err
		}

		{{- if not (eq .Body "")}}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		{{- end}}
		{{- else}}
		if err := ctx.BindQuery(&in{{.Body}}); err != nil {
			return err
		}
		{{- end}}
		{{- if .HasVars}}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		{{- end}}
		httprpc.SetOperation(ctx,"/{{$svrName}}/{{.Name}}")
		{{- if .IsStreamingServer}}
		stream := New{{.Name}}ServerStream(httprpc.NewHttpServerStream(ctx.Response(), httprpc.CodecForRequest(ctx.Request(), "Accept")))
		h := ctx.Middleware(middleware.Chain(opt.Mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			err := srv.{{.Name}}Stream(ctx, req.(*{{.Request}}), stream)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}))
		_, err := h(ctx, &in)
		return err
		{{- else}}
		h := ctx.Middleware(middleware.Chain(opt.Mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
		{{- if .IsPublish}}
			return nil, srv.{{.Name}}(ctx, req.(*{{.Request}}))
		{{- else}}
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		{{- end}}
		}))
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		{{- if .IsPublish}}
		reply := out
		{{- else}}
		reply := out.(*{{.Reply}})
		{{- end}}
		return ctx.Result(200, reply{{.ResponseBody}})
		{{- end}}
	}
}
{{end}}

type {{.ServiceType}}HTTPClient interface {
{{- range .MethodSets}}
	{{- if .IsStreamingServer}}
	{{.Name}}Stream(ctx context.Context, req *{{.Request}}, opts ...httprpc.CallOption) (st *{{.Name}}ClientStream, err error)
	{{- else }}
	{{- if .IsPublish}}
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts ...httprpc.CallOption) error
	{{- else}}
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts ...httprpc.CallOption) (rsp *{{.Reply}}, err error)
	{{- end}}
	{{- end }}
{{- end}}
}
	
type {{.ServiceType}}HTTPClientImpl struct{
	cc *httprpc.ClientConn
}
	
func New{{.ServiceType}}HTTPClient (conn *httprpc.ClientConn) {{.ServiceType}}HTTPClient {
	return &{{.ServiceType}}HTTPClientImpl{conn}
}

{{range .MethodSets}}
	{{- if .IsStreamingServer}}
type {{.Name}}ClientStream struct {
	recver httprpc.IClientStream
}

func New{{.Name}}ClientStream(recver httprpc.IClientStream) *{{.Name}}ClientStream {
	return &{{.Name}}ClientStream{recver: recver}
}

func (s *{{.Name}}ClientStream) Recv() (*{{.Reply}}, error) {
	v := new({{.Reply}})
	err := s.recver.RecvMsg(v)
	if v == nil {
		return nil, err
	}
	return v, err
}

func (c *{{$svrType}}HTTPClientImpl) {{.Name}}Stream(ctx context.Context, in *{{.Request}}, opts ...httprpc.CallOption) (*{{.Name}}ClientStream, error) {
	pattern := "{{.Path}}"
	path := binding.EncodeURL(pattern, in, {{not .HasBody}})
	opts = append(opts, httprpc.Operation("/{{$svrName}}/{{.Name}}"))
	opts = append(opts, httprpc.PathTemplate(pattern))
	{{if .HasBody -}}
	stream, err := c.cc.Stream(ctx, "{{.Method}}", path, in{{.Body}}, opts...)
	{{else -}}
	stream, err := c.cc.Stream(ctx, "{{.Method}}", path, nil, opts...)
	{{end -}}
	if err != nil {
		return nil, err
	}
	return New{{.Name}}ClientStream(stream), err
}
	{{- else }}
	{{- if .IsPublish}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts ...httprpc.CallOption) error {
	var out empty.Empty
	pattern := "{{.Path}}"
	path := binding.EncodeURL(pattern, in, {{not .HasBody}})
	opts = append(opts, httprpc.Operation("/{{$svrName}}/{{.Name}}"))
	opts = append(opts, httprpc.PathTemplate(pattern))
	{{if .HasBody -}}
	err := c.cc.Invoke(ctx, "{{.Method}}", path, in{{.Body}}, &out{{.ResponseBody}}, opts...)
	{{else -}} 
	err := c.cc.Invoke(ctx, "{{.Method}}", path, nil, &out{{.ResponseBody}}, opts...)
	{{end -}}
	if err != nil {
		return err
	}
	return err
}
	{{- else }}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts ...httprpc.CallOption) (*{{.Reply}}, error) {
	var out {{.Reply}}
	pattern := "{{.Path}}"
	path := binding.EncodeURL(pattern, in, {{not .HasBody}})
	opts = append(opts, httprpc.Operation("/{{$svrName}}/{{.Name}}"))
	opts = append(opts, httprpc.PathTemplate(pattern))
	{{if .HasBody -}}
	err := c.cc.Invoke(ctx, "{{.Method}}", path, in{{.Body}}, &out{{.ResponseBody}}, opts...)
	{{else -}} 
	err := c.cc.Invoke(ctx, "{{.Method}}", path, nil, &out{{.ResponseBody}}, opts...)
	{{end -}}
	if err != nil {
		return nil, err
	}
	return &out, err
}
	{{- end }}
	{{- end }}
{{end}}
`

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
	Async       bool
}

type methodDesc struct {
	// method
	Name    string
	Num     int
	Request string
	Reply   string
	// http_rule
	Path              string
	Method            string
	HasVars           bool
	HasBody           bool
	Body              string
	ResponseBody      string
	IsStreamingServer bool
	IsPublish         bool
	Async             bool
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
