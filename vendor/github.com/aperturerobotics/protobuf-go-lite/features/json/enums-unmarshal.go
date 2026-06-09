// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
)

func (g *jsonGenerator) genEnumUnmarshaler(enum *protogen.Enum) {
	g.P("// UnmarshalProtoJSON unmarshals the ", enum.GoIdent, " from JSON.")
	g.P("func (x *", enum.GoIdent, ") UnmarshalProtoJSON(s *", jsonPluginPackage.Ident("UnmarshalState"), ") {")
	// We read the enum, passing only the original mapping to the unmarshaler.
	g.P("v := s.ReadEnum(", enum.GoIdent, "_value)")
	g.P("if err := s.Err(); err != nil {")
	g.P(`s.SetErrorf("could not read `, enum.Desc.Name(), ` enum: %v", err)`)
	g.P("return")
	g.P("}")
	g.P("*x = ", enum.GoIdent, "(v)")
	g.P("}")
	g.P()
}

func (g *jsonGenerator) genStdEnumUnmarshaler(enum *protogen.Enum) {
	g.P("// UnmarshalText unmarshals the ", enum.GoIdent, " from text.")
	g.P("func (x *", enum.GoIdent, ") UnmarshalText(b []byte) error {")
	g.P("i, err := ", jsonPluginPackage.Ident("ParseEnumString"), "(string(b), ", enum.GoIdent, "_value)")
	g.P("if err != nil {")
	g.P("return err")
	g.P("}")
	g.P("*x = ", enum.GoIdent, "(i)")
	g.P("return nil")
	g.P("}")
	g.P()

	g.P("// UnmarshalJSON unmarshals the ", enum.GoIdent, " from JSON.")
	g.P("func (x *", enum.GoIdent, ") UnmarshalJSON(b []byte) error {")
	g.P("return ", jsonPluginPackage.Ident("DefaultUnmarshalerConfig"), ".Unmarshal(b, x)")
	g.P("}")
	g.P()
}
