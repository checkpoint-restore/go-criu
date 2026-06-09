package generator

import (
	"fmt"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/gofeaturespb"
)

func validateEditionFile(file *protogen.File) error {
	if file.Desc.Syntax() != protoreflect.Editions {
		return nil
	}
	if err := validateFeatureSet(file.Desc.FullName(), file.Proto.GetOptions().GetFeatures()); err != nil {
		return err
	}
	for _, enum := range file.Enums {
		if err := validateEditionEnum(enum); err != nil {
			return err
		}
	}
	for _, message := range file.Messages {
		if err := validateEditionMessage(message); err != nil {
			return err
		}
	}
	return nil
}

func validateEditionEnum(enum *protogen.Enum) error {
	if enum.Desc.IsClosed() {
		return fmt.Errorf("%s: closed Edition enums are not supported by protobuf-go-lite", enum.Desc.FullName())
	}
	if err := validateFeatureSet(enum.Desc.FullName(), enum.Desc.Options().(*descriptorpb.EnumOptions).GetFeatures()); err != nil {
		return err
	}
	return nil
}

func validateEditionMessage(message *protogen.Message) error {
	if err := validateFeatureSet(message.Desc.FullName(), message.Desc.Options().(*descriptorpb.MessageOptions).GetFeatures()); err != nil {
		return err
	}
	for _, enum := range message.Enums {
		if err := validateEditionEnum(enum); err != nil {
			return err
		}
	}
	for _, field := range message.Fields {
		if err := validateEditionField(field); err != nil {
			return err
		}
	}
	for _, nested := range message.Messages {
		if err := validateEditionMessage(nested); err != nil {
			return err
		}
	}
	return nil
}

func validateEditionField(field *protogen.Field) error {
	if field.Enum != nil && field.Enum.Desc.IsClosed() {
		return fmt.Errorf("%s: closed Edition enum field %s is not supported by protobuf-go-lite", field.Desc.FullName(), field.Enum.Desc.FullName())
	}
	return validateFeatureSet(field.Desc.FullName(), field.Desc.Options().(*descriptorpb.FieldOptions).GetFeatures())
}

func validateFeatureSet(name protoreflect.FullName, features *descriptorpb.FeatureSet) error {
	if features == nil {
		return nil
	}
	if features.GetJsonFormat() == descriptorpb.FeatureSet_LEGACY_BEST_EFFORT {
		return fmt.Errorf("%s: Edition json_format LEGACY_BEST_EFFORT is not supported by protobuf-go-lite", name)
	}
	if !proto.HasExtension(features, gofeaturespb.E_Go) {
		return nil
	}
	ext := proto.GetExtension(features, gofeaturespb.E_Go)
	goFeatures, ok := ext.(*gofeaturespb.GoFeatures)
	if !ok {
		return fmt.Errorf("%s: unexpected Go feature extension type %T", name, ext)
	}
	switch goFeatures.GetApiLevel() {
	case gofeaturespb.GoFeatures_API_LEVEL_UNSPECIFIED, gofeaturespb.GoFeatures_API_OPEN:
		return nil
	case gofeaturespb.GoFeatures_API_HYBRID, gofeaturespb.GoFeatures_API_OPAQUE:
		return fmt.Errorf("%s: Edition Go API level %s is not supported by protobuf-go-lite", name, goFeatures.GetApiLevel())
	default:
		return fmt.Errorf("%s: unknown Edition Go API level %d", name, goFeatures.GetApiLevel())
	}
}
