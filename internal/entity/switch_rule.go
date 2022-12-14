package entity

import (
	"fmt"

	"github.com/GGXXLL/rule"

	"github.com/GGXXLL/rule/dto"
	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"github.com/knadh/koanf"
	"github.com/xeipuuv/gojsonschema"
)

type SwitchRule struct {
	style    string
	by       string
	cases    map[string]rule.Ruler
	fallback rule.Ruler
}

func NewSwitchRule() *SwitchRule {
	return &SwitchRule{
		style: "switch",
		cases: make(map[string]rule.Ruler),
	}
}

func (s *SwitchRule) ValidateWithSchema(schema gojsonschema.JSONLoader) error {
	var err multierror.Error
	for i := range s.cases {
		errors := s.cases[i].ValidateWithSchema(schema)
		if errors != nil {
			err.Errors = append(err.Errors, errors)
		}

	}
	errors := s.fallback.ValidateWithSchema(schema)
	if errors != nil {
		err.Errors = append(err.Errors, errors)
	}
	if err.Len() > 0 {
		return &err
	}
	return nil
}

func (s *SwitchRule) Unmarshal(reader *koanf.Koanf) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %s", r)
		}
	}()

	s.style = reader.String("style")
	s.by = reader.MustString("by")
	cases := reader.Slices("rule")
	for i := len(cases) - 1; i >= 0; i-- {
		style := cases[i].String("style")
		s.cases[cases[i].MustString("case")], err = NewRuler(style)
		if err != nil {
			return err
		}
		err = s.cases[cases[i].MustString("case")].Unmarshal(cases[i])
		if err != nil {
			return err
		}
	}
	style := reader.String("default.style")
	s.fallback, err = NewRuler(style)
	if err != nil {
		return err
	}
	err = s.fallback.Unmarshal(reader.Cut("default"))
	if err != nil {
		return err
	}

	return nil
}

func (s *SwitchRule) Calculate(payload interface{}) (dto.Data, error) {
	m, ok := payload.(dto.Payload)
	if !ok {
		m = structs.Map(payload)
	}
	by, ok := m[s.by]
	if !ok {
		return nil, fmt.Errorf("switch by non-exist key %s", s.by)
	}
	byStr, ok := by.(string)
	if !ok {
		return nil, fmt.Errorf("can only switch by string type, got: %s", s.by)
	}
	c, ok := s.cases[byStr]
	if !ok {
		if s.fallback == nil {
			return dto.Data{}, nil
		}
		return s.fallback.Calculate(payload)
	}
	return c.Calculate(payload)
}

func (s *SwitchRule) Compile() error {
	for i := range s.cases {
		if err := s.cases[i].Compile(); err != nil {
			return err
		}
	}
	if s.fallback == nil {
		return nil
	}
	return s.fallback.Compile()
}
