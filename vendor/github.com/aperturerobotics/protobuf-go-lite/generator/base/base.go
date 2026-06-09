// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generator_base

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"github.com/aperturerobotics/protobuf-go-lite/generator/fieldsem"
	"github.com/aperturerobotics/protobuf-go-lite/internal/encoding/tag"
	"github.com/aperturerobotics/protobuf-go-lite/internal/genid"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Standard library dependencies.
const (
	base64Package  = protogen.GoImportPath("encoding/base64")
	mathPackage    = protogen.GoImportPath("math")
	reflectPackage = protogen.GoImportPath("reflect")
	sortPackage    = protogen.GoImportPath("sort")
	stringsPackage = protogen.GoImportPath("strings")
	strconvPackage = protogen.GoImportPath("strconv")
	syncPackage    = protogen.GoImportPath("sync")
	timePackage    = protogen.GoImportPath("time")
	utf8Package    = protogen.GoImportPath("unicode/utf8")
	errorsPackage  = protogen.GoImportPath("errors")
	fmtPackage     = protogen.GoImportPath("fmt")
)

// GenerateFile generates the contents of a .pb.go file.
func GenerateFile(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	f := newFileInfo(file)

	genStandaloneComments(g, f, int32(genid.FileDescriptorProto_Syntax_field_number))
	genStandaloneComments(g, f, int32(genid.FileDescriptorProto_Package_field_number))

	for i, imps := 0, f.Desc.Imports(); i < imps.Len(); i++ {
		genImport(gen, g, f, imps.Get(i))
	}
	for _, enum := range f.allEnums {
		genEnum(g, enum)
	}
	for _, message := range f.allMessages {
		genMessage(g, f, message)
	}

	// NOTE: extensions not supported
	// NOTE: reflect not supported
}

// genStandaloneComments prints all leading comments for a FileDescriptorProto
// location identified by the field number n.
func genStandaloneComments(g *protogen.GeneratedFile, f *fileInfo, n int32) {
	loc := f.Desc.SourceLocations().ByPath(protoreflect.SourcePath{n})
	for _, s := range loc.LeadingDetachedComments {
		g.P(protogen.Comments(s))
		g.P()
	}
	if s := loc.LeadingComments; s != "" {
		g.P(protogen.Comments(s))
		g.P()
	}
}

func genImport(gen *protogen.Plugin, g *protogen.GeneratedFile, f *fileInfo, imp protoreflect.FileImport) {
	impFile, ok := gen.FilesByPath[imp.Path()]
	if !ok {
		return
	}
	if impFile.GoImportPath == f.GoImportPath {
		// Don't generate imports or aliases for types in the same Go package.
		return
	}
	// Generate imports for all non-weak dependencies, even if they are not
	// referenced, because other code and tools depend on having the
	// full transitive closure of protocol buffer types in the binary.
	if !imp.IsWeak {
		g.Import(impFile.GoImportPath)
	}
	if !imp.IsPublic {
		return
	}

	// Generate public imports by generating the imported file, parsing it,
	// and extracting every symbol that should receive a forwarding declaration.
	impGen := gen.NewGeneratedFile(impFile.GeneratedFilenamePrefix+".pb.go", impFile.GoImportPath)
	GenerateFile(gen, impFile, impGen)
	impGen.Skip()
	b, err := impGen.Content()
	if err != nil {
		gen.Error(err)
		return
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "", b, parser.ParseComments)
	if err != nil {
		gen.Error(err)
		return
	}
	genForward := func(tok token.Token, name string, expr ast.Expr) {
		// Don't import unexported symbols.
		r, _ := utf8.DecodeRuneInString(name)
		if !unicode.IsUpper(r) {
			return
		}
		// Don't import the FileDescriptor.
		if name == impFile.GoDescriptorIdent.GoName {
			return
		}
		// Don't import decls referencing a symbol defined in another package.
		// i.e., don't import decls which are themselves public imports:
		//
		//	type T = somepackage.T
		if _, ok := expr.(*ast.SelectorExpr); ok {
			return
		}
		g.P(tok, " ", name, " = ", impFile.GoImportPath.Ident(name))
	}
	g.P("// Symbols defined in public import of ", imp.Path(), ".")
	g.P()
	for _, decl := range astFile.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					genForward(decl.Tok, spec.Name.Name, spec.Type)
				case *ast.ValueSpec:
					for i, name := range spec.Names {
						var expr ast.Expr
						if i < len(spec.Values) {
							expr = spec.Values[i]
						}
						genForward(decl.Tok, name.Name, expr)
					}
				case *ast.ImportSpec:
				default:
					panic(fmt.Sprintf("can't generate forward for spec type %T", spec))
				}
			}
		}
	}
	g.P()
}

