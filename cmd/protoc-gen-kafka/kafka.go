package main

import (
	"fmt"
	"strings"

	"github.com/liuwangchen/toy/transport/rpc/kafkarpc"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contextPackage        = protogen.GoImportPath("context")
	transportKafkaPackage = protogen.GoImportPath("github.com/liuwangchen/toy/transport/rpc/kafkarpc")
	middlewarePackage     = protogen.GoImportPath("github.com/liuwangchen/toy/transport/middleware")
	emptyPackage          = protogen.GoImportPath("github.com/golang/protobuf/ptypes/empty")
)

// generateFile generates a kafka.pb.go file containing goctopus errors definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 || !isKafkaRule(file.Services) {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_kafka.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-kafka DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// protoc-gen-go-kafka %s", release))
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
	g.P("var _ = new(", emptyPackage.Ident("Empty"), ")")
	g.P("var _ = ", middlewarePackage.Ident("Chain()"))
	g.P("const _ = ", transportKafkaPackage.Ident("SupportPackageIsVersion1"))
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
	// kafka Server.
	sd := &serviceDesc{
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Metadata:    file.Desc.Path(),
	}
	dynamicTopic, ok := proto.GetExtension(service.Desc.Options(), kafkarpc.E_DynamicTopic).(string)
	if ok {
		sd.DynamicTopic = dynamicTopic
	}

	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		if method.Output.GoIdent.GoName != "Empty" {
			continue
		}

		sd.Methods = append(sd.Methods, buildMethodDesc(g, method))
	}
	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func isKafkaRule(services []*protogen.Service) bool {
	return true
}

func buildMethodDesc(g *protogen.GeneratedFile, m *protogen.Method) *methodDesc {
	method := &methodDesc{
		Name:    m.GoName,
		Request: g.QualifiedGoIdent(m.Input.GoIdent),
		Reply:   g.QualifiedGoIdent(m.Output.GoIdent),
	}
	autoAck, ok := proto.GetExtension(m.Desc.Options(), kafkarpc.E_AutoAck).(bool)
	if ok {
		method.AutoAck = autoAck
	}
	queue, ok := proto.GetExtension(m.Desc.Options(), kafkarpc.E_Queue).(string)
	if ok {
		method.Queue = queue
	}
	return method
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
