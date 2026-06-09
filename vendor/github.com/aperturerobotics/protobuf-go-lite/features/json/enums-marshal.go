// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
)

func (g *jsonGenerator) genEnumMarshaler(enum *protogen.Enum) {
	g.P("// MarshalProtoJSON marshals the ", enum.GoIdent, " to JSON.")
	g.P("func (x ", enum.GoIdent, ") MarshalProtoJSON(s *", jsonPluginPackage.Ident("MarshalState"), ") {")
	// We write the enum, passing only the original mapping to the marshaler.
	g.P("s.WriteEnum(int32(x), ", enum.GoIdent, "_name)")
	g.P("}")
	g.P()
}

func (g *jsonGenerator) genStdEnumMarshaler(enum *protogen.Enum) {
	g.P("// MarshalText marshals the ", enum.GoIdent, " to text.")
	g.P("func (x ", enum.GoIdent, ") MarshalText() ([]byte, error) {")
	g.P("return []byte(", jsonPluginPackage.Ident("GetEnumString"), "(int32(x), ", enum.GoIdent, "_name)), nil")
	g.P("}")
	g.P()

	g.P("// MarshalJSON marshals the ", enum.GoIdent, " to JSON.")
	g.P("func (x ", enum.GoIdent, ") MarshalJSON() ([]byte, error) {")
	g.P("return ", jsonPluginPackage.Ident("DefaultMarshalerConfig"), ".Marshal(x)")
	g.P("}")
	g.P()
}
