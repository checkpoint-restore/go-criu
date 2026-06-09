// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The protoc-gen-go-lite binary is a protoc plugin to generate Go code for
// both proto2 and proto3 versions of the protocol buffer language.
//
// For more information about the usage of this plugin, see:
// https://protobuf.dev/reference/go/go-generated.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"github.com/aperturerobotics/protobuf-go-lite/generator"
	"github.com/aperturerobotics/protobuf-go-lite/internal/version"

	_ "github.com/aperturerobotics/protobuf-go-lite/features/clone"
	_ "github.com/aperturerobotics/protobuf-go-lite/features/equal"
	_ "github.com/aperturerobotics/protobuf-go-lite/features/json"
	_ "github.com/aperturerobotics/protobuf-go-lite/features/marshal"
	_ "github.com/aperturerobotics/protobuf-go-lite/features/size"
	_ "github.com/aperturerobotics/protobuf-go-lite/features/text"
	_ "github.com/aperturerobotics/protobuf-go-lite/features/unmarshal"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%v %v\n", filepath.Base(os.Args[0]), version.String())
		os.Exit(0)
	}

	var cfg generator.Config
	var features string
	var f flag.FlagSet

	f.BoolVar(&cfg.AllowEmpty, "allow-empty", false, "allow generation of empty files")
	f.StringVar(&features, "features", "all", "list of features to generate (separated by '+')")
	f.StringVar(&cfg.BuildTag, "buildTag", "", "the go:build tag to set on generated files")

	protogen.Options{
		ParamFunc: f.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		gen, err := generator.NewGenerator(plugin, strings.Split(features, "+"), &cfg)
		if err != nil {
			return err
		}
		gen.Generate()

		return nil
	})
}
