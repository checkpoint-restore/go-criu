// Copyright (c) 2022 PlanetScale Inc. All rights reserved.

package equal

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/aperturerobotics/protobuf-go-lite/generator"
)

func init() {
	generator.RegisterFeature("equal", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &equal{GeneratedFile: gen}
	})
}

type equal struct {
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*equal)(nil)

func (p *equal) Name() string { return "equal" }

func (p *equal) GenerateFile(file *protogen.File) bool {
	for _, message := range file.Messages {
		p.message(message)
	}
	return p.once
}

const (
	equalName        = "EqualVT"
	equalMessageName = "EqualMessageVT"
)

func (p *equal) message(message *protogen.Message) {
	for _, nested := range message.Messages {
		p.message(nested)
	}

	if message.Desc.IsMapEntry() {
		return
	}

	p.once = true

	ccTypeName := message.GoIdent.GoName
	p.P(`func (this *`, ccTypeName, `) `, equalName, `(that *`, ccTypeName, `) bool {`)

	p.P(`if this == that {`)
	p.P(`	return true`)
	p.P(`} else if this == nil || that == nil {`)
	p.P(`	return false`)
	p.P(`}`)

	slices.SortFunc(message.Fields, func(a, b *protogen.Field) int {
		return cmp.Compare(a.Desc.Number(), b.Desc.Number())
	})

	{
		oneofs := make(map[string]struct{}, len(message.Fields))
		for _, field := range message.Fields {
			oneof := field.Oneof != nil && !field.Oneof.Desc.IsSynthetic()
			if !oneof {
				continue
			}

			fieldname := field.Oneof.GoName
			if _, ok := oneofs[fieldname]; ok {
				continue
			}
			oneofs[fieldname] = struct{}{}

			p.P(`if this.`, fieldname, ` == nil && that.`, fieldname, ` != nil {`)
			p.P(`	return false`)
			p.P(`} else if this.`, fieldname, ` != nil {`)
			p.P(`	if that.`, fieldname, ` == nil {`)
			p.P(`		return false`)
			p.P(`	}`)
			ccInterfaceName := fmt.Sprintf("is%s", field.Oneof.GoIdent.GoName)
			p.P(`if !this.`, fieldname, `.(interface{ `, equalName, `(`, ccInterfaceName, `) bool }).`, equalName, `(that.`, fieldname, `) {`)
			p.P(`return false`)
			p.P(`}`)
			p.P(`}`)
		}
	}

	for _, field := range message.Fields {
		oneof := field.Oneof != nil && !field.Oneof.Desc.IsSynthetic()
		if !oneof {
			p.field(field, field.Desc.HasPresence() && !field.Desc.IsList() && !field.Desc.IsMap())
		}
	}

	p.P(`return string(this.unknownFields) == string(that.unknownFields)`)
	p.P(`}`)
	p.P()

	p.P(`func (this *`, ccTypeName, `) `, equalMessageName, `(thatMsg any) bool {`)
	p.P(`that, ok := thatMsg.(*`, ccTypeName, `)`)
	p.P(`if !ok {`)
	p.P(`return false`)
	p.P(`}`)
	p.P(`return this.`, equalName, `(that)`)
	p.P(`}`)

	for _, field := range message.Fields {
		oneof := field.Oneof != nil && !field.Oneof.Desc.IsSynthetic()
		if !oneof {
			continue
		}
		p.oneof(field)
	}
}

func (p *equal) oneof(field *protogen.Field) {
	ccTypeName := field.GoIdent.GoName
	ccInterfaceName := fmt.Sprintf("is%s", field.Oneof.GoIdent.GoName)
	fieldname := field.GoName

	p.P(`func (this *`, ccTypeName, `) `, equalName, `(thatIface `, ccInterfaceName, `) bool {`)
	p.P(`that, ok := thatIface.(*`, ccTypeName, `)`)
	p.P(`if !ok {`)
	p.P(`return false`)
	p.P(`}`)
	p.P(`if this == that {`)
	p.P(`return true`)
	p.P(`}`)
	p.P(`if this == nil && that != nil || this != nil && that == nil {`)
	p.P(`return false`)
	p.P(`}`)

	lhs := fmt.Sprintf("this.%s", fieldname)
	rhs := fmt.Sprintf("that.%s", fieldname)
	kind := field.Desc.Kind()
	switch {
	case isScalar(kind):
		p.compareScalar(lhs, rhs, false)
	case kind == protoreflect.BytesKind:
		if p.Config.HelperCodegen() {
			p.helperCheck("EqualBytes", lhs, rhs)
			break
		}
		p.compareBytes(lhs, rhs, false)
	case kind == protoreflect.MessageKind || kind == protoreflect.GroupKind:
		if p.Config.HelperCodegen() {
			p.helperCheck("EqualVTImplicit", lhs, rhs, p.emptyMessageFunc(field.Message))
			break
		}
		p.compareCall(lhs, rhs, field.Message, false)
	default:
		panic("not implemented")
	}
	p.P(`return true`)
	p.P(`}`)
	p.P()
}

func (p *equal) helperCall(name string, args ...any) []any {
	call := []any{p.Helper(name), `(`}
	for i, arg := range args {
		if i != 0 {
			call = append(call, `, `)
		}
		call = append(call, arg)
	}
	call = append(call, `)`)
	return call
}

func (p *equal) helperCheck(name string, args ...any) {
	line := []any{`if !`}
	line = append(line, p.helperCall(name, args...)...)
	line = append(line, ` {`)
	p.P(line...)
	p.P(`return false`)
	p.P(`}`)
}

