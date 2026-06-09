// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
)

func (g *jsonGenerator) genEnum(enum *protogen.Enum) {
	if !hasDisableJsonComment(enum.Comments.Leading) {
		g.genEnumMarshaler(enum)
		g.genStdEnumMarshaler(enum)

		g.genEnumUnmarshaler(enum)
		g.genStdEnumUnmarshaler(enum)
	}
}
