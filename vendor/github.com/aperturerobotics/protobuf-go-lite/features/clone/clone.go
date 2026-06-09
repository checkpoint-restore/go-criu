// Copyright (c) 2021 PlanetScale Inc. All rights reserved.
// Copyright (c) 2013, The GoGo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clone

import (
	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/aperturerobotics/protobuf-go-lite/generator"
)

const (
	cloneName        = "CloneVT"
	cloneMessageName = "CloneMessageVT"
	cloneOneofName   = "CloneOneofVT"
)

func init() {
	generator.RegisterFeature("clone", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &clone{GeneratedFile: gen}
	})
}

type clone struct {
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*clone)(nil)

func (p *clone) Name() string {
	return "clone"
}

func (p *clone) GenerateFile(file *protogen.File) bool {
	for _, message := range file.Messages {
		p.processMessage(message)
	}

	return p.once
}

// cloneOneofField generates the statements for cloning a oneof field
func (p *clone) cloneOneofField(lhsBase, rhsBase string, oneof *protogen.Oneof) {
	fieldname := oneof.GoName
	ccInterfaceName := "is" + oneof.GoIdent.GoName
	lhs := lhsBase + "." + fieldname
	rhs := rhsBase + "." + fieldname
	p.P(`if `, rhs, ` != nil {`)
	p.P(lhs, ` = `, rhs, `.(interface{ `, cloneOneofName, `() `, ccInterfaceName, ` }).`, cloneOneofName, `()`)
	p.P(`}`)
}

// cloneFieldSingular generates the code for cloning a singular, non-oneof field.
func (p *clone) cloneFieldSingular(lhs, rhs string, kind protoreflect.Kind) {
	switch {
	case kind == protoreflect.MessageKind, kind == protoreflect.GroupKind:
		// switch {
		// case p.IsWellKnownType(message):
		// p.P(lhs, ` = (*`, message.GoIdent, `)((*`, p.WellKnownTypeMap(message), `)(`, rhs, `).`, cloneName, `())`)
		// case p.IsLocalMessage(message):
		// default:
		p.P(lhs, ` = `, rhs, `.`, cloneName, `()`)
		//}
	case kind == protoreflect.BytesKind:
		p.P(lhs, ` = `, p.QualifiedGoIdent(protogen.GoImportPath("slices").Ident("Clone")), `(`, rhs, `)`)
	case isScalar(kind):
		p.P(lhs, ` = `, rhs)
	default:
		panic("unexpected")
	}
}

// cloneField generates the code for cloning a field in a protobuf.
func (p *clone) cloneField(lhsBase, rhsBase string, field *protogen.Field) {
	// At this point, if we encounter a non-synthetic oneof, we assume it to be the representative
	// field for that oneof.
	if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
		p.cloneOneofField(lhsBase, rhsBase, field.Oneof)
		return
	}

	if !p.FieldSemantics(field).Reference {
		panic("method should not be invoked for non-reference fields")
	}

	fieldname := field.GoName
	lhs := lhsBase + "." + fieldname
	rhs := rhsBase + "." + fieldname

	// At this point, we are only looking at reference types (pointers, maps, slices, interfaces), which can all
	// be nil.
	p.P(`if rhs := `, rhs, `; rhs != nil {`)
	rhs = "rhs"

	fieldKind := field.Desc.Kind()

	if field.Desc.Cardinality() == protoreflect.Repeated { // maps and slices
		if isScalar(fieldKind) && field.Desc.IsList() {
			// Generated code optimization: use slices.Clone for slices whose elements aren't reference types.
			p.P(lhs, ` = `, p.QualifiedGoIdent(protogen.GoImportPath("slices").Ident("Clone")), `(`, rhs, `)`)
		} else if field.Desc.IsMap() && isScalar(field.Message.Fields[1].Desc.Kind()) {
			// Use maps.Clone for maps with scalar values
			p.P(lhs, ` = `, p.QualifiedGoIdent(protogen.GoImportPath("maps").Ident("Clone")), `(`, rhs, `)`)
		} else {
			goType, _ := p.FieldGoType(field)
			p.P(lhs, ` = make(`, goType, `, len(`, rhs, `))`)
			if field.Desc.IsMap() {
				// For maps, the type of the value field determines what code is generated for cloning
				// an entry.
				valueField := field.Message.Fields[1]
				fieldKind = valueField.Desc.Kind()
			}
			p.P(`for k, v := range `, rhs, ` {`)
			p.cloneFieldSingular(lhs+"[k]", "v", fieldKind)
			p.P(`}`)
		}
	} else if isScalar(fieldKind) {
		p.P(`tmpVal := *`, rhs)
		p.P(lhs, ` = &tmpVal`)
	} else {
		p.cloneFieldSingular(lhs, rhs, fieldKind)
	}
	p.P(`}`)
}

func (p *clone) generateCloneMethodsForMessage(message *protogen.Message) {
	ccTypeName := message.GoIdent.GoName
	p.P(`func (m *`, ccTypeName, `) `, cloneName, `() *`, ccTypeName, ` {`)
	p.body(ccTypeName, message, true)
	p.P(`}`)
	p.P()

	p.P(`func (m *`, ccTypeName, `) `, cloneMessageName, `() `, p.QualifiedGoIdent(protogen.ProtobufGoLitePackage.Ident("CloneMessage")), ` {`)
	p.P(`return m.`, cloneName, `()`)
	p.P(`}`)
	p.P()
}

