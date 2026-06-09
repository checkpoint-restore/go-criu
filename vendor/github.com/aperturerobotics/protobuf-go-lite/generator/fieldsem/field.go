// Package fieldsem resolves protobuf descriptors to generated Go field
// representation semantics.
package fieldsem

import (
	"fmt"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Qualifier qualifies Go identifiers for the generated file currently being
// emitted.
type Qualifier interface {
	QualifiedGoIdent(protogen.GoIdent) string
}

// Field describes the generated Go representation and semantic decisions for a
// protobuf field.
type Field struct {
	Type string

	Pointer     bool
	Reference   bool
	Required    bool
	Packed      bool
	List        bool
	Map         bool
	RealOneof   bool
	Synthetic   bool
	Weak        bool
	EmitDefault bool
}

// Resolve resolves the generated Go representation for field.
func Resolve(q Qualifier, field *protogen.Field) Field {
	sem := Field{
		Required:  field.Desc.Cardinality() == protoreflect.Required,
		Packed:    field.Desc.IsPacked(),
		List:      field.Desc.IsList(),
		Map:       field.Desc.IsMap(),
		RealOneof: field.Oneof != nil && !field.Oneof.Desc.IsSynthetic(),
		Synthetic: field.Oneof != nil && field.Oneof.Desc.IsSynthetic(),
		Weak:      field.Desc.IsWeak(),
	}
	if sem.Weak {
		sem.Type = "struct{}"
		return sem
	}

	goType := ""
	pointer := field.Desc.HasPresence() && !sem.RealOneof
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.EnumKind:
		goType = q.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "[]byte"
		pointer = false
	case protoreflect.MessageKind, protoreflect.GroupKind:
		goType = "*" + q.QualifiedGoIdent(field.Message.GoIdent)
		pointer = false
	}

	switch {
	case sem.List:
		sem.Type = "[]" + goType
		sem.Pointer = false
	case sem.Map:
		key := Resolve(q, field.Message.Fields[0])
		val := Resolve(q, field.Message.Fields[1])
		sem.Type = fmt.Sprintf("map[%v]%v", key.Type, val.Type)
		sem.Pointer = false
	default:
		sem.Type = goType
		sem.Pointer = pointer
	}

	sem.Reference = sem.Pointer || sem.List || sem.Map || sem.RealOneof ||
		field.Desc.Kind() == protoreflect.BytesKind ||
		field.Desc.Kind() == protoreflect.MessageKind ||
		field.Desc.Kind() == protoreflect.GroupKind
	sem.EmitDefault = field.Desc.HasPresence() && !sem.List && !sem.Map
	return sem
}
