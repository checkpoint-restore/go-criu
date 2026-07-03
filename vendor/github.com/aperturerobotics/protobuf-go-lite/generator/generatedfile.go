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
	"EncodeBool":                    {GoName: "EncodeBool", GoImportPath: vtHelpersPackage},
	"EncodeBytes":                   {GoName: "EncodeBytes", GoImportPath: vtHelpersPackage},
	"EncodeFixed32":                 {GoName: "EncodeFixed32", GoImportPath: vtHelpersPackage},
	"EncodeFixed64":                 {GoName: "EncodeFixed64", GoImportPath: vtHelpersPackage},
	"EncodeRawBytes":                {GoName: "EncodeRawBytes", GoImportPath: vtHelpersPackage},
	"EncodeString":                  {GoName: "EncodeString", GoImportPath: vtHelpersPackage},
	"EncodeVarint":                  {GoName: "EncodeVarint", GoImportPath: vtHelpersPackage},
	"EncodeVarintPacked":            {GoName: "EncodeVarintPacked", GoImportPath: vtHelpersPackage},
	"EncodeZigzag32":                {GoName: "EncodeZigzag32", GoImportPath: vtHelpersPackage},
	"EncodeZigzag32Packed":          {GoName: "EncodeZigzag32Packed", GoImportPath: vtHelpersPackage},
	"EncodeZigzag64":                {GoName: "EncodeZigzag64", GoImportPath: vtHelpersPackage},
	"EncodeZigzag64Packed":          {GoName: "EncodeZigzag64Packed", GoImportPath: vtHelpersPackage},
	"PackedFixedElementCount":       {GoName: "PackedFixedElementCount", GoImportPath: vtHelpersPackage},
	"PackedVarintElementCount":      {GoName: "PackedVarintElementCount", GoImportPath: vtHelpersPackage},
	"SizeOfVarint":                  {GoName: "SizeOfVarint", GoImportPath: vtHelpersPackage},
	"SizeOfZigzag":                  {GoName: "SizeOfZigzag", GoImportPath: vtHelpersPackage},
	"Skip":                          {GoName: "Skip", GoImportPath: vtHelpersPackage},
	"SkipWithin":                    {GoName: "SkipWithin", GoImportPath: vtHelpersPackage},
	"ErrInvalidLength":              {GoName: "ErrInvalidLength", GoImportPath: vtHelpersPackage},
	"ErrIntOverflow":                {GoName: "ErrIntOverflow", GoImportPath: vtHelpersPackage},
	"ErrUnexpectedEndOfGroup":       {GoName: "ErrUnexpectedEndOfGroup", GoImportPath: vtHelpersPackage},
	"DecodeVarint":                  {GoName: "DecodeVarint", GoImportPath: vtHelpersPackage},
	"DecodeVarintInt32":             {GoName: "DecodeVarintInt32", GoImportPath: vtHelpersPackage},
	"DecodeVarintInt64":             {GoName: "DecodeVarintInt64", GoImportPath: vtHelpersPackage},
	"DecodeVarintUint32":            {GoName: "DecodeVarintUint32", GoImportPath: vtHelpersPackage},
	"DecodeVarintBool":              {GoName: "DecodeVarintBool", GoImportPath: vtHelpersPackage},
	"DecodeSint32":                  {GoName: "DecodeSint32", GoImportPath: vtHelpersPackage},
	"DecodeSint64":                  {GoName: "DecodeSint64", GoImportPath: vtHelpersPackage},
	"DecodeFixed32":                 {GoName: "DecodeFixed32", GoImportPath: vtHelpersPackage},
	"DecodeFixed64":                 {GoName: "DecodeFixed64", GoImportPath: vtHelpersPackage},
	"DecodeFloat32":                 {GoName: "DecodeFloat32", GoImportPath: vtHelpersPackage},
	"DecodeFloat64":                 {GoName: "DecodeFloat64", GoImportPath: vtHelpersPackage},
	"DecodeBytes":                   {GoName: "DecodeBytes", GoImportPath: vtHelpersPackage},
	"DecodeBytesAppend":             {GoName: "DecodeBytesAppend", GoImportPath: vtHelpersPackage},
	"DecodeLengthDelimited":         {GoName: "DecodeLengthDelimited", GoImportPath: vtHelpersPackage},
	"DecodeString":                  {GoName: "DecodeString", GoImportPath: vtHelpersPackage},
	"DecodeStringUnsafe":            {GoName: "DecodeStringUnsafe", GoImportPath: vtHelpersPackage},
	"CloneBytes":                    {GoName: "CloneBytes", GoImportPath: vtHelpersPackage},
	"CloneBytesMap":                 {GoName: "CloneBytesMap", GoImportPath: vtHelpersPackage},
	"CloneBytesSlice":               {GoName: "CloneBytesSlice", GoImportPath: vtHelpersPackage},
	"CloneMap":                      {GoName: "CloneMap", GoImportPath: vtHelpersPackage},
	"ClonePtr":                      {GoName: "ClonePtr", GoImportPath: vtHelpersPackage},
	"CloneSlice":                    {GoName: "CloneSlice", GoImportPath: vtHelpersPackage},
	"CloneVTMap":                    {GoName: "CloneVTMap", GoImportPath: vtHelpersPackage},
	"CloneVTSlice":                  {GoName: "CloneVTSlice", GoImportPath: vtHelpersPackage},
	"CloneVTValue":                  {GoName: "CloneVTValue", GoImportPath: vtHelpersPackage},
	"EqualBytes":                    {GoName: "EqualBytes", GoImportPath: vtHelpersPackage},
	"EqualBytesMap":                 {GoName: "EqualBytesMap", GoImportPath: vtHelpersPackage},
	"EqualBytesPresent":             {GoName: "EqualBytesPresent", GoImportPath: vtHelpersPackage},
	"EqualBytesSlice":               {GoName: "EqualBytesSlice", GoImportPath: vtHelpersPackage},
	"EqualMap":                      {GoName: "EqualMap", GoImportPath: vtHelpersPackage},
	"EqualPtr":                      {GoName: "EqualPtr", GoImportPath: vtHelpersPackage},
	"EqualSlice":                    {GoName: "EqualSlice", GoImportPath: vtHelpersPackage},
	"EqualVTImplicit":               {GoName: "EqualVTImplicit", GoImportPath: vtHelpersPackage},
	"EqualVTMapImplicit":            {GoName: "EqualVTMapImplicit", GoImportPath: vtHelpersPackage},
	"EqualVTSliceImplicit":          {GoName: "EqualVTSliceImplicit", GoImportPath: vtHelpersPackage},
	"IsEqualVT":                     {GoName: "IsEqualVT", GoImportPath: vtHelpersPackage},
	"SizeBoolNonZero":               {GoName: "SizeBoolNonZero", GoImportPath: vtHelpersPackage},
	"SizeBoolPacked":                {GoName: "SizeBoolPacked", GoImportPath: vtHelpersPackage},
	"SizeBoolPtr":                   {GoName: "SizeBoolPtr", GoImportPath: vtHelpersPackage},
	"SizeBoolSlice":                 {GoName: "SizeBoolSlice", GoImportPath: vtHelpersPackage},
	"SizeBoolValue":                 {GoName: "SizeBoolValue", GoImportPath: vtHelpersPackage},
	"SizeBytesNonEmpty":             {GoName: "SizeBytesNonEmpty", GoImportPath: vtHelpersPackage},
	"SizeBytesPresent":              {GoName: "SizeBytesPresent", GoImportPath: vtHelpersPackage},
	"SizeBytesSlice":                {GoName: "SizeBytesSlice", GoImportPath: vtHelpersPackage},
	"SizeBytesValue":                {GoName: "SizeBytesValue", GoImportPath: vtHelpersPackage},
	"SizeFixed32NonZero":            {GoName: "SizeFixed32NonZero", GoImportPath: vtHelpersPackage},
	"SizeFixed32Packed":             {GoName: "SizeFixed32Packed", GoImportPath: vtHelpersPackage},
	"SizeFixed32Ptr":                {GoName: "SizeFixed32Ptr", GoImportPath: vtHelpersPackage},
	"SizeFixed32Slice":              {GoName: "SizeFixed32Slice", GoImportPath: vtHelpersPackage},
	"SizeFixed32Value":              {GoName: "SizeFixed32Value", GoImportPath: vtHelpersPackage},
	"SizeFixed64NonZero":            {GoName: "SizeFixed64NonZero", GoImportPath: vtHelpersPackage},
	"SizeFixed64Packed":             {GoName: "SizeFixed64Packed", GoImportPath: vtHelpersPackage},
	"SizeFixed64Ptr":                {GoName: "SizeFixed64Ptr", GoImportPath: vtHelpersPackage},
	"SizeFixed64Slice":              {GoName: "SizeFixed64Slice", GoImportPath: vtHelpersPackage},
	"SizeFixed64Value":              {GoName: "SizeFixed64Value", GoImportPath: vtHelpersPackage},
	"SizeGroup":                     {GoName: "SizeGroup", GoImportPath: vtHelpersPackage},
	"SizeMessage":                   {GoName: "SizeMessage", GoImportPath: vtHelpersPackage},
	"SizeStringNonEmpty":            {GoName: "SizeStringNonEmpty", GoImportPath: vtHelpersPackage},
	"SizeStringPtr":                 {GoName: "SizeStringPtr", GoImportPath: vtHelpersPackage},
	"SizeStringSlice":               {GoName: "SizeStringSlice", GoImportPath: vtHelpersPackage},
	"SizeStringValue":               {GoName: "SizeStringValue", GoImportPath: vtHelpersPackage},
	"SizeVarintNonZero":             {GoName: "SizeVarintNonZero", GoImportPath: vtHelpersPackage},
	"SizeVarintPacked":              {GoName: "SizeVarintPacked", GoImportPath: vtHelpersPackage},
	"SizeVarintPtr":                 {GoName: "SizeVarintPtr", GoImportPath: vtHelpersPackage},
	"SizeVarintSlice":               {GoName: "SizeVarintSlice", GoImportPath: vtHelpersPackage},
	"SizeVarintValue":               {GoName: "SizeVarintValue", GoImportPath: vtHelpersPackage},
	"SizeZigzagNonZero":             {GoName: "SizeZigzagNonZero", GoImportPath: vtHelpersPackage},
	"SizeZigzagPacked":              {GoName: "SizeZigzagPacked", GoImportPath: vtHelpersPackage},
	"SizeZigzagPtr":                 {GoName: "SizeZigzagPtr", GoImportPath: vtHelpersPackage},
	"SizeZigzagSlice":               {GoName: "SizeZigzagSlice", GoImportPath: vtHelpersPackage},
	"SizeZigzagValue":               {GoName: "SizeZigzagValue", GoImportPath: vtHelpersPackage},
	"TextBuilder":                   {GoName: "TextBuilder", GoImportPath: vtHelpersPackage},
	"TextFinishMessage":             {GoName: "TextFinishMessage", GoImportPath: vtHelpersPackage},
	"TextSortedMapKeys":             {GoName: "TextSortedMapKeys", GoImportPath: vtHelpersPackage},
	"TextStartMessage":              {GoName: "TextStartMessage", GoImportPath: vtHelpersPackage},
	"TextWriteBool":                 {GoName: "TextWriteBool", GoImportPath: vtHelpersPackage},
	"TextWriteBytes":                {GoName: "TextWriteBytes", GoImportPath: vtHelpersPackage},
	"TextWriteFieldPrefix":          {GoName: "TextWriteFieldPrefix", GoImportPath: vtHelpersPackage},
	"TextWriteFloat32":              {GoName: "TextWriteFloat32", GoImportPath: vtHelpersPackage},
	"TextWriteFloat64":              {GoName: "TextWriteFloat64", GoImportPath: vtHelpersPackage},
	"TextWriteInt":                  {GoName: "TextWriteInt", GoImportPath: vtHelpersPackage},
	"TextWriteListEnd":              {GoName: "TextWriteListEnd", GoImportPath: vtHelpersPackage},
	"TextWriteListSeparator":        {GoName: "TextWriteListSeparator", GoImportPath: vtHelpersPackage},
	"TextWriteListStart":            {GoName: "TextWriteListStart", GoImportPath: vtHelpersPackage},
	"TextWriteMapEnd":               {GoName: "TextWriteMapEnd", GoImportPath: vtHelpersPackage},
	"TextWriteMapEntryPrefix":       {GoName: "TextWriteMapEntryPrefix", GoImportPath: vtHelpersPackage},
	"TextWriteMapKeyValueSeparator": {GoName: "TextWriteMapKeyValueSeparator", GoImportPath: vtHelpersPackage},
	"TextWriteMapStart":             {GoName: "TextWriteMapStart", GoImportPath: vtHelpersPackage},
	"TextWriteString":               {GoName: "TextWriteString", GoImportPath: vtHelpersPackage},
	"TextWriteStringer":             {GoName: "TextWriteStringer", GoImportPath: vtHelpersPackage},
	"TextWriteTextMarshaler":        {GoName: "TextWriteTextMarshaler", GoImportPath: vtHelpersPackage},
	"TextWriteUint":                 {GoName: "TextWriteUint", GoImportPath: vtHelpersPackage},
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
