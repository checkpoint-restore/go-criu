// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (g *jsonGenerator) genMessageMarshaler(message *protogen.Message) {
	g.P("// MarshalProtoJSON marshals the ", message.GoIdent, " message to JSON.")
	g.P("func (x *", message.GoIdent, ") MarshalProtoJSON(s *", jsonPluginPackage.Ident("MarshalState"), ") {")

	g.P("if x == nil {")
	g.P("s.WriteNil()")
	g.P("return")
	g.P("}")

	g.P("s.WriteObjectStart()")

	// If the message doesn't have any fields, there's nothing to do.
	if len(message.Fields) == 0 {
		g.P("s.WriteObjectEnd()")
		g.P("}") // end func (x *{message.GoIdent}) MarshalProtoJSON()
		g.P()
		return
	}

	// wroteField keeps track of whether we wrote a field, so that we know when to add a comma before the next.
	g.P("var wroteField bool")

nextField:
	for _, field := range message.Fields {
		var (
			fieldGoName   any = fieldGoName(field)
			sem               = g.FieldSemantics(field)
			nilable           = g.fieldIsNilable(field)
			fieldJsonName     = field.Desc.JSONName()
		)

		if field.Desc.IsMap() {
			// If the field is a map, the field type is a MapEntry message.
			// In the MapEntry message, the first field is the key, and the second field is the value.
			key := field.Message.Fields[0]
			value := field.Message.Fields[1]

			// We emit the field if the map is not nil
			g.P("if x.", fieldGoName, ` != nil || s.HasField("`, fieldJsonName, `") {`)

			// Write a comma if this isn't the first field.
			g.P("s.WriteMoreIf(&wroteField)")

			// Write the field name and a colon.
			g.P(`s.WriteObjectField("`, fieldJsonName, `")`)

			g.P("s.WriteObjectStart()")

			// wroteElement keeps track of whether we wrote an element of the map, so that we know when to add a comma before the next.
			g.P("var wroteElement bool")

			g.P("for k, v := range x.", fieldGoName, " {")

			// Write a comma if this isn't the first element of the map.
			g.P("s.WriteMoreIf(&wroteElement)")

			// Write the key and a a colon. Since they key can be of other types than string, we use the library to convert those.
			g.P("s.WriteObject", g.libNameForField(key), "Field(k)")

			switch value.Desc.Kind() {
			default:
				// Scalar types can be written by the library.
				g.P("s.Write", g.libNameForField(value), "(v)")
			case protoreflect.EnumKind:
				g.P("v.MarshalProtoJSON(s)")
			case protoreflect.MessageKind, protoreflect.GroupKind:
				g.P(`v.MarshalProtoJSON(s.WithField("`, fieldJsonName, `"))`)
			}

			g.P("}") // end for k, v := range x.{fieldGoName} {
			g.P("s.WriteObjectEnd()")
			g.P("}") // end if x.{fieldGoName} != nil {

			continue nextField
		}

		if field.Desc.IsList() {
			// We emit the field if the list is not empty or if it's specified in the field mask.
			g.P("if len(x.", fieldGoName, `) > 0 || s.HasField("`, fieldJsonName, `") {`)

			// Write a comma if this isn't the first field.
			g.P("s.WriteMoreIf(&wroteField)")

			// Write the field name and a colon.
			g.P(`s.WriteObjectField("`, fieldJsonName, `")`)

			switch field.Desc.Kind() {
			default:
				g.P("s.Write", g.libNameForField(field), "Array(x.", fieldGoName, ")")
			case protoreflect.EnumKind:
				g.P("s.WriteArrayStart()")

				// wroteElement keeps track of whether we wrote an element of the list, so that we know when to add a comma before the next.
				g.P("var wroteElement bool")

				g.P("for _, element := range x.", fieldGoName, " {")

				// Write a comma if this isn't the first element of the list.
				g.P("s.WriteMoreIf(&wroteElement)")

				g.P("element.MarshalProtoJSON(s)")

				g.P("}") // end for _, element := range x.{fieldGoName} {
				g.P("s.WriteArrayEnd()")
			case protoreflect.MessageKind, protoreflect.GroupKind:
				g.P("s.WriteArrayStart()")

				// wroteElement keeps track of whether we wrote an element of the list, so that we know when to add a comma before the next.
				g.P("var wroteElement bool")

				g.P("for _, element := range x.", fieldGoName, " {")

				// Write a comma if this isn't the first element of the list.
				g.P("s.WriteMoreIf(&wroteElement)")

				// If the list element is of type message, and the message has a marshaler, use that.
				g.P(`element.MarshalProtoJSON(s.WithField("`, fieldJsonName, `"))`)
				// Otherwise delegate to the library.
				// g.P("// NOTE: ", field.Message.GoIdent.GoName, " does not seem to implement MarshalProtoJSON.")
				// g.P(jsonPluginPackage.Ident("MarshalMessage"), "(s, ", ifThenElse(nullable, "", "&"), "element)")

				g.P("}") // end for _, element := range x.{fieldGoName} {
				g.P("s.WriteArrayEnd()")
			}

			g.P("}") // end if len(x.{fieldGoName}) > 0 {

			continue nextField
		}

		// The identifier of the message is x, but in case of a oneof, we'll be operating on ov.
		messageOrOneofIdent := "x"

		// If this is the first field in a oneof, write the if statement that checks for nil
		// and start the switch statement for the oneof type.
		if field.Oneof != nil && field == field.Oneof.Fields[0] && !field.Oneof.Desc.IsSynthetic() {
			// NOTE: we don't support field masks here (yet).
			g.P("if x.", field.Oneof.GoName, " != nil {")
			g.P("switch ov := x.", field.Oneof.GoName, ".(type) {")
		}

		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			// If we're in a oneof, check if this is the field that's set in the oneof.
			g.P("case *", field.GoIdent.GoName, ":")
			messageOrOneofIdent = "ov"
		} else {
			// If we're not in a oneof, start "if not zero value".
			if nilable {
				// If this field is nullable, we emit it if it's not nil or if it's specified in the field mask.
				g.P("if ", messageOrOneofIdent, ".", fieldGoName, ` != nil || s.HasField("`, fieldJsonName, `") {`)
			} else {
				// If this field is not nullable, we emit it if it's not the zero value or if it's specified in the field mask.
				switch field.Desc.Kind() {
				case protoreflect.BoolKind:
					g.P("if ", messageOrOneofIdent, ".", fieldGoName, ` || s.HasField("`, fieldJsonName, `") {`)
				case protoreflect.EnumKind:
					g.P("if ", messageOrOneofIdent, ".", fieldGoName, ` != 0 || s.HasField("`, fieldJsonName, `") {`)
				case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
					protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
					protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
					protoreflect.Uint64Kind, protoreflect.Fixed64Kind,
					protoreflect.FloatKind,
					protoreflect.DoubleKind:
					g.P("if ", messageOrOneofIdent, ".", fieldGoName, ` != 0 || s.HasField("`, fieldJsonName, `") {`)
				case protoreflect.StringKind:
					g.P("if ", messageOrOneofIdent, ".", fieldGoName, ` != "" || s.HasField("`, fieldJsonName, `") {`)
				case protoreflect.BytesKind:
					g.P("if len(", messageOrOneofIdent, ".", fieldGoName, `) > 0 || s.HasField("`, fieldJsonName, `") {`)
				case protoreflect.MessageKind, protoreflect.GroupKind:
					// For not-nullable messages we have a dummy check.
					g.P("if true { ")
				}
			}
		}

		// Write a comma if this isn't the first field.
		g.P("s.WriteMoreIf(&wroteField)")

		// Write the field name and a colon.
		g.P(`s.WriteObjectField("`, fieldJsonName, `")`)

		switch field.Desc.Kind() {
		default:
			// Scalar types can be written by the library.
			if sem.Pointer {
				g.P("s.Write", g.libNameForField(field), "(*", messageOrOneofIdent, ".", fieldGoName, ")")
			} else {
				g.P("s.Write", g.libNameForField(field), "(", messageOrOneofIdent, ".", fieldGoName, ")")
			}
		case protoreflect.BytesKind:
			g.P("s.Write", g.libNameForField(field), "(", messageOrOneofIdent, ".", fieldGoName, ")")
		case protoreflect.EnumKind:
			// If the field is of type enum, and the enum has a marshaler, use that.
			if sem.Pointer {
				g.P("(*", messageOrOneofIdent, ".", fieldGoName, ").MarshalProtoJSON(s)")
			} else {
				g.P(messageOrOneofIdent, ".", fieldGoName, ".MarshalProtoJSON(s)")
			}
			// Otherwise we write the enum with the standard settings.
			// g.P("s.WriteEnum(int32(", messageOrOneofIdent, ".", fieldGoName, "), ", field.Enum.GoIdent, "_name)")
		case protoreflect.MessageKind, protoreflect.GroupKind:
			// If the field is of type message, and the message has a marshaler, use that.
			g.P(messageOrOneofIdent, ".", fieldGoName, `.MarshalProtoJSON(s.WithField("`, fieldJsonName, `"))`)
			// Otherwise delegate to the library.
			//	g.P("// NOTE: ", field.Message.GoIdent.GoName, " does not seem to implement MarshalProtoJSON.")
			// g.P(jsonPluginPackage.Ident("MarshalMessage"), "(s, ", ifThenElse(nullable, "", "&"), messageOrOneofIdent, ".", fieldGoName, ")")
		}

		// If we're not in a oneof, end the "if not zero".
		if field.Oneof == nil || field.Oneof.Desc.IsSynthetic() {
			g.P("}") // end if x.{field.GoName} != zero value {
		}

		// If this is the last field in the oneof, close the switch and if statements.
		if field.Oneof != nil && field == field.Oneof.Fields[len(field.Oneof.Fields)-1] && !field.Oneof.Desc.IsSynthetic() {
			g.P("}") // end switch v := x.{field.Oneof.GoName}.(type) {
			g.P("}") // end if x.{field.Oneof.GoName} != nil {
		}
	}

	g.P("s.WriteObjectEnd()")

	g.P("}") // end func (x *{message.GoIdent}) MarshalProtoJSON()
	g.P()
}

func (g *jsonGenerator) genStdMessageMarshaler(message *protogen.Message) {
	g.P("// MarshalJSON marshals the ", message.GoIdent, " to JSON.")
	g.P("func (x *", message.GoIdent, ") MarshalJSON() ([]byte, error) {")
	g.P("return ", jsonPluginPackage.Ident("DefaultMarshalerConfig"), ".Marshal(x)")
	g.P("}")
	g.P()
}
