// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generator_base

import (
	"strings"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"github.com/aperturerobotics/protobuf-go-lite/internal/genid"
)

// Specialized support for well-known types are hard-coded into the generator
// as opposed to being injected in adjacent .go sources in the generated package
// in order to support specialized build systems like Bazel that always generate
// dynamically from the source .proto files.

func GenPackageKnownComment(f *protogen.File) protogen.Comments {
	switch f.Desc.Path() {
	case genid.File_google_protobuf_timestamp_proto:
		return ` Package timestamppb contains generated types for ` + genid.File_google_protobuf_timestamp_proto + `.

 The Timestamp message represents a timestamp,
 an instant in time since the Unix epoch (January 1st, 1970).


 Conversion to a Go Time

 The AsTime method can be used to convert a Timestamp message to a
 standard Go time.Time value in UTC:

	t := ts.AsTime()
	... // make use of t as a time.Time

 Converting to a time.Time is a common operation so that the extensive
 set of time-based operations provided by the time package can be leveraged.
 See https://golang.org/pkg/time for more information.

 The AsTime method performs the conversion on a best-effort basis. Timestamps
 with denormal values (e.g., nanoseconds beyond 0 and 99999999, inclusive)
 are normalized during the conversion to a time.Time. To manually check for
 invalid Timestamps per the documented limitations in timestamp.proto,
 additionally call the CheckValid method:

	if err := ts.CheckValid(); err != nil {
		... // handle error
	}


 Conversion from a Go Time

 The timestamppb.New function can be used to construct a Timestamp message
 from a standard Go time.Time value:

	ts := timestamppb.New(t)
	... // make use of ts as a *timestamppb.Timestamp

 In order to construct a Timestamp representing the current time, use Now:

	ts := timestamppb.Now()
	... // make use of ts as a *timestamppb.Timestamp

`
	case genid.File_google_protobuf_duration_proto:
		return ` Package durationpb contains generated types for ` + genid.File_google_protobuf_duration_proto + `.

 The Duration message represents a signed span of time.


 Conversion to a Go Duration

 The AsDuration method can be used to convert a Duration message to a
 standard Go time.Duration value:

	d := dur.AsDuration()
	... // make use of d as a time.Duration

 Converting to a time.Duration is a common operation so that the extensive
 set of time-based operations provided by the time package can be leveraged.
 See https://golang.org/pkg/time for more information.

 The AsDuration method performs the conversion on a best-effort basis.
 Durations with denormal values (e.g., nanoseconds beyond -99999999 and
 +99999999, inclusive; or seconds and nanoseconds with opposite signs)
 are normalized during the conversion to a time.Duration. To manually check for
 invalid Duration per the documented limitations in duration.proto,
 additionally call the CheckValid method:

	if err := dur.CheckValid(); err != nil {
		... // handle error
	}

 Note that the documented limitations in duration.proto does not protect a
 Duration from overflowing the representable range of a time.Duration in Go.
 The AsDuration method uses saturation arithmetic such that an overflow clamps
 the resulting value to the closest representable value (e.g., math.MaxInt64
 for positive overflow and math.MinInt64 for negative overflow).


 Conversion from a Go Duration

 The durationpb.New function can be used to construct a Duration message
 from a standard Go time.Duration value:

	dur := durationpb.New(d)
	... // make use of d as a *durationpb.Duration

`
	case genid.File_google_protobuf_struct_proto:
		return ` Package structpb contains generated types for ` + genid.File_google_protobuf_struct_proto + `.

 # Example usage

 Consider the following example JSON object:

	{
		"firstName": "John",
		"lastName": "Smith",
		"isAlive": true,
		"age": 27,
		"address": {
			"streetAddress": "21 2nd Street",
			"city": "New York",
			"state": "NY",
			"postalCode": "10021-3100"
		},
		"phoneNumbers": [
			{
				"type": "home",
				"number": "212 555-1234"
			},
			{
				"type": "office",
				"number": "646 555-4567"
			}
		],
		"children": [],
		"spouse": null
	}

 To construct a Value message representing the above JSON object:

	m, err := structpb.NewValue(map[string]interface{}{
		"firstName": "John",
		"lastName":  "Smith",
		"isAlive":   true,
		"age":       27,
		"address": map[string]interface{}{
			"streetAddress": "21 2nd Street",
			"city":          "New York",
			"state":         "NY",
			"postalCode":    "10021-3100",
		},
		"phoneNumbers": []interface{}{
			map[string]interface{}{
				"type":   "home",
				"number": "212 555-1234",
			},
			map[string]interface{}{
				"type":   "office",
				"number": "646 555-4567",
			},
		},
		"children": []interface{}{},
		"spouse":   nil,
	})
	if err != nil {
		... // handle error
	}
	... // make use of m as a *structpb.Value
`
	default:
		return ""
	}
}

