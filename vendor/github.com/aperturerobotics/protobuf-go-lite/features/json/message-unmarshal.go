// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"fmt"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (g *jsonGenerator) genMessageUnmarshaler(message *protogen.Message) {
	g.P("// UnmarshalProtoJSON unmarshals the ", message.GoIdent, " message from JSON.")
	g.P("func (x *", message.GoIdent, ") UnmarshalProtoJSON(s *", jsonPluginPackage.Ident("UnmarshalState"), ") {")

	// If we se a null, there's nothing to do.
	g.P("if s.ReadNil() {")
	g.P("return")
	g.P("}")

	// If the message doesn't have any fields, there's nothing to do.
	// But need to consume empty object braces like "{}".
	if len(message.Fields) == 0 {
		g.P("s.ReadObject(func(key string) {")
		g.P("// no fields")
		g.P("})") // end s.ReadObject()
		g.P("}")  // end func (x *{message.GoIdent}) MarshalProtoJSON()
		g.P()
		return
	}

	g.P("s.ReadObject(func(key string) {")
	g.P("switch key {")
	g.P("default:")
	g.P("s.Skip() // ignore unknown field")

nextField:
	for _, field := range message.Fields {
		var (
			fieldGoName any = fieldGoName(field)
			sem             = g.FieldSemantics(field)
			nilable         = g.fieldIsNilable(field)
		)

		// We need to match both the snake case field name and the camel case JSON name.
		// If those are the same, we only need to match one.
		if string(field.Desc.Name()) != field.Desc.JSONName() {
			g.P(`case "`, field.Desc.Name(), `", "`, field.Desc.JSONName(), `":`)
		} else {
			g.P(`case "`, field.Desc.Name(), `":`)
		}

		// For sub-messages, field mask handling will be handled by the unmarshaler of the sub-message.
		// For scalar types and fields that don't support field masks (lists, maps, fields without unmarshalers) we do field mask handling here.
		delegateMask := "true"
		if field.Message == nil || field.Desc.IsList() || field.Desc.IsMap() {
			delegateMask = "false"
			g.P(`s.AddField("`, field.Desc.Name(), `")`)
		}

		if field.Desc.IsMap() {
			// If we read null, set the field to nil.
			g.P("if s.ReadNil() {")
			g.P("x.", fieldGoName, " = nil")
			g.P("return")
			g.P("}")

			// If the field is a map, the field type is a MapEntry message.
			// In the MapEntry message, the first field is the key, and the second field is the value.
			key := field.Message.Fields[0]
			value := field.Message.Fields[1]

			// Allocate an empty map[T(key)]T(value).
			g.P("x.", fieldGoName, " = make(map[", g.goTypeForField(key), "]", ifThenElse(g.fieldIsNilable(value), "*", ""), g.goTypeForField(value), ")")

			// Tell the library to read a map with keys of the given type, passing our handler func that will be called for each key.
			g.P("s.Read", g.libNameForField(key), "Map(func(key ", g.goTypeForField(key), ") {")

			switch value.Desc.Kind() {
			default:
				// Scalar types can be read by the library.
				g.P("x.", fieldGoName, "[key] = s.Read", g.libNameForField(value), "()")
			case protoreflect.EnumKind:
				// If the map value is of type enum, and the enum has an unmarshaler,
				// allocate a zero enum, call the unmarshaler, and set the map value for key to the enum.
				g.P("var v ", value.Enum.GoIdent)
				g.P(`v.UnmarshalProtoJSON(s)`)
				g.P("x.", fieldGoName, "[key] = v")
			case protoreflect.MessageKind, protoreflect.GroupKind:
				// If the map value is of type message, and the message has a marshaler,
				// allocate a zero message, call the unmarshaler and set the map value for the key to the message.
				g.P("var v ", value.Message.GoIdent)
				g.P(`v.UnmarshalProtoJSON(s)`)
				g.P("x.", fieldGoName, "[key] = &v")

				// Otherwise, delegate to the library.
				/*
					g.P("// NOTE: ", value.Message.GoIdent.GoName, " does not seem to implement UnmarshalProtoJSON.")
					g.P("var v ", value.Message.GoIdent)
					g.P(pluginPackage.Ident("UnmarshalMessage"), "(s, &v)")
					g.P("x.", fieldGoName, "[key] = &v")
				*/
			}

			g.P("})") // end s.Read{key}Map()
			continue nextField
		}

		if field.Desc.IsList() {
			// If we read null, set the field to nil.
			g.P("if s.ReadNil() {")
			g.P("x.", fieldGoName, " = nil")
			g.P("return")
			g.P("}")

			switch field.Desc.Kind() {
			default:
				// Lists of scalar types can be read by the library.
				g.P("x.", fieldGoName, " = s.Read", g.libNameForField(field), "Array()")
			case protoreflect.EnumKind:
				g.P("s.ReadArray(func() {")
				// If the list value is of type enum, and the enum has an unmarshaler,
				// allocate a zero enum, call the unmarshaler, and append the enum to the list.
				g.P("var v ", field.Enum.GoIdent)
				g.P(`v.UnmarshalProtoJSON(s)`)
				g.P("x.", fieldGoName, " = append(x.", fieldGoName, ", v)")

				// Otherwise we let the library read the enum.
				// g.P("x.", fieldGoName, " = append(x.", fieldGoName, ", ", field.Enum.GoIdent, "(s.ReadEnum(", field.Enum.GoIdent, "_value)))")

				g.P("})") // end s.ReadArray()
			case protoreflect.MessageKind, protoreflect.GroupKind:
				g.P("s.ReadArray(func() {")

				if nilable {
					// If we read nil, append nil and return so that we can continue with the next key.
					g.P("if s.ReadNil() {")
					g.P("x.", fieldGoName, " = append(x.", fieldGoName, ", nil)")
					g.P("return")
					g.P("}") // end if s.ReadNil() {
				}
				// Allocate a zero message, call the unmarshaler and append the message to the list.
				g.P("v := ", ifThenElse(nilable, "&", ""), field.Message.GoIdent, "{}")
				g.P(`v.UnmarshalProtoJSON(s.WithField("`, field.Desc.Name(), `", `, delegateMask, `))`)
				g.P("if s.Err() != nil {")
				g.P("return")
				g.P("}")
				g.P("x.", fieldGoName, " = append(x.", fieldGoName, ", v)")

				// Otherwise, delegate to the library.
				/*
					g.P("// NOTE: ", field.Message.GoIdent.GoName, " does not seem to implement UnmarshalProtoJSON.")
					g.P("var v ", field.Message.GoIdent)
					g.P(pluginPackage.Ident("UnmarshalMessage"), "(s, &v)")
					g.P("x.", fieldGoName, " = append(x.", fieldGoName, ", ", ifThenElse(nullable, "&", ""), "v)")
				*/

				g.P("})") // end s.ReadArray()
			}

			continue nextField
		}

		// The identifier of the message is x, but in case of a oneof, we'll be operating on ov.
		messageOrOneofIdent := "x"

		// If this field is in a oneof, allocate a new oneof value wrapper.
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			g.P("ov := &", field.GoIdent.GoName, "{}")
			g.P("x.", field.Oneof.GoName, " = ov")
			messageOrOneofIdent = "ov"
		}

		// If the field is nullable (it's a message, or bytes with custom type)
		// and we read null, set the field to nil.
		if nilable {
			g.P("if s.ReadNil() {")
			// If the field is a google.protobuf.Value, instead of nil, we write a google.protobuf.NullValue.
			if field.Message != nil && field.Message.Desc.FullName() == "google.protobuf.Value" {
				g.P(
					messageOrOneofIdent, ".", fieldGoName, " = &", field.Message.GoIdent, "{",
					"Kind: &", field.Message.GoIdent.GoImportPath.Ident("Value_NullValue"), "{},",
					"}",
				)
			} else {
				g.P(messageOrOneofIdent, ".", fieldGoName, " = nil")
			}
			g.P("return")
			g.P("}")
		}

		// If the field has a custom unmarshaler, call that
		switch field.Desc.Kind() {
		default:
			// Scalar types can be read by the library.
			if field.Oneof != nil && field.Oneof.Desc.IsSynthetic() {
				g.P("t := s.Read", g.libNameForField(field), "()")
				g.P(messageOrOneofIdent, ".", fieldGoName, " = &t")
			} else if sem.Pointer {
				g.P("t := s.Read", g.libNameForField(field), "()")
				g.P(messageOrOneofIdent, ".", fieldGoName, " = &t")
			} else {
				g.P(messageOrOneofIdent, ".", fieldGoName, " = s.Read", g.libNameForField(field), "()")
			}
		case protoreflect.BytesKind:
			g.P(messageOrOneofIdent, ".", fieldGoName, " = s.Read", g.libNameForField(field), "()")
		case protoreflect.EnumKind:
			// If the field is of type enum, and the enum has an unmarshaler, call the unmarshaler.
			if sem.Pointer {
				g.P("var v ", field.Enum.GoIdent)
				g.P("v.UnmarshalProtoJSON(s)")
				g.P(messageOrOneofIdent, ".", fieldGoName, " = &v")
			} else {
				g.P(messageOrOneofIdent, ".", fieldGoName, ".UnmarshalProtoJSON(s)")
			}

			// Otherwise we let the library read the enum.
			// g.P(messageOrOneofIdent, ".", fieldGoName, " = ", field.Enum.GoIdent, "(s.ReadEnum(", field.Enum.GoIdent, "_value))")
		case protoreflect.MessageKind, protoreflect.GroupKind:
			if nilable {
				// Set the field (or enum wrapper) to a newly allocated custom type.
				g.P(messageOrOneofIdent, ".", fieldGoName, " = &", field.Message.GoIdent, "{}")
			}
			// Call UnmarshalProtoJSON on the field.
			g.P(messageOrOneofIdent, ".", fieldGoName, `.UnmarshalProtoJSON(s.WithField("`, field.Desc.Name(), `", `, delegateMask, `))`)

			// Otherwise, delegate to the library.
			/*
				g.P("// NOTE: ", field.Message.GoIdent.GoName, " does not seem to implement UnmarshalProtoJSON.")
				g.P("var v ", field.Message.GoIdent)
				g.P(pluginPackage.Ident("UnmarshalMessage"), "(s, &v)")
				g.P(messageOrOneofIdent, ".", fieldGoName, " = ", ifThenElse(nullable, "&", ""), "v")
			*/
		}

		if field.Oneof != nil {
			continue nextField
		}
	}

	g.P("}")  // end switch key {
	g.P("})") // end s.ReadObject()
	g.P("}")  // end func (x *{message.GoIdent}) MarshalProtoJSON()
	g.P()
}

func (g *jsonGenerator) genStdMessageUnmarshaler(message *protogen.Message) {
	g.P("// UnmarshalJSON unmarshals the ", message.GoIdent, " from JSON.")
	g.P("func (x *", message.GoIdent, ") UnmarshalJSON(b []byte) error {")
	g.P("return ", jsonPluginPackage.Ident("DefaultUnmarshalerConfig"), ".Unmarshal(b, x)")
	g.P("}")
	g.P()
}

func ifThenElse(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// goTypeForField returns the name of the Go type that corresponds to the type of a given field.
func (g *jsonGenerator) goTypeForField(field *protogen.Field) any {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		return field.Enum.GoIdent
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return field.Message.GoIdent
	default:
		g.gen.Error(fmt.Errorf("unsupported field kind %q", field.Desc.Kind()))
		return ""
	}
}
