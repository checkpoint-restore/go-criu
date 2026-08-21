package generator

import (
	"encoding/base64"
	"slices"
	"strconv"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type flatOption struct {
	Name   string
	Values []string
}

func collectMessageOptions(file *protogen.File, message *protogen.Message) []flatOption {
	dp := descriptorProtoForMessage(file, message)
	if dp == nil {
		return nil
	}
	opts := dp.GetOptions()
	if opts == nil {
		return nil
	}

	var order []string
	byName := make(map[string][]string)
	opts.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}
		flattenOptionValue(string(fd.FullName()), fd, v, &order, byName)
		return true
	})

	slices.Sort(order)
	out := make([]flatOption, 0, len(order))
	for _, name := range order {
		out = append(out, flatOption{Name: name, Values: byName[name]})
	}
	return out
}

func descriptorProtoForMessage(file *protogen.File, message *protogen.Message) *descriptorpb.DescriptorProto {
	path := messageDescriptorPath(message.Desc)
	list := file.Proto.GetMessageType()
	var cur *descriptorpb.DescriptorProto
	for i, idx := range path {
		if idx < 0 || idx >= len(list) {
			return nil
		}
		cur = list[idx]
		if i == len(path)-1 {
			return cur
		}
		list = cur.GetNestedType()
	}
	return nil
}

func messageDescriptorPath(md protoreflect.MessageDescriptor) []int {
	var path []int
	for {
		path = append(path, md.Index())
		parent, ok := md.Parent().(protoreflect.MessageDescriptor)
		if !ok {
			break
		}
		md = parent
	}
	slices.Reverse(path)
	return path
}

func flattenOptionValue(prefix string, fd protoreflect.FieldDescriptor, v protoreflect.Value, order *[]string, byName map[string][]string) {
	if fd.IsList() {
		list := v.List()
		for i := 0; i < list.Len(); i++ {
			flattenOptionSingular(prefix, fd, list.Get(i), order, byName)
		}
		return
	}
	if fd.IsMap() {
		mapv := v.Map()
		mapv.Range(func(k protoreflect.MapKey, mv protoreflect.Value) bool {
			keyName := prefix + "." + k.String()
			flattenOptionSingular(keyName, fd.MapValue(), mv, order, byName)
			return true
		})
		return
	}
	flattenOptionSingular(prefix, fd, v, order, byName)
}

func flattenOptionSingular(prefix string, fd protoreflect.FieldDescriptor, v protoreflect.Value, order *[]string, byName map[string][]string) {
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		msg := v.Message()
		msg.Range(func(fd2 protoreflect.FieldDescriptor, v2 protoreflect.Value) bool {
			flattenOptionValue(prefix+"."+string(fd2.Name()), fd2, v2, order, byName)
			return true
		})
	default:
		appendFlatOption(prefix, scalarOptionString(fd, v), order, byName)
	}
}

func appendFlatOption(name, value string, order *[]string, byName map[string][]string) {
	if _, ok := byName[name]; !ok {
		*order = append(*order, name)
	}
	byName[name] = append(byName[name], value)
}

func scalarOptionString(fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return strconv.FormatBool(v.Bool())
	case protoreflect.EnumKind:
		if ev := fd.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return string(ev.Name())
		}
		return strconv.FormatInt(int64(v.Enum()), 10)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return strconv.FormatInt(v.Int(), 10)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(v.Int(), 10)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return strconv.FormatUint(v.Uint(), 10)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(v.Uint(), 10)
	case protoreflect.FloatKind:
		return strconv.FormatFloat(v.Float(), 'g', -1, 32)
	case protoreflect.DoubleKind:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.BytesKind:
		return base64.StdEncoding.EncodeToString(v.Bytes())
	default:
		return v.String()
	}
}
