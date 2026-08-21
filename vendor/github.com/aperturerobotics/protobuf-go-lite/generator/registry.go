package generator

import (
	"cmp"
	"errors"
	"slices"
	"strconv"

	"github.com/aperturerobotics/protobuf-go-lite/compiler/protogen"
)

const registryPackage = protogen.GoImportPath("github.com/aperturerobotics/protobuf-go-lite/registry")

func (gen *Generator) generateRegistry(p *GeneratedFile, file *protogen.File) {
	var messages []*protogen.Message
	var collect func([]*protogen.Message)
	collect = func(msgs []*protogen.Message) {
		for _, message := range msgs {
			collect(message.Messages)
			if message.Desc.IsMapEntry() {
				continue
			}
			messages = append(messages, message)
		}
	}
	collect(file.Messages)
	if len(messages) == 0 {
		return
	}
	slices.SortFunc(messages, func(a, b *protogen.Message) int {
		return cmp.Compare(a.Desc.FullName(), b.Desc.FullName())
	})

	registryIdent := p.QualifiedGoIdent(registryPackage.Ident("Register"))
	entryIdent := p.QualifiedGoIdent(registryPackage.Ident("Entry"))
	optionIdent := p.QualifiedGoIdent(registryPackage.Ident("Option"))
	messageIdent := p.QualifiedGoIdent(protogen.ProtobufGoLitePackage.Ident("Message"))

	p.P("func init() {")
	for _, message := range messages {
		fullName := string(message.Desc.FullName())
		typeURL := "type.googleapis.com/" + fullName
		opts := collectMessageOptions(file, message)

		p.P(registryIdent, "(", entryIdent, "{")
		p.P("FullName: ", strconv.Quote(fullName), ",")
		p.P("TypeURL: ", strconv.Quote(typeURL), ",")
		p.P("New: func() ", messageIdent, " {")
		p.P("return new(", message.GoIdent.GoName, ")")
		p.P("},")
		if len(opts) > 0 {
			p.P("Options: []", optionIdent, "{")
			for _, opt := range opts {
				p.P("{")
				p.P("Name: ", strconv.Quote(opt.Name), ",")
				p.P("Values: []string{")
				for _, v := range opt.Values {
					p.P(strconv.Quote(v), ",")
				}
				p.P("},")
				p.P("},")
			}
			p.P("},")
		}
		p.P("})")
	}
	p.P("}")
	p.P()
}

func validateRegistryFeatures(featureNames []string) error {
	all := slices.Contains(featureNames, "all")
	for _, required := range []string{"size", "marshal", "unmarshal"} {
		if all {
			if _, ok := defaultFeatures[required]; ok {
				continue
			}
		} else if slices.Contains(featureNames, required) {
			continue
		}
		return errRegistryFeatures
	}
	return nil
}

var errRegistryFeatures = errors.New("registry=true requires size, marshal, and unmarshal features")