func genEnum(g *protogen.GeneratedFile, e *enumInfo) {
	// Enum type declaration.
	g.Annotate(e.GoIdent.GoName, e.Location)
	leadingComments := appendDeprecationSuffix(e.Comments.Leading,
		e.Desc.ParentFile(),
		e.Desc.Options().(*descriptorpb.EnumOptions).GetDeprecated())
	g.P(leadingComments,
		"type ", e.GoIdent, " int32")

	// Enum value constants.
	g.P("const (")
	for _, value := range e.Values {
		g.Annotate(value.GoIdent.GoName, value.Location)
		leadingComments := appendDeprecationSuffix(value.Comments.Leading,
			value.Desc.ParentFile(),
			value.Desc.Options().(*descriptorpb.EnumValueOptions).GetDeprecated())
		g.P(leadingComments,
			value.GoIdent, " ", e.GoIdent, " = ", value.Desc.Number(),
			trailingComment(value.Comments.Trailing))
	}
	g.P(")")
	g.P()

	// Enum value maps.
	g.P("// Enum value maps for ", e.GoIdent, ".")
	g.P("var (")
	g.P(e.GoIdent.GoName+"_name", " = map[int32]string{")
	for _, value := range e.Values {
		duplicate := ""
		if value.Desc != e.Desc.Values().ByNumber(value.Desc.Number()) {
			duplicate = "// Duplicate value: "
		}
		g.P(duplicate, value.Desc.Number(), ": ", strconv.Quote(string(value.Desc.Name())), ",")
	}
	g.P("}")
	g.P(e.GoIdent.GoName+"_value", " = map[string]int32{")
	for _, value := range e.Values {
		g.P(strconv.Quote(string(value.Desc.Name())), ": ", value.Desc.Number(), ",")
	}
	g.P("}")
	g.P(")")
	g.P()

	// Enum method.
	//
	// NOTE: A pointer value is needed to represent presence in proto2.
	// Since a proto2 message can reference a proto3 enum, it is useful to
	// always generate this method (even on proto3 enums) to support that case.
	g.P("func (x ", e.GoIdent, ") Enum() *", e.GoIdent, " {")
	g.P("p := new(", e.GoIdent, ")")
	g.P("*p = x")
	g.P("return p")
	g.P("}")
	g.P()

	// String method.
	//
	// Returns the enum value as a string, either as the name if the number is
	// resolvable, or the number formatted as a string.
	g.P("func (x ", e.GoIdent, ") String() string {")
	g.P("name, valid := ", e.GoIdent.GoName+"_name[int32(x)]")
	g.P("if valid { return name }")
	g.P("return ", strconvPackage.Ident("Itoa"), "(int(x))")
	g.P("}")
	g.P()
}

func genMessage(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	/*
		if m.Desc.IsMapEntry() {
			return
		}
	*/

	// Message type declaration.
	g.Annotate(m.GoIdent.GoName, m.Location)
	leadingComments := appendDeprecationSuffix(m.Comments.Leading,
		m.Desc.ParentFile(),
		m.Desc.Options().(*descriptorpb.MessageOptions).GetDeprecated())
	g.P(leadingComments,
		"type ", m.GoIdent, " struct {")
	genMessageFields(g, f, m)
	g.P("}")
	g.P()

	genMessageKnownFunctions(g, m)
	genMessageDefaultDecls(g, f, m)
	genMessageMethods(g, f, m)
	genMessageOneofWrapperTypes(g, f, m)
}

func genMessageFields(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	sf := f.allMessageFieldsByPtr[m]
	genMessageInternalFields(g, sf)
	for _, field := range m.Fields {
		genMessageField(g, f, m, field, sf)
	}
}

func genMessageInternalFields(g *protogen.GeneratedFile, sf *structFields) {
	g.P(genid.UnknownFields_goname, " ", "[]byte") // NOTE: this is inlined version of protoimpl.UnknownFields
	sf.append(genid.UnknownFields_goname)
	// NOTE: extensions, weak fields not supported.
}

