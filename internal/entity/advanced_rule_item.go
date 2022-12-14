package entity

import (
	"fmt"

	"github.com/GGXXLL/rule"
	"github.com/GGXXLL/rule/msg"
	"github.com/antonmedv/expr/compiler"
	"github.com/antonmedv/expr/parser"

	"github.com/GGXXLL/rule/dto"
	"github.com/antonmedv/expr/vm"
	"github.com/hashicorp/go-multierror"
	"github.com/knadh/koanf"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type AdvancedRuleItem struct {
	cond    string
	then    dto.Data
	child   rule.Ruler
	program *vm.Program
}

func (ar *AdvancedRuleItem) ValidateWithSchema(schema gojsonschema.JSONLoader) error {
	if ar.then == nil && ar.child != nil {
		return ar.child.ValidateWithSchema(schema)
	}
	document := gojsonschema.NewGoLoader(ar.then)
	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return errors.Wrap(err, "fails to validate with json schema")
	}
	if !result.Valid() {
		var err multierror.Error
		for i := range result.Errors() {
			if result.Errors()[i] != nil {
				err.Errors = append(err.Errors, fmt.Errorf(result.Errors()[i].String()))
			}
		}
		if err.Len() > 0 {
			return &err
		}
		return nil
	}
	return nil
}

func (ar *AdvancedRuleItem) Unmarshal(reader *koanf.Koanf) error {
	ar.cond = reader.MustString("if")
	if len(ar.cond) == 0 {
		return errors.New("if condition not found in advanced rule")
	}
	err := reader.Unmarshal("then", &ar.then)
	if err != nil {
		return err
	}
	if ar.then == nil && reader.Exists("child") {
		style := reader.MustString("child.style")
		if style == "" {
			return errors.New("missing child style")
		}
		item, err := NewRuler(style)
		if err != nil {
			return err
		}
		err = item.Unmarshal(reader.Cut("child"))
		if err != nil {
			return err
		}
		ar.child = item
	}
	return nil
}

func (ar *AdvancedRuleItem) Compile() error {
	return ar.CompileWithFunc(func(s string) (*vm.Program, error) {
		tree, err := parser.Parse(s)
		if err != nil {
			return nil, err
		}
		return compiler.Compile(tree, nil)
	})
}

func (ar *AdvancedRuleItem) CompileWithFunc(compileFunc rule.CompileFunc) error {
	var err error
	ar.then = convert(ar.then)
	ar.program, err = compileFunc(ar.cond)
	if err != nil {
		return err
	}
	if ar.program == nil {
		return fmt.Errorf("invalid expression: %s", ar.cond)
	}
	if ar.child != nil {
		if err = ar.child.Compile(); err != nil {
			return err
		}
	}
	return err
}

func (ar *AdvancedRuleItem) Calculate(payload interface{}) (dto.Data, error) {
	output, err := vm.Run(ar.program, payload)
	if err != nil {
		return nil, errors.Wrap(err, msg.ErrorRules)
	}
	if i, ok := output.(int); ok && i == 0 {
		return nil, nil
	}
	if b, ok := output.(bool); ok && !b {
		return nil, nil
	}
	if ar.then != nil {
		return ar.then, nil
	}
	if ar.child != nil {
		return ar.child.Calculate(payload)
	}
	return nil, nil
}
