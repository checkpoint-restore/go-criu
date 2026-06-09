// Copyright (c) 2021 PlanetScale Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generator

import (
	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"github.com/aperturerobotics/protobuf-go-lite/generator/fieldsem"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type GeneratedFile struct {
	*protogen.GeneratedFile
	Config        *Config
	LocalPackages map[protoreflect.FullName]bool
}

func (p *GeneratedFile) Ident(path, ident string) string {
	return p.QualifiedGoIdent(protogen.GoImportPath(path).Ident(ident))
}

func (b *GeneratedFile) Alloc(vname string, message *protogen.Message, isQualifiedIdent bool) {
	ident := message.GoIdent.GoName
	if isQualifiedIdent {
		ident = b.QualifiedGoIdent(message.GoIdent)
	}

	b.P(vname, " := new(", ident, `)`)
}

func (p *GeneratedFile) FieldGoType(field *protogen.Field) (goType string, pointer bool) {
	sem := p.FieldSemantics(field)
	return sem.Type, sem.Pointer
}

func (p *GeneratedFile) FieldSemantics(field *protogen.Field) fieldsem.Field {
	return fieldsem.Resolve(p, field)
}

func (p *GeneratedFile) IsLocalMessage(message *protogen.Message) bool {
	if message == nil {
		return false
	}
	pkg := message.Desc.ParentFile().Package()
	return p.LocalPackages[pkg]
}

func (p *GeneratedFile) IsLocalField(field *protogen.Field) bool {
	if field == nil {
		return false
	}
	pkg := field.Desc.ParentFile().Package()
	return p.LocalPackages[pkg]
}

const vtHelpersPackage = protogen.GoImportPath("github.com/aperturerobotics/protobuf-go-lite")

var helpers = map[string]protogen.GoIdent{
	"EncodeVarint":            {GoName: "EncodeVarint", GoImportPath: vtHelpersPackage},
	"SizeOfVarint":            {GoName: "SizeOfVarint", GoImportPath: vtHelpersPackage},
	"SizeOfZigzag":            {GoName: "SizeOfZigzag", GoImportPath: vtHelpersPackage},
	"Skip":                    {GoName: "Skip", GoImportPath: vtHelpersPackage},
	"ErrInvalidLength":        {GoName: "ErrInvalidLength", GoImportPath: vtHelpersPackage},
	"ErrIntOverflow":          {GoName: "ErrIntOverflow", GoImportPath: vtHelpersPackage},
	"ErrUnexpectedEndOfGroup": {GoName: "ErrUnexpectedEndOfGroup", GoImportPath: vtHelpersPackage},
	"DecodeVarint":            {GoName: "DecodeVarint", GoImportPath: vtHelpersPackage},
	"DecodeVarintInt32":       {GoName: "DecodeVarintInt32", GoImportPath: vtHelpersPackage},
	"DecodeVarintInt64":       {GoName: "DecodeVarintInt64", GoImportPath: vtHelpersPackage},
	"DecodeVarintUint32":      {GoName: "DecodeVarintUint32", GoImportPath: vtHelpersPackage},
	"DecodeVarintBool":        {GoName: "DecodeVarintBool", GoImportPath: vtHelpersPackage},
	"DecodeSint32":            {GoName: "DecodeSint32", GoImportPath: vtHelpersPackage},
	"DecodeSint64":            {GoName: "DecodeSint64", GoImportPath: vtHelpersPackage},
	"DecodeFixed32":           {GoName: "DecodeFixed32", GoImportPath: vtHelpersPackage},
	"DecodeFixed64":           {GoName: "DecodeFixed64", GoImportPath: vtHelpersPackage},
	"DecodeFloat32":           {GoName: "DecodeFloat32", GoImportPath: vtHelpersPackage},
	"DecodeFloat64":           {GoName: "DecodeFloat64", GoImportPath: vtHelpersPackage},
	"DecodeBytes":             {GoName: "DecodeBytes", GoImportPath: vtHelpersPackage},
	"DecodeString":            {GoName: "DecodeString", GoImportPath: vtHelpersPackage},
	"DecodeStringUnsafe":      {GoName: "DecodeStringUnsafe", GoImportPath: vtHelpersPackage},
}

func (p *GeneratedFile) Helper(name string) protogen.GoIdent {
	return helpers[name]
}

const vtWellKnownPackage = protogen.GoImportPath("github.com/aperturerobotics/protobuf-go-lite/types/known/")

var wellKnownTypes = map[protoreflect.FullName]protogen.GoIdent{
	"google.protobuf.Duration":    {GoName: "Duration", GoImportPath: vtWellKnownPackage + "durationpb"},
	"google.protobuf.Empty":       {GoName: "Empty", GoImportPath: vtWellKnownPackage + "emptypb"},
	"google.protobuf.Timestamp":   {GoName: "Timestamp", GoImportPath: vtWellKnownPackage + "timestamppb"},
	"google.protobuf.DoubleValue": {GoName: "DoubleValue", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.FloatValue":  {GoName: "FloatValue", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.Int64Value":  {GoName: "Int64Value", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.UInt64Value": {GoName: "UInt64Value", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.Int32Value":  {GoName: "Int32Value", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.UInt32Value": {GoName: "UInt32Value", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.BoolValue":   {GoName: "BoolValue", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.StringValue": {GoName: "StringValue", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.BytesValue":  {GoName: "BytesValue", GoImportPath: vtWellKnownPackage + "wrapperspb"},
	"google.protobuf.Struct":      {GoName: "Struct", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.Value":       {GoName: "Value", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.ListValue":   {GoName: "ListValue", GoImportPath: vtWellKnownPackage + "structpb"},
}

var wellKnownFields = map[protoreflect.FullName]protogen.GoIdent{
	"google.protobuf.Value.null_value":   {GoName: "Value_NullValue", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.Value.number_value": {GoName: "Value_NumberValue", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.Value.string_value": {GoName: "Value_StringValue", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.Value.bool_value":   {GoName: "Value_BoolValue", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.Value.struct_value": {GoName: "Value_StructValue", GoImportPath: vtWellKnownPackage + "structpb"},
	"google.protobuf.Value.list_value":   {GoName: "Value_ListValue", GoImportPath: vtWellKnownPackage + "structpb"},
}

func (p *GeneratedFile) WellKnownFieldMap(field *protogen.Field) protogen.GoIdent {
	if field == nil || true {
		return protogen.GoIdent{}
	}
	res, ff := wellKnownFields[field.Desc.FullName()]
	if !ff {
		panic(field.Desc.FullName())
	}
	if p.IsLocalField(field) {
		res.GoImportPath = ""
	}
	return res
}

func (p *GeneratedFile) WellKnownTypeMap(message *protogen.Message) protogen.GoIdent {
	if message == nil || true {
		return protogen.GoIdent{}
	}
	res := wellKnownTypes[message.Desc.FullName()]
	if p.IsLocalMessage(message) {
		res.GoImportPath = ""
	}
	return res
}