func (p *equal) emptyMessageFunc(message *protogen.Message) string {
	ident := p.QualifiedGoIdent(message.GoIdent)
	return fmt.Sprintf("func() *%s { return &%s{} }", ident, ident)
}

func (p *equal) helperField(field *protogen.Field, nullable bool) {
	fieldname := field.GoName
	lhs := fmt.Sprintf("this.%s", fieldname)
	rhs := fmt.Sprintf("that.%s", fieldname)

	if field.Desc.IsMap() {
		valueField := field.Message.Fields[1]
		switch kind := valueField.Desc.Kind(); {
		case isScalar(kind):
			p.helperCheck("EqualMap", lhs, rhs)
		case kind == protoreflect.BytesKind:
			p.helperCheck("EqualBytesMap", lhs, rhs)
		case kind == protoreflect.MessageKind || kind == protoreflect.GroupKind:
			p.helperCheck("EqualVTMapImplicit", lhs, rhs, p.emptyMessageFunc(valueField.Message))
		default:
			panic("not implemented")
		}
		return
	}

	if field.Desc.IsList() {
		switch kind := field.Desc.Kind(); {
		case isScalar(kind):
			p.helperCheck("EqualSlice", lhs, rhs)
		case kind == protoreflect.BytesKind:
			p.helperCheck("EqualBytesSlice", lhs, rhs)
		case kind == protoreflect.MessageKind || kind == protoreflect.GroupKind:
			p.helperCheck("EqualVTSliceImplicit", lhs, rhs, p.emptyMessageFunc(field.Message))
		default:
			panic("not implemented")
		}
		return
	}

	kind := field.Desc.Kind()
	switch {
	case isScalar(kind):
		if nullable {
			p.helperCheck("EqualPtr", lhs, rhs)
			return
		}
		p.compareScalar(lhs, rhs, false)
	case kind == protoreflect.BytesKind:
		if nullable {
			p.helperCheck("EqualBytesPresent", lhs, rhs)
			return
		}
		p.helperCheck("EqualBytes", lhs, rhs)
	case kind == protoreflect.MessageKind || kind == protoreflect.GroupKind:
		if nullable {
			p.helperCheck("IsEqualVT", lhs, rhs)
			return
		}
		p.helperCheck("EqualVTImplicit", lhs, rhs, p.emptyMessageFunc(field.Message))
	default:
		panic("not implemented")
	}
}

func (p *equal) field(field *protogen.Field, nullable bool) {
	if p.Config.HelperCodegen() {
		p.helperField(field, nullable)
		return
	}

	fieldname := field.GoName

	repeated := field.Desc.Cardinality() == protoreflect.Repeated
	lhs := fmt.Sprintf("this.%s", fieldname)
	rhs := fmt.Sprintf("that.%s", fieldname)

	if repeated {
		p.P(`if len(`, lhs, `) != len(`, rhs, `) {`)
		p.P(`	return false`)
		p.P(`}`)
		p.P(`for i, vx := range `, lhs, ` {`)
		if field.Desc.IsMap() {
			p.P(`vy, ok := `, rhs, `[i]`)
			p.P(`if !ok {`)
			p.P(`return false`)
			p.P(`}`)

			field = field.Message.Fields[1]
		} else {
			p.P(`vy := `, rhs, `[i]`)
		}
		lhs, rhs = "vx", "vy"
		nullable = false
	}

	kind := field.Desc.Kind()
	switch {
	case isScalar(kind):
		p.compareScalar(lhs, rhs, nullable)

	case kind == protoreflect.BytesKind:
		p.compareBytes(lhs, rhs, nullable)

	case kind == protoreflect.MessageKind || kind == protoreflect.GroupKind:
		p.compareCall(lhs, rhs, field.Message, nullable)

	default:
		panic("not implemented")
	}

	if repeated {
		// close for loop
		p.P(`}`)
	}
}

func (p *equal) compareScalar(lhs, rhs string, nullable bool) {
	if nullable {
		p.P(`if p, q := `, lhs, `, `, rhs, `; (p == nil && q != nil) || (p != nil && (q == nil || *p != *q)) {`)
	} else {
		p.P(`if `, lhs, ` != `, rhs, ` {`)
	}
	p.P(`	return false`)
	p.P(`}`)
}

func (p *equal) compareBytes(lhs, rhs string, nullable bool) {
	if nullable {
		p.P(`if p, q := `, lhs, `, `, rhs, `; (p == nil && q != nil) || (p != nil && q == nil) || string(p) != string(q) {`)
	} else {
		// Inlined call to bytes.Equal()
		p.P(`if string(`, lhs, `) != string(`, rhs, `) {`)
	}
	p.P(`	return false`)
	p.P(`}`)
}

func (p *equal) compareCall(lhs, rhs string, msg *protogen.Message, nullable bool) {
	if !nullable {
		// The p != q check is mostly intended to catch the lhs = nil, rhs = nil case in which we would pointlessly
		// allocate not just one but two empty values. However, it also provides us with an extra scope to establish
		// our p and q variables.
		p.P(`if p, q := `, lhs, `, `, rhs, `; p != q {`)
		defer p.P(`}`)

		p.P(`if p == nil {`)
		p.P(`p = &`, p.QualifiedGoIdent(msg.GoIdent), `{}`)
		p.P(`}`)
		p.P(`if q == nil {`)
		p.P(`q = &`, p.QualifiedGoIdent(msg.GoIdent), `{}`)
		p.P(`}`)
		lhs, rhs = "p", "q"
	}
	p.P(`if !`, lhs, `.`, equalName, `(`, rhs, `) {`)
	p.P(`	return false`)
	p.P(`}`)
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
