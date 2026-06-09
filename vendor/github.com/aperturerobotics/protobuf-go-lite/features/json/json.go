// Copyright © 2024 Aperture Robotics, LLC.
// Copyright © 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"strings"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"github.com/aperturerobotics/protobuf-go-lite/generator"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	jsonPluginPackage = protogen.GoImportPath("github.com/aperturerobotics/protobuf-go-lite/json")
)

var disableJsonComment = "protobuf-go-lite:disable-json"

// hasDisableJsonComment checks if a comments section has the disable json comment.
func hasDisableJsonComment(comments protogen.Comments) bool {
	for _, line := range strings.Split(strings.TrimSuffix(string(comments), "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == disableJsonComment {
			return true
		}
	}
	return false
}

type jsonGenerator struct {
	gen  *protogen.Plugin
	file *protogen.File
	*generator.GeneratedFile
}

func init() {
	generator.RegisterFeature("json", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &jsonGenerator{GeneratedFile: gen}
	})
}

func (g *jsonGenerator) GenerateFile(file *protogen.File) bool {
	g.file = file

	// If the file doesn't have marshalers or unmarshalers, we can skip it.
	if !g.fileHasAnyMarshaler() {
		return false
	}

	if file.Desc.Syntax() != protoreflect.Proto3 && file.Desc.Syntax() != protoreflect.Editions {
		g.P("// NOTE: protobuf-go-lite json only supports proto3 and editions: ", file.Desc.Syntax().String(), " is not supported.")
		g.P()
		return true
	}

	g.generateFileContent()
	return true
}

func (g *jsonGenerator) fileHasAnyMarshaler() bool {
	return len(g.file.Enums) != 0 || len(g.file.Messages) != 0
}

func (g *jsonGenerator) generateFileContent() {
	for _, enum := range g.file.Enums {
		g.genEnum(enum)
	}
	for _, message := range g.file.Messages {
		g.genMessage(message)
	}
}