// body generates the code for the actual cloning logic of a structure containing the given fields.
// In practice, those can be the fields of a message.
// The object to be cloned is assumed to be called "m".
func (p *clone) body(ccTypeName string, message *protogen.Message, cloneUnknownFields bool) {
	// The method body for a message or a oneof wrapper always starts with a nil check.
	p.P(`if m == nil {`)
	// We use an explicitly typed nil to avoid returning the nil interface in the oneof wrapper
	// case.
	p.P(`return (*`, ccTypeName, `)(nil)`)
	p.P(`}`)

	fields := message.Fields
	// Make a first pass over the fields, in which we initialize all non-reference fields via direct
	// struct literal initialization, and extract all other (reference) fields for a second pass.
	// Do not require qualified name because CloneVT generates in same file with definition.
	p.Alloc("r", message, false)
	var refFields []*protogen.Field
	oneofFields := make(map[string]struct{}, len(fields))

	for _, field := range fields {
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			// Use the first field in a oneof as the representative for that oneof, disregard
			// the other fields in that oneof.
			if _, ok := oneofFields[field.Oneof.GoName]; !ok {
				refFields = append(refFields, field)
				oneofFields[field.Oneof.GoName] = struct{}{}
			}
			continue
		}

		if !p.FieldSemantics(field).Reference {
			p.P(`r.`, field.GoName, ` = m.`, field.GoName)
			continue
		}
		// Shortcut: for types where we know that an optimized clone method exists, we can call it directly as it is
		// nil-safe.
		if field.Desc.Cardinality() != protoreflect.Repeated {
			switch {
			case p.IsLocalMessage(field.Message):
				p.P(`r.`, field.GoName, ` = m.`, field.GoName, `.`, cloneName, `()`)
				continue
			}
		}
		refFields = append(refFields, field)
	}

	// Generate explicit assignment statements for all reference fields.
	for _, field := range refFields {
		p.cloneField("r", "m", field)
	}

	if cloneUnknownFields {
		// Clone unknown fields, if any
		p.P(`if len(m.unknownFields) > 0 {`)
		p.P(`r.unknownFields = `, p.QualifiedGoIdent(protogen.GoImportPath("slices").Ident("Clone")), `(m.unknownFields)`)
		p.P(`}`)
	}

	p.P(`return r`)
}

func (p *clone) bodyForOneOf(ccTypeName string, field *protogen.Field) {
	// The method body for a message or a oneof wrapper always starts with a nil check.
	p.P(`if m == nil {`)
	// We use an explicitly typed nil to avoid returning the nil interface in the oneof wrapper
	// case.
	p.P(`return (*`, ccTypeName, `)(nil)`)
	p.P(`}`)

	p.P("r", " := new(", ccTypeName, `)`)

	if !oneofWrapperReference(field) {
		p.P(`r.`, field.GoName, ` = m.`, field.GoName)
		p.P(`return r`)
		return
	}
	// Shortcut: for types where we know that an optimized clone method exists, we can call it directly as it is
	// nil-safe.
	if field.Desc.Cardinality() != protoreflect.Repeated && field.Message != nil {
		p.P(`r.`, field.GoName, ` = m.`, field.GoName, `.`, cloneName, `()`)
		p.P(`return r`)
		return
	}

	// Generate explicit assignment statements for reference field.
	p.cloneField("r", "m", field)

	p.P(`return r`)
}

// generateCloneMethodsForOneof generates the clone method for the oneof wrapper type of a
// field in a oneof.
func (p *clone) generateCloneMethodsForOneof(field *protogen.Field) {
	// Generate CloneVT() *TheType
	ccTypeName := field.GoIdent.GoName
	p.P(`func (m *`, ccTypeName, `) `, cloneName, `() *`, ccTypeName, ` {`)

	// Create a "fake" field for the single oneof member, pretending it is not a oneof field.
	fieldInOneof := *field
	fieldInOneof.Oneof = nil
	// If we have a scalar field in a oneof, that field is never nullable, even when using proto2
	p.bodyForOneOf(ccTypeName, &fieldInOneof)
	p.P(`}`)
	p.P()

	// Generate CloneOneofVT() isFooBar_Body
	ccInterfaceName := "is" + field.Oneof.GoIdent.GoName
	p.P(`func (m *`, ccTypeName, `) `, cloneOneofName, `() `, ccInterfaceName, ` {`)
	p.P(`return m.`, cloneName, `()`)
	p.P(`}`)
	p.P()
}

func (p *clone) processMessageOneofs(message *protogen.Message) {
	for _, field := range message.Fields {
		if field.Oneof == nil || field.Oneof.Desc.IsSynthetic() {
			continue
		}
		p.generateCloneMethodsForOneof(field)
	}
}

func (p *clone) processMessage(message *protogen.Message) {
	for _, nested := range message.Messages {
		p.processMessage(nested)
	}

	if message.Desc.IsMapEntry() {
		return
	}

	p.once = true

	p.generateCloneMethodsForMessage(message)
	p.processMessageOneofs(message)
}

func oneofWrapperReference(field *protogen.Field) bool {
	return field.Desc.Cardinality() == protoreflect.Repeated ||
		field.Desc.Kind() == protoreflect.BytesKind ||
		field.Desc.Kind() == protoreflect.MessageKind ||
		field.Desc.Kind() == protoreflect.GroupKind
}

func isScalar(kind protoreflect.Kind) bool {
	switch kind {
	case
		protoreflect.BoolKind,
		protoreflect.StringKind,
		protoreflect.DoubleKind, protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind,
		protoreflect.FloatKind, protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind,
		protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Sint32Kind,
		protoreflect.EnumKind:
		return true
	}
	return false
}
