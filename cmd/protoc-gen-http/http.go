package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/liuwangchen/toy/transport/rpc"
	"github.com/liuwangchen/toy/transport/rpc/httprpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contextPackage       = protogen.GoImportPath("context")
	reflectPackage       = protogen.GoImportPath("reflect")
	transportHTTPPackage = protogen.GoImportPath("github.com/liuwangchen/toy/transport/rpc/httprpc")
	bindingPackage       = protogen.GoImportPath("github.com/liuwangchen/toy/transport/rpc/httprpc/binding")
	middlewarePackage    = protogen.GoImportPath("github.com/liuwangchen/toy/transport/middleware")
	asyncPackage         = protogen.GoImportPath("github.com/liuwangchen/toy/pkg/async")
)

var methodSets = make(map[string]int)

// generateFile generates a _http.pb.go file containing goctopus errors definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 || !hasHTTPRule(file.Services) {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_http.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-http. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// protoc-gen-go-http %s", release))
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(gen, file, g)
	return g
}

// generateFileContent generates the goctopus errors definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the goctopus package it is being compiled against.")
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	g.P("var _ = ", bindingPackage.Ident("EncodeURL"))
	g.P("var _ = ", reflectPackage.Ident("TypeOf"))
	g.P("const _ = ", transportHTTPPackage.Ident("SupportPackageIsVersion1"))
	g.P("var _ = ", middlewarePackage.Ident("Chain()"))
	g.P("var _ = new(", asyncPackage.Ident("Async"), ")")
	g.P()

	for _, service := range file.Services {
		genService(gen, file, g, service)
	}
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &serviceDesc{
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Metadata:    file.Desc.Path(),
	}
	serviceAsync, ok := proto.GetExtension(service.Desc.Options(), rpc.E_ServiceAsync).(bool)
	if ok {
		sd.Async = serviceAsync
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() {
			continue
		}

		rule, ok := proto.GetExtension(method.Desc.Options(), httprpc.E_Rule).(*httprpc.HttpRule)
		if rule != nil && ok {
			for _, bind := range rule.AdditionalBindings {
				methodDesc := buildHTTPRule(g, method, bind, method.Desc.IsStreamingServer())
				if methodDesc.Async {
					serviceAsync = true
				}
				sd.Methods = append(sd.Methods, methodDesc)
			}
			methodDesc := buildHTTPRule(g, method, rule, method.Desc.IsStreamingServer())
			if methodDesc.Async {
				serviceAsync = true
			}
			sd.Methods = append(sd.Methods, methodDesc)
		} else {
			path := fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())
			methodDesc := buildMethodDesc(g, method, "POST", path, method.Desc.IsStreamingServer())
			methodDesc.HasBody = true
			if methodDesc.Async {
				serviceAsync = true
			}
			sd.Methods = append(sd.Methods, methodDesc)
		}
	}
	sd.Async = serviceAsync
	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func hasHTTPRule(services []*protogen.Service) bool {
	for _, service := range services {
		for _, method := range service.Methods {
			if method.Desc.IsStreamingClient() {
				continue
			}
			return true
		}
	}
	return false
}

func buildHTTPRule(g *protogen.GeneratedFile, m *protogen.Method, rule *httprpc.HttpRule, isStreamingServer bool) *methodDesc {
	var (
		path         string
		method       string
		body         string
		responseBody string
	)

	switch pattern := rule.Pattern.(type) {
	case *httprpc.HttpRule_Get:
		path = pattern.Get
		method = "GET"
	case *httprpc.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
	case *httprpc.HttpRule_Post:
		path = pattern.Post
		method = "POST"
	case *httprpc.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
	case *httprpc.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
	case *httprpc.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	body = rule.Body
	responseBody = rule.ResponseBody
	md := buildMethodDesc(g, m, method, path, isStreamingServer)
	if method == "GET" || method == "DELETE" {
		if body != "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s body should not be declared.\n", method, path)
		}
	} else {
		if body == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s does not declare a body.\n", method, path)
		}
	}
	if body == "*" {
		md.HasBody = true
		md.Body = ""
	} else if body != "" {
		md.HasBody = true
		md.Body = "." + camelCaseVars(body)
	} else {
		md.HasBody = false
	}
	if responseBody == "*" {
		md.ResponseBody = ""
	} else if responseBody != "" {
		md.ResponseBody = "." + camelCaseVars(responseBody)
	}
	return md
}

func buildMethodDesc(g *protogen.GeneratedFile, m *protogen.Method, method, path string, isStreamingServer bool) *methodDesc {
	defer func() { methodSets[m.GoName]++ }()

	vars := buildPathVars(path)
	fields := m.Input.Desc.Fields()

	for v, s := range vars {
		if s != nil {
			path = replacePath(v, *s, path)
		}
		for _, field := range strings.Split(v, ".") {
			if strings.TrimSpace(field) == "" {
				continue
			}
			if strings.Contains(field, ":") {
				field = strings.Split(field, ":")[0]
			}
			fd := fields.ByName(protoreflect.Name(field))
			if fd == nil {
				fmt.Fprintf(os.Stderr, "\u001B[31mERROR\u001B[m: The corresponding field '%s' declaration in message could not be found in '%s'\n", v, path)
				os.Exit(2)
			}
			if fd.IsMap() {
				fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a map.\n", v)
			} else if fd.IsList() {
				fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a list.\n", v)
			} else if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
				fields = fd.Message().Fields()
			}
		}
	}

	// 是否是推送
	isPublish := m.Desc.Output().FullName() == "google.protobuf.Empty"

	async, _ := proto.GetExtension(m.Desc.Options(), rpc.E_Async).(bool)

	return &methodDesc{
		Name:              m.GoName,
		Num:               methodSets[m.GoName],
		Request:           g.QualifiedGoIdent(m.Input.GoIdent),
		Reply:             g.QualifiedGoIdent(m.Output.GoIdent),
		Path:              path,
		Method:            method,
		HasVars:           len(vars) > 0,
		IsStreamingServer: isStreamingServer,
		IsPublish:         isPublish,
		Async:             async,
	}
}

func buildPathVars(path string) (res map[string]*string) {
	if strings.HasSuffix(path, "/") {
		fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: Path %s should not end with \"/\" \n", path)
	}
	res = make(map[string]*string)
	pattern := regexp.MustCompile(`(?i){([a-z\.0-9_\s]*)=?([^{}]*)}`)
	matches := pattern.FindAllStringSubmatch(path, -1)
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if len(name) > 1 && len(m[2]) > 0 {
			res[name] = &m[2]
		} else {
			res[name] = nil
		}
	}
	return
}

func replacePath(name string, value string, path string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i){([\s]*%s[\s]*)=?([^{}]*)}`, name))
	idx := pattern.FindStringIndex(path)
	if len(idx) > 0 {
		path = fmt.Sprintf("%s{%s:%s}%s",
			path[:idx[0]], // The start of the match
			name,
			strings.ReplaceAll(value, "*", ".*"),
			path[idx[1]:],
		)
	}
	return path
}

func camelCaseVars(s string) string {
	subs := strings.Split(s, ".")
	vars := make([]string, 0, len(subs))
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")
}

// camelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercases names, it's extremely unlikely to have two fields
// with different capitalizations.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

const deprecationComment = "// Deprecated: Do not use."
