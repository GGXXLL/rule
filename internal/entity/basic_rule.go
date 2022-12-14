package entity

import (
	"fmt"

	"github.com/GGXXLL/rule/dto"
	"github.com/hashicorp/go-multierror"
	"github.com/knadh/koanf"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type BasicRule struct {
	style string
	data  dto.Data `yaml:"then"`
}

func (br *BasicRule) ValidateWithSchema(schema gojsonschema.JSONLoader) error {
	document := gojsonschema.NewGoLoader(br.data)
	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return errors.Wrap(err, "fails to validate with json schema")
	}
	if !result.Valid() {
		var err multierror.Error
		for i := range result.Errors() {
			err.Errors = append(err.Errors, fmt.Errorf(result.Errors()[i].String()))
		}
		return &err
	}
	return nil
}

func NewBasicRule() *BasicRule {
	return &BasicRule{style: "basic", data: dto.Data{}}
}

func (br *BasicRule) Unmarshal(reader *koanf.Koanf) error {
	br.style = reader.String("style")
	err := reader.Unmarshal("rule", &br.data)
	if err != nil {
		return err
	}
	return nil
}

func (br *BasicRule) Compile() error {
	br.data = convert(br.data)
	return nil
}

func (br *BasicRule) Calculate(interface{}) (dto.Data, error) {
	return br.data, nil
}