func genMessageKnownFunctions(g *protogen.GeneratedFile, m *messageInfo) {
	switch m.Desc.FullName() {
	case genid.Timestamp_message_fullname:
		g.P("// Now constructs a new Timestamp from the current time.")
		g.P("func Now() *Timestamp {")
		g.P("	return New(", timePackage.Ident("Now"), "())")
		g.P("}")
		g.P()

		g.P("// AsRFC3339 returns the timestamp formatted as an RFC3339 string.")
		g.P("func (x *Timestamp) AsRFC3339() string {")
		g.P("	return x.AsTime().Format(", timePackage.Ident("RFC3339Nano"), ")")
		g.P("}")
		g.P()

		g.P("// New constructs a new Timestamp from the provided time.Time.")
		g.P("func New(t ", timePackage.Ident("Time"), ") *Timestamp {")
		g.P("	return &Timestamp{Seconds: int64(t.Unix()), Nanos: int32(t.Nanosecond())}")
		g.P("}")
		g.P()

		g.P("// AsTime converts x to a time.Time.")
		g.P("func (x *Timestamp) AsTime() ", timePackage.Ident("Time"), " {")
		g.P("	return ", timePackage.Ident("Unix"), "(int64(x.GetSeconds()), int64(x.GetNanos())).UTC()")
		g.P("}")
		g.P()

		g.P("// IsValid reports whether the timestamp is valid.")
		g.P("// It is equivalent to CheckValid == nil.")
		g.P("func (x *Timestamp) IsValid() bool {")
		g.P("	return x.check() == 0")
		g.P("}")
		g.P()

		g.P("// CheckValid returns an error if the timestamp is invalid.")
		g.P("// In particular, it checks whether the value represents a date that is")
		g.P("// in the range of 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.")
		g.P("// An error is reported for a nil Timestamp.")
		g.P("func (x *Timestamp) CheckValid() error {")
		g.P("	switch x.check() {")
		g.P("	case invalidNil:")
		g.P("		return ", errorsPackage.Ident("New"), "(\"invalid nil Timestamp\")")
		g.P("	case invalidUnderflow:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"timestamp (%v) before 0001-01-01\", x)")
		g.P("	case invalidOverflow:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"timestamp (%v) after 9999-12-31\", x)")
		g.P("	case invalidNanos:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"timestamp (%v) has out-of-range nanos\", x)")
		g.P("	default:")
		g.P("		return nil")
		g.P("	}")
		g.P("}")
		g.P()

		g.P("const (")
		g.P("	_ = iota")
		g.P("	invalidNil")
		g.P("	invalidUnderflow")
		g.P("	invalidOverflow")
		g.P("	invalidNanos")
		g.P(")")
		g.P()

		g.P("func (x *Timestamp) check() uint {")
		g.P("	const minTimestamp = -62135596800  // Seconds between 1970-01-01T00:00:00Z and 0001-01-01T00:00:00Z, inclusive")
		g.P("	const maxTimestamp = +253402300799 // Seconds between 1970-01-01T00:00:00Z and 9999-12-31T23:59:59Z, inclusive")
		g.P("	secs := x.GetSeconds()")
		g.P("	nanos := x.GetNanos()")
		g.P("	switch {")
		g.P("	case x == nil:")
		g.P("		return invalidNil")
		g.P("	case secs < minTimestamp:")
		g.P("		return invalidUnderflow")
		g.P("	case secs > maxTimestamp:")
		g.P("		return invalidOverflow")
		g.P("	case nanos < 0 || nanos >= 1e9:")
		g.P("		return invalidNanos")
		g.P("	default:")
		g.P("		return 0")
		g.P("	}")
		g.P("}")
		g.P()

	case genid.Duration_message_fullname:
		g.P("// New constructs a new Duration from the provided time.Duration.")
		g.P("func New(d ", timePackage.Ident("Duration"), ") *Duration {")
		g.P("	nanos := d.Nanoseconds()")
		g.P("	secs := nanos / 1e9")
		g.P("	nanos -= secs * 1e9")
		g.P("	return &Duration{Seconds: int64(secs), Nanos: int32(nanos)}")
		g.P("}")
		g.P()

		g.P("// AsDuration converts x to a time.Duration,")
		g.P("// returning the closest duration value in the event of overflow.")
		g.P("func (x *Duration) AsDuration() ", timePackage.Ident("Duration"), " {")
		g.P("	secs := x.GetSeconds()")
		g.P("	nanos := x.GetNanos()")
		g.P("	d := ", timePackage.Ident("Duration"), "(secs) * ", timePackage.Ident("Second"))
		g.P("	overflow := d/", timePackage.Ident("Second"), " != ", timePackage.Ident("Duration"), "(secs)")
		g.P("	d += ", timePackage.Ident("Duration"), "(nanos) * ", timePackage.Ident("Nanosecond"))
		g.P("	overflow = overflow || (secs < 0 && nanos < 0 && d > 0)")
		g.P("	overflow = overflow || (secs > 0 && nanos > 0 && d < 0)")
		g.P("	if overflow {")
		g.P("		switch {")
		g.P("		case secs < 0:")
		g.P("			return ", timePackage.Ident("Duration"), "(", mathPackage.Ident("MinInt64"), ")")
		g.P("		case secs > 0:")
		g.P("			return ", timePackage.Ident("Duration"), "(", mathPackage.Ident("MaxInt64"), ")")
		g.P("		}")
		g.P("	}")
		g.P("	return d")
		g.P("}")
		g.P()

		g.P("// IsValid reports whether the duration is valid.")
		g.P("// It is equivalent to CheckValid == nil.")
		g.P("func (x *Duration) IsValid() bool {")
		g.P("	return x.check() == 0")
		g.P("}")
		g.P()

		g.P("// CheckValid returns an error if the duration is invalid.")
		g.P("// In particular, it checks whether the value is within the range of")
		g.P("// -10000 years to +10000 years inclusive.")
		g.P("// An error is reported for a nil Duration.")
		g.P("func (x *Duration) CheckValid() error {")
		g.P("	switch x.check() {")
		g.P("	case invalidNil:")
		g.P("		return ", errorsPackage.Ident("New"), "(\"invalid nil Duration\")")
		g.P("	case invalidUnderflow:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"duration (%v) exceeds -10000 years\", x)")
		g.P("	case invalidOverflow:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"duration (%v) exceeds +10000 years\", x)")
		g.P("	case invalidNanosRange:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"duration (%v) has out-of-range nanos\", x)")
		g.P("	case invalidNanosSign:")
		g.P("		return ", fmtPackage.Ident("Errorf"), "(\"duration (%v) has seconds and nanos with different signs\", x)")
		g.P("	default:")
		g.P("		return nil")
		g.P("	}")
		g.P("}")
		g.P()

		g.P("const (")
		g.P("	_ = iota")
		g.P("	invalidNil")
		g.P("	invalidUnderflow")
		g.P("	invalidOverflow")
		g.P("	invalidNanosRange")
		g.P("	invalidNanosSign")
		g.P(")")
		g.P()

		g.P("func (x *Duration) check() uint {")
		g.P("	const absDuration = 315576000000 // 10000yr * 365.25day/yr * 24hr/day * 60min/hr * 60sec/min")
		g.P("	secs := x.GetSeconds()")
		g.P("	nanos := x.GetNanos()")
		g.P("	switch {")
		g.P("	case x == nil:")
		g.P("		return invalidNil")
		g.P("	case secs < -absDuration:")
		g.P("		return invalidUnderflow")
		g.P("	case secs > +absDuration:")
		g.P("		return invalidOverflow")
		g.P("	case nanos <= -1e9 || nanos >= +1e9:")
		g.P("		return invalidNanosRange")
		g.P("	case (secs > 0 && nanos < 0) || (secs < 0 && nanos > 0):")
		g.P("		return invalidNanosSign")
		g.P("	default:")
		g.P("		return 0")
		g.P("	}")
		g.P("}")
		g.P()

	case genid.Struct_message_fullname:
		g.P("// NewStruct constructs a Struct from a general-purpose Go map.")
		g.P("// The map keys must be valid UTF-8.")
		g.P("// The map values are converted using NewValue.")
		g.P("func NewStruct(v map[string]interface{}) (*Struct, error) {")
		g.P("	x := &Struct{Fields: make(map[string]*Value, len(v))}")
		g.P("	for k, v := range v {")
		g.P("		if !", utf8Package.Ident("ValidString"), "(k) {")
		g.P("			return nil, ", fmtPackage.Ident("Errorf"), "(\"invalid UTF-8 in string: %q\", k)")
		g.P("		}")
		g.P("		var err error")
		g.P("		x.Fields[k], err = NewValue(v)")
		g.P("		if err != nil {")
		g.P("			return nil, err")
		g.P("		}")
		g.P("	}")
		g.P("	return x, nil")
		g.P("}")
		g.P()

		g.P("// AsMap converts x to a general-purpose Go map.")
		g.P("// The map values are converted by calling Value.AsInterface.")
		g.P("func (x *Struct) AsMap() map[string]interface{} {")
		g.P("	f := x.GetFields()")
		g.P("	vs := make(map[string]interface{}, len(f))")
		g.P("	for k, v := range f {")
		g.P("		vs[k] = v.AsInterface()")
		g.P("	}")
		g.P("	return vs")
		g.P("}")
		g.P()

	case genid.ListValue_message_fullname:
		g.P("// NewList constructs a ListValue from a general-purpose Go slice.")
		g.P("// The slice elements are converted using NewValue.")
		g.P("func NewList(v []interface{}) (*ListValue, error) {")
		g.P("	x := &ListValue{Values: make([]*Value, len(v))}")
		g.P("	for i, v := range v {")
		g.P("		var err error")
		g.P("		x.Values[i], err = NewValue(v)")
		g.P("		if err != nil {")
		g.P("			return nil, err")
		g.P("		}")
		g.P("	}")
		g.P("	return x, nil")
		g.P("}")
		g.P()

		g.P("// AsSlice converts x to a general-purpose Go slice.")
		g.P("// The slice elements are converted by calling Value.AsInterface.")
		g.P("func (x *ListValue) AsSlice() []interface{} {")
		g.P("	vals := x.GetValues()")
		g.P("	vs := make([]interface{}, len(vals))")
		g.P("	for i, v := range vals {")
		g.P("		vs[i] = v.AsInterface()")
		g.P("	}")
		g.P("	return vs")
		g.P("}")
		g.P()

	case genid.Value_message_fullname:
		g.P("// NewValue constructs a Value from a general-purpose Go interface.")
		g.P("//")
		g.P("//	╔════════════════════════╤════════════════════════════════════════════╗")
		g.P("//	║ Go type                │ Conversion                                 ║")
		g.P("//	╠════════════════════════╪════════════════════════════════════════════╣")
		g.P("//	║ nil                    │ stored as NullValue                        ║")
		g.P("//	║ bool                   │ stored as BoolValue                        ║")
		g.P("//	║ int, int32, int64      │ stored as NumberValue                      ║")
		g.P("//	║ uint, uint32, uint64   │ stored as NumberValue                      ║")
		g.P("//	║ float32, float64       │ stored as NumberValue                      ║")
		g.P("//	║ string                 │ stored as StringValue; must be valid UTF-8 ║")
		g.P("//	║ []byte                 │ stored as StringValue; base64-encoded      ║")
		g.P("//	║ map[string]interface{} │ stored as StructValue                      ║")
		g.P("//	║ []interface{}          │ stored as ListValue                        ║")
		g.P("//	╚════════════════════════╧════════════════════════════════════════════╝")
		g.P("//")
		g.P("// When converting an int64 or uint64 to a NumberValue, numeric precision loss")
		g.P("// is possible since they are stored as a float64.")
		g.P("func NewValue(v interface{}) (*Value, error) {")
		g.P("	switch v := v.(type) {")
		g.P("	case nil:")
		g.P("		return NewNullValue(), nil")
		g.P("	case bool:")
		g.P("		return NewBoolValue(v), nil")
		g.P("	case int:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case int32:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case int64:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case uint:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case uint32:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case uint64:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case float32:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case float64:")
		g.P("		return NewNumberValue(float64(v)), nil")
		g.P("	case string:")
		g.P("		if !", utf8Package.Ident("ValidString"), "(v) {")
		g.P("			return nil, ", fmtPackage.Ident("Errorf"), "(\"invalid UTF-8 in string: %q\", v)")
		g.P("		}")
		g.P("		return NewStringValue(v), nil")
		g.P("	case []byte:")
		g.P("		s := ", base64Package.Ident("StdEncoding"), ".EncodeToString(v)")
		g.P("		return NewStringValue(s), nil")
		g.P("	case map[string]interface{}:")
		g.P("		v2, err := NewStruct(v)")
		g.P("		if err != nil {")
		g.P("			return nil, err")
		g.P("		}")
		g.P("		return NewStructValue(v2), nil")
		g.P("	case []interface{}:")
		g.P("		v2, err := NewList(v)")
		g.P("		if err != nil {")
		g.P("			return nil, err")
		g.P("		}")
		g.P("		return NewListValue(v2), nil")
		g.P("	default:")
		g.P("		return nil, ", errorsPackage.Ident("New"), "(\"invalid type\")")
		g.P("	}")
		g.P("}")
		g.P()

		g.P("// NewNullValue constructs a new null Value.")
		g.P("func NewNullValue() *Value {")
		g.P("	return &Value{Kind: &Value_NullValue{NullValue: NullValue_NULL_VALUE}}")
		g.P("}")
		g.P()

		g.P("// NewBoolValue constructs a new boolean Value.")
		g.P("func NewBoolValue(v bool) *Value {")
		g.P("	return &Value{Kind: &Value_BoolValue{BoolValue: v}}")
		g.P("}")
		g.P()

		g.P("// NewNumberValue constructs a new number Value.")
		g.P("func NewNumberValue(v float64) *Value {")
		g.P("	return &Value{Kind: &Value_NumberValue{NumberValue: v}}")
		g.P("}")
		g.P()

		g.P("// NewStringValue constructs a new string Value.")
		g.P("func NewStringValue(v string) *Value {")
		g.P("	return &Value{Kind: &Value_StringValue{StringValue: v}}")
		g.P("}")
		g.P()

		g.P("// NewStructValue constructs a new struct Value.")
		g.P("func NewStructValue(v *Struct) *Value {")
		g.P("	return &Value{Kind: &Value_StructValue{StructValue: v}}")
		g.P("}")
		g.P()

		g.P("// NewListValue constructs a new list Value.")
		g.P("func NewListValue(v *ListValue) *Value {")
		g.P("	return &Value{Kind: &Value_ListValue{ListValue: v}}")
		g.P("}")
		g.P()

		g.P("// AsInterface converts x to a general-purpose Go interface.")
		g.P("//")
		g.P("// Calling Value.MarshalJSON and \"encoding/json\".Marshal on this output produce")
		g.P("// semantically equivalent JSON (assuming no errors occur).")
		g.P("//")
		g.P("// Floating-point values (i.e., \"NaN\", \"Infinity\", and \"-Infinity\") are")
		g.P("// converted as strings to remain compatible with MarshalJSON.")
		g.P("func (x *Value) AsInterface() interface{} {")
		g.P("	switch v := x.GetKind().(type) {")
		g.P("	case *Value_NumberValue:")
		g.P("		if v != nil {")
		g.P("			switch {")
		g.P("			case ", mathPackage.Ident("IsNaN"), "(v.NumberValue):")
		g.P("				return \"NaN\"")
		g.P("			case ", mathPackage.Ident("IsInf"), "(v.NumberValue, +1):")
		g.P("				return \"Infinity\"")
		g.P("			case ", mathPackage.Ident("IsInf"), "(v.NumberValue, -1):")
		g.P("				return \"-Infinity\"")
		g.P("			default:")
		g.P("				return v.NumberValue")
		g.P("			}")
		g.P("		}")
		g.P("	case *Value_StringValue:")
		g.P("		if v != nil {")
		g.P("			return v.StringValue")
		g.P("		}")
		g.P("	case *Value_BoolValue:")
		g.P("		if v != nil {")
		g.P("			return v.BoolValue")
		g.P("		}")
		g.P("	case *Value_StructValue:")
		g.P("		if v != nil {")
		g.P("			return v.StructValue.AsMap()")
		g.P("		}")
		g.P("	case *Value_ListValue:")
		g.P("		if v != nil {")
		g.P("			return v.ListValue.AsSlice()")
		g.P("		}")
		g.P("	}")
		g.P("	return nil")
		g.P("}")
		g.P()

	case genid.FieldMask_message_fullname:
		g.P("// New constructs a field mask from a list of paths and verifies that")
		g.P("// each one is valid according to the specified message type.")
		g.P("func New(m any, paths ...string) (*FieldMask, error) {")
		g.P("	x := new(FieldMask)")
		g.P("	return x, x.Append(m, paths...)")
		g.P("}")
		g.P()

		g.P("// Union returns the union of all the paths in the input field masks.")
		g.P("func Union(mx *FieldMask, my *FieldMask, ms ...*FieldMask) *FieldMask {")
		g.P("	var out []string")
		g.P("	out = append(out, mx.GetPaths()...)")
		g.P("	out = append(out, my.GetPaths()...)")
		g.P("	for _, m := range ms {")
		g.P("		out = append(out, m.GetPaths()...)")
		g.P("	}")
		g.P("	return &FieldMask{Paths: normalizePaths(out)}")
		g.P("}")
		g.P()

		g.P("// Intersect returns the intersection of all the paths in the input field masks.")
		g.P("func Intersect(mx *FieldMask, my *FieldMask, ms ...*FieldMask) *FieldMask {")
		g.P("	var ss1, ss2 []string // reused buffers for performance")
		g.P("	intersect := func(out, in []string) []string {")
		g.P("		ss1 = normalizePaths(append(ss1[:0], in...))")
		g.P("		ss2 = normalizePaths(append(ss2[:0], out...))")
		g.P("		out = out[:0]")
		g.P("		for i1, i2 := 0, 0; i1 < len(ss1) && i2 < len(ss2); {")
		g.P("			switch s1, s2 := ss1[i1], ss2[i2]; {")
		g.P("			case hasPathPrefix(s1, s2):")
		g.P("				out = append(out, s1)")
		g.P("				i1++")
		g.P("			case hasPathPrefix(s2, s1):")
		g.P("				out = append(out, s2)")
		g.P("				i2++")
		g.P("			case lessPath(s1, s2):")
		g.P("				i1++")
		g.P("			case lessPath(s2, s1):")
		g.P("				i2++")
		g.P("			}")
		g.P("		}")
		g.P("		return out")
		g.P("	}")
		g.P()
		g.P("	out := Union(mx, my, ms...).GetPaths()")
		g.P("	out = intersect(out, mx.GetPaths())")
		g.P("	out = intersect(out, my.GetPaths())")
		g.P("	for _, m := range ms {")
		g.P("		out = intersect(out, m.GetPaths())")
		g.P("	}")
		g.P("	return &FieldMask{Paths: normalizePaths(out)}")
		g.P("}")
		g.P()

		g.P("// Normalize converts the mask to its canonical form where all paths are sorted")
		g.P("// and redundant paths are removed.")
		g.P("func (x *FieldMask) Normalize() {")
		g.P("	x.Paths = normalizePaths(x.Paths)")
		g.P("}")
		g.P()
		g.P("func normalizePaths(paths []string) []string {")
		g.P("	", sortPackage.Ident("Slice"), "(paths, func(i, j int) bool {")
		g.P("		return lessPath(paths[i], paths[j])")
		g.P("	})")
		g.P()
		g.P("	// Elide any path that is a prefix match on the previous.")
		g.P("	out := paths[:0]")
		g.P("	for _, path := range paths {")
		g.P("		if len(out) > 0 && hasPathPrefix(path, out[len(out)-1]) {")
		g.P("			continue")
		g.P("		}")
		g.P("		out = append(out, path)")
		g.P("	}")
		g.P("	return out")
		g.P("}")
		g.P()

		g.P("// hasPathPrefix is like strings.HasPrefix, but further checks for either")
		g.P("// an exact matche or that the prefix is delimited by a dot.")
		g.P("func hasPathPrefix(path, prefix string) bool {")
		g.P("	return ", stringsPackage.Ident("HasPrefix"), "(path, prefix) && (len(path) == len(prefix) || path[len(prefix)] == '.')")
		g.P("}")
		g.P()

		g.P("// lessPath is a lexicographical comparison where dot is specially treated")
		g.P("// as the smallest symbol.")
		g.P("func lessPath(x, y string) bool {")
		g.P("	for i := 0; i < len(x) && i < len(y); i++ {")
		g.P("		if x[i] != y[i] {")
		g.P("			return (x[i] - '.') < (y[i] - '.')")
		g.P("		}")
		g.P("	}")
		g.P("	return len(x) < len(y)")
		g.P("}")
		g.P()

		g.P("// rangeFields is like strings.Split(path, \".\"), but avoids allocations by")
		g.P("// iterating over each field in place and calling a iterator function.")
		g.P("func rangeFields(path string, f func(field string) bool) bool {")
		g.P("	for {")
		g.P("		var field string")
		g.P("		if i := ", stringsPackage.Ident("IndexByte"), "(path, '.'); i >= 0 {")
		g.P("			field, path = path[:i], path[i:]")
		g.P("		} else {")
		g.P("			field, path = path, \"\"")
		g.P("		}")
		g.P()
		g.P("		if !f(field) {")
		g.P("			return false")
		g.P("		}")
		g.P()
		g.P("		if len(path) == 0 {")
		g.P("			return true")
		g.P("		}")
		g.P("		path = ", stringsPackage.Ident("TrimPrefix"), "(path, \".\")")
		g.P("	}")
		g.P("}")
		g.P()

	case genid.BoolValue_message_fullname,
		genid.Int32Value_message_fullname,
		genid.Int64Value_message_fullname,
		genid.UInt32Value_message_fullname,
		genid.UInt64Value_message_fullname,
		genid.FloatValue_message_fullname,
		genid.DoubleValue_message_fullname,
		genid.StringValue_message_fullname,
		genid.BytesValue_message_fullname:
		funcName := strings.TrimSuffix(m.GoIdent.GoName, "Value")
		typeName := strings.ToLower(funcName)
		switch typeName {
		case "float":
			typeName = "float32"
		case "double":
			typeName = "float64"
		case "bytes":
			typeName = "[]byte"
		}

		g.P("// ", funcName, " stores v in a new ", m.GoIdent, " and returns a pointer to it.")
		g.P("func ", funcName, "(v ", typeName, ") *", m.GoIdent, " {")
		g.P("	return &", m.GoIdent, "{Value: v}")
		g.P("}")
		g.P()
	}
}
