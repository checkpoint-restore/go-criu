// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"fmt"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (g *jsonGenerator) genMessage(message *protogen.Message) {
	// Generate marshalers and unmarshalers for all enums defined in the message.
	for _, enum := range message.Enums {
		g.genEnum(enum)
	}

	// Generate marshalers and unmarshalers for all sub-messages defined in the message.
	for _, message := range message.Messages {
		g.genMessage(message)
	}

	// skip early if the disable comment is present
	if hasDisableJsonComment(message.Comments.Leading) {
		return
	}

	g.genMessageMarshaler(message)
	g.genStdMessageMarshaler(message)

	g.genMessageUnmarshaler(message)
	g.genStdMessageUnmarshaler(message)
}

func (g *jsonGenerator) fieldIsNilable(field *protogen.Field) bool {
	sem := g.FieldSemantics(field)
	if sem.Pointer {
		return true
	}
	switch field.Desc.Kind() {
	case protoreflect.BytesKind:
		return field.Desc.HasPresence() && !sem.RealOneof
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return true
	default:
		return false
	}
}

func fieldGoName(field *protogen.Field) any {
	var fieldGoName any = field.GoName
	return fieldGoName
}

// libNameForField returns the name used in the protojson func that corresponds to the type of a given field.
func (g *jsonGenerator) libNameForField(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "Bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "Int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "Uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "Int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "Uint64"
	case protoreflect.FloatKind:
		return "Float32"
	case protoreflect.DoubleKind:
		return "Float64"
	case protoreflect.StringKind:
		return "String"
	case protoreflect.BytesKind:
		return "Bytes"
	default:
		g.gen.Error(fmt.Errorf("unsupported field kind %q", field.Desc.Kind()))
		return ""
	}
}