func genMessageField(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, field *protogen.Field, sf *structFields) {
	if oneof := field.Oneof; oneof != nil && !oneof.Desc.IsSynthetic() {
		// It would be a bit simpler to iterate over the oneofs below,
		// but generating the field here keeps the contents of the Go
		// struct in the same order as the contents of the source
		// .proto file.
		if oneof.Fields[0] != field {
			return // only generate for first appearance
		}

		tags := structTags{
			{"protobuf_oneof", string(oneof.Desc.Name())},
		}
		if m.isTracked {
			tags = append(tags, gotrackTags...)
		}

		g.Annotate(m.GoIdent.GoName+"."+oneof.GoName, oneof.Location)
		leadingComments := oneof.Comments.Leading
		if leadingComments != "" {
			leadingComments += "\n"
		}
		ss := []string{fmt.Sprintf(" Types that are assignable to %s:\n", oneof.GoName)}
		for _, field := range oneof.Fields {
			ss = append(ss, "\t*"+field.GoIdent.GoName+"\n")
		}
		leadingComments += protogen.Comments(strings.Join(ss, ""))
		g.P(leadingComments,
			oneof.GoName, " ", oneofInterfaceName(oneof), tags)
		sf.append(oneof.GoName)
		return
	}
	goType, pointer := fieldGoType(g, f, field)
	if pointer {
		goType = "*" + goType
	}
	tags := structTags{
		{"protobuf", fieldProtobufTagValue(field)},
		{"json", fieldJSONTagValue(field)},
	}
	if field.Desc.IsMap() {
		key := field.Message.Fields[0]
		val := field.Message.Fields[1]
		tags = append(tags, structTags{
			{"protobuf_key", fieldProtobufTagValue(key)},
			{"protobuf_val", fieldProtobufTagValue(val)},
		}...)
	}
	if m.isTracked {
		tags = append(tags, gotrackTags...)
	}

	name := field.GoName
	if field.Desc.IsWeak() {
		name = genid.WeakFieldPrefix_goname + name
	}
	g.Annotate(m.GoIdent.GoName+"."+name, field.Location)
	leadingComments := appendDeprecationSuffix(field.Comments.Leading,
		field.Desc.ParentFile(),
		field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
	g.P(leadingComments,
		name, " ", goType, tags,
		trailingComment(field.Comments.Trailing))
	sf.append(field.GoName)
}

// genMessageDefaultDecls generates consts and vars holding the default
// values of fields.
func genMessageDefaultDecls(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	var consts, vars []string
	for _, field := range m.Fields {
		if !field.Desc.HasDefault() {
			continue
		}
		name := "Default_" + m.GoIdent.GoName + "_" + field.GoName
		goType, _ := fieldGoType(g, f, field)
		defVal := field.Desc.Default()
		switch field.Desc.Kind() {
		case protoreflect.StringKind:
			consts = append(consts, fmt.Sprintf("%s = %s(%q)", name, goType, defVal.String()))
		case protoreflect.BytesKind:
			vars = append(vars, fmt.Sprintf("%s = %s(%q)", name, goType, defVal.Bytes()))
		case protoreflect.EnumKind:
			idx := field.Desc.DefaultEnumValue().Index()
			val := field.Enum.Values[idx]
			if val.GoIdent.GoImportPath == f.GoImportPath {
				consts = append(consts, fmt.Sprintf("%s = %s", name, g.QualifiedGoIdent(val.GoIdent)))
			} else {
				// If the enum value is declared in a different Go package,
				// reference it by number since the name may not be correct.
				// See https://github.com/golang/protobuf/issues/513.
				consts = append(consts, fmt.Sprintf("%s = %s(%d) // %s",
					name, g.QualifiedGoIdent(field.Enum.GoIdent), val.Desc.Number(), g.QualifiedGoIdent(val.GoIdent)))
			}
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			if f := defVal.Float(); math.IsNaN(f) || math.IsInf(f, 0) {
				var fn, arg string
				switch f := defVal.Float(); {
				case math.IsInf(f, -1):
					fn, arg = g.QualifiedGoIdent(mathPackage.Ident("Inf")), "-1"
				case math.IsInf(f, +1):
					fn, arg = g.QualifiedGoIdent(mathPackage.Ident("Inf")), "+1"
				case math.IsNaN(f):
					fn, arg = g.QualifiedGoIdent(mathPackage.Ident("NaN")), ""
				}
				vars = append(vars, fmt.Sprintf("%s = %s(%s(%s))", name, goType, fn, arg))
			} else {
				consts = append(consts, fmt.Sprintf("%s = %s(%v)", name, goType, f))
			}
		default:
			consts = append(consts, fmt.Sprintf("%s = %s(%v)", name, goType, defVal.Interface()))
		}
	}
	if len(consts) > 0 {
		g.P("// Default values for ", m.GoIdent, " fields.")
		g.P("const (")
		for _, s := range consts {
			g.P(s)
		}
		g.P(")")
	}
	if len(vars) > 0 {
		g.P("// Default values for ", m.GoIdent, " fields.")
		g.P("var (")
		for _, s := range vars {
			g.P(s)
		}
		g.P(")")
	}
	g.P()
}

func genMessageMethods(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	genMessageBaseMethods(g, f, m)
	genMessageGetterMethods(g, f, m)
}

func genMessageBaseMethods(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	// Reset method.
	g.P("func (x *", m.GoIdent, ") Reset() {")
	g.P("*x = ", m.GoIdent, "{}")
	g.P("}")
	g.P()

	// TODO String method?.

	// ProtoMessage method.
	g.P("func (*", m.GoIdent, ") ProtoMessage() {}")
	g.P()
}

func genMessageGetterMethods(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	for _, field := range m.Fields {
		genNoInterfacePragma(g, m.isTracked)

		// Getter for parent oneof.
		if oneof := field.Oneof; oneof != nil && oneof.Fields[0] == field && !oneof.Desc.IsSynthetic() {
			g.Annotate(m.GoIdent.GoName+".Get"+oneof.GoName, oneof.Location)
			g.P("func (m *", m.GoIdent.GoName, ") Get", oneof.GoName, "() ", oneofInterfaceName(oneof), " {")
			g.P("if m != nil {")
			g.P("return m.", oneof.GoName)
			g.P("}")
			g.P("return nil")
			g.P("}")
			g.P()
		}

		// Getter for message field.
		goType, pointer := fieldGoType(g, f, field)
		defaultValue := fieldDefaultValue(g, f, m, field)
		g.Annotate(m.GoIdent.GoName+".Get"+field.GoName, field.Location)
		leadingComments := appendDeprecationSuffix("",
			field.Desc.ParentFile(),
			field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
		switch {
		case field.Desc.IsWeak():
		// NOTE: weak fields not supported
		case field.Oneof != nil && !field.Oneof.Desc.IsSynthetic():
			g.P(leadingComments, "func (x *", m.GoIdent, ") Get", field.GoName, "() ", goType, " {")
			g.P("if x, ok := x.Get", field.Oneof.GoName, "().(*", field.GoIdent, "); ok {")
			g.P("return x.", field.GoName)
			g.P("}")
			g.P("return ", defaultValue)
			g.P("}")
		default:
			g.P(leadingComments, "func (x *", m.GoIdent, ") Get", field.GoName, "() ", goType, " {")
			if !field.Desc.HasPresence() || defaultValue == "nil" {
				g.P("if x != nil {")
			} else {
				g.P("if x != nil && x.", field.GoName, " != nil {")
			}
			star := ""
			if pointer {
				star = "*"
			}
			g.P("return ", star, " x.", field.GoName)
			g.P("}")
			g.P("return ", defaultValue)
			g.P("}")
		}
		g.P()
	}
}

// fieldGoType returns the Go type used for a field.
//
// If it returns pointer=true, the struct field is a pointer to the type.
func fieldGoType(g *protogen.GeneratedFile, f *fileInfo, field *protogen.Field) (goType string, pointer bool) {
	sem := fieldsem.Resolve(g, field)
	return sem.Type, sem.Pointer
}

func fieldProtobufTagValue(field *protogen.Field) string {
	var enumName string
	//if field.Desc.Kind() == protoreflect.EnumKind {
	// enumName = protoimpl.X.LegacyEnumName(field.Enum.Desc)
	//}
	return tag.Marshal(field.Desc, enumName)
}

func fieldDefaultValue(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, field *protogen.Field) string {
	if field.Desc.IsList() {
		return "nil"
	}
	if field.Desc.HasDefault() {
		defVarName := "Default_" + m.GoIdent.GoName + "_" + field.GoName
		if field.Desc.Kind() == protoreflect.BytesKind {
			return "append([]byte(nil), " + defVarName + "...)"
		}
		return defVarName
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.StringKind:
		return `""`
	case protoreflect.MessageKind, protoreflect.GroupKind, protoreflect.BytesKind:
		return "nil"
	case protoreflect.EnumKind:
		val := field.Enum.Values[0]
		if val.GoIdent.GoImportPath == f.GoImportPath {
			return g.QualifiedGoIdent(val.GoIdent)
		} else {
			// If the enum value is declared in a different Go package,
			// reference it by number since the name may not be correct.
			// See https://github.com/golang/protobuf/issues/513.
			return g.QualifiedGoIdent(field.Enum.GoIdent) + "(" + strconv.FormatInt(int64(val.Desc.Number()), 10) + ")"
		}
	default:
		return "0"
	}
}

func fieldJSONTagValue(field *protogen.Field) string {
	return field.Desc.JSONName() + ",omitempty"
}

// genMessageOneofWrapperTypes generates the oneof wrapper types and
// associates the types with the parent message type.
func genMessageOneofWrapperTypes(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	for _, oneof := range m.Oneofs {
		if oneof.Desc.IsSynthetic() {
			continue
		}
		ifName := oneofInterfaceName(oneof)
		g.P("type ", ifName, " interface {")
		g.P(ifName, "()")
		g.P("}")
		g.P()
		for _, field := range oneof.Fields {
			g.Annotate(field.GoIdent.GoName, field.Location)
			g.Annotate(field.GoIdent.GoName+"."+field.GoName, field.Location)
			g.P("type ", field.GoIdent, " struct {")
			goType, _ := fieldGoType(g, f, field)
			tags := structTags{
				{"protobuf", fieldProtobufTagValue(field)},
			}
			if m.isTracked {
				tags = append(tags, gotrackTags...)
			}
			leadingComments := appendDeprecationSuffix(field.Comments.Leading,
				field.Desc.ParentFile(),
				field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
			g.P(leadingComments,
				field.GoName, " ", goType, tags,
				trailingComment(field.Comments.Trailing))
			g.P("}")
			g.P()
		}
		for _, field := range oneof.Fields {
			g.P("func (*", field.GoIdent, ") ", ifName, "() {}")
			g.P()
		}
	}
}

// oneofInterfaceName returns the name of the interface type implemented by
// the oneof field value types.
func oneofInterfaceName(oneof *protogen.Oneof) string {
	return "is" + oneof.GoIdent.GoName
}

// genNoInterfacePragma generates a standalone "nointerface" pragma to
// decorate methods with field-tracking support.
func genNoInterfacePragma(g *protogen.GeneratedFile, tracked bool) {
	if tracked {
		g.P("//go:nointerface")
		g.P()
	}
}

var gotrackTags = structTags{{"go", "track"}}

// structTags is a data structure for build idiomatic Go struct tags.
// Each [2]string is a key-value pair, where value is the unescaped string.
//
// Example: structTags{{"key", "value"}}.String() -> `key:"value"`
type structTags [][2]string

func (tags structTags) String() string {
	if len(tags) == 0 {
		return ""
	}
	var ss []string
	for _, tag := range tags {
		// NOTE: When quoting the value, we need to make sure the backtick
		// character does not appear. Convert all cases to the escaped hex form.
		key := tag[0]
		val := strings.Replace(strconv.Quote(tag[1]), "`", `\x60`, -1)
		ss = append(ss, fmt.Sprintf("%s:%s", key, val))
	}
	return "`" + strings.Join(ss, " ") + "`"
}

// appendDeprecationSuffix optionally appends a deprecation notice as a suffix.
func appendDeprecationSuffix(prefix protogen.Comments, parentFile protoreflect.FileDescriptor, deprecated bool) protogen.Comments {
	fileDeprecated := parentFile.Options().(*descriptorpb.FileOptions).GetDeprecated()
	if !deprecated && !fileDeprecated {
		return prefix
	}
	if prefix != "" {
		prefix += "\n"
	}
	if fileDeprecated {
		return prefix + " Deprecated: The entire proto file " + protogen.Comments(parentFile.Path()) + " is marked as deprecated.\n"
	}
	return prefix + " Deprecated: Marked as deprecated in " + protogen.Comments(parentFile.Path()) + ".\n"
}

// trailingComment is like protogen.Comments, but lacks a trailing newline.
type trailingComment protogen.Comments

func (c trailingComment) String() string {
	s := strings.TrimSuffix(protogen.Comments(c).String(), "\n")
	if strings.Contains(s, "\n") {
		// We don't support multi-lined trailing comments as it is unclear
		// how to best render them in the generated code.
		return ""
	}
	return s
}
