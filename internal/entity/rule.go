package entity

import (
	"fmt"
	"io"
	"strconv"

	"github.com/GGXXLL/rule"
	"github.com/antonmedv/expr/vm"

	"github.com/GGXXLL/rule/dto"
	"github.com/knadh/koanf"
	kyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func NewRuler(style string) (rule.Ruler, error) {
	switch style {
	case "advanced":
		return NewAdvancedRule(), nil
	case "basic":
		return NewBasicRule(), nil
	case "switch":
		return NewSwitchRule(), nil
	case "":
		return NewBasicRule(), nil
	default:
		return nil, fmt.Errorf("unsupported style %s", style)
	}
}

type Config struct {
	Style string       `yaml:"style"`
	Rules []rule.Ruler `yaml:"rule"`
}

type CentralRules struct {
	Style string `yaml:"style"`
	Rule  struct {
		List []struct {
			Name     string   `yaml:"name"`
			Icon     string   `yaml:"icon"`
			Path     string   `yaml:"path"`
			Tabs     []string `yaml:"tabs"`
			ID       string   `yaml:"id"`
			Children []struct {
				Name     string        `yaml:"name"`
				Icon     string        `yaml:"icon"`
				Path     string        `yaml:"path"`
				ID       string        `yaml:"id"`
				Tabs     []string      `yaml:"tabs"`
				Children []interface{} `yaml:"child"`
			} `yaml:"child"`
		} `yaml:"list"`
	} `yaml:"rule"`
}

type ErrInvalidRules struct {
	detail string
}

func (e *ErrInvalidRules) Error() string {
	return e.detail
}

// convert Yaml在反序列化时，会把字段反序列化成map[interface{}]interface{}
// 而这个结构在序列化json时会出错。
// 通过这个函数，把map[interface{}]interface{}用递归转为
// map[string]interface{}
func convert(i interface{}) dto.Data {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			if vv, ok := k.(int); ok {
				m2[strconv.Itoa(vv)] = convert(v)
			}
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i.(dto.Data)
}

func newRules(reader io.Reader) (rule.Ruler, error) {
	var (
		b   []byte
		err error
	)
	c := koanf.New(".")
	b, err = io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "reader is not valid")
	}

	err = c.Load(rawbytes.Provider(b), kyaml.Parser())
	if err != nil {
		return nil, errors.Wrap(err, "cannot load yaml")
	}

	ruler, err := NewRuler(c.String("style"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid rules")
	}
	err = ruler.Unmarshal(c)
	if err != nil {
		return nil, errors.Wrap(err, "invalid rules")
	}
	return ruler, nil
}

func NewRules(reader io.Reader) (rule.Ruler, error) {
	ruler, err := newRules(reader)
	if err != nil {
		return nil, err
	}
	err = ruler.Compile()
	if err != nil {
		return nil, errors.Wrap(err, "can't compile")
	}
	return ruler, nil
}

func NewCustomRules(reader io.Reader, compileFunc func(string) (*vm.Program, error)) (rule.Ruler, error) {
	ruler, err := newRules(reader)
	if err != nil {
		return nil, err
	}
	customRuler, ok := ruler.(rule.CustomRuler)
	if !ok {
		return nil, errors.New("this rule do not support custom")
	}
	err = customRuler.CompileWithFunc(compileFunc)
	if err != nil {
		return nil, errors.Wrap(err, "can't compile")
	}
	return ruler, nil
}

func ValidateRules(reader io.Reader) error {
	var tmp rule.Ruler

	value, err := io.ReadAll(reader)
	if err != nil {
		return &ErrInvalidRules{err.Error()}
	}

	c := koanf.New(".")
	err = c.Load(rawbytes.Provider(value), kyaml.Parser())
	if err != nil {
		return &ErrInvalidRules{err.Error()}
	}
	tmp, err = NewRuler(c.String("style"))
	if err != nil {
		return &ErrInvalidRules{err.Error()}
	}
	if err = tmp.Unmarshal(c); err != nil {
		return &ErrInvalidRules{err.Error()}
	}
	if err := tmp.Compile(); err != nil {
		return &ErrInvalidRules{err.Error()}
	}
	if err := runTests(tmp, c); err != nil {
		return &ErrInvalidRules{err.Error()}
	}
	if err := runSchemaValidation(tmp, c); err != nil {
		return &ErrInvalidRules{err.Error()}
	}
	return nil
}

func runTests(ruler rule.Ruler, c *koanf.Koanf) error {
	if !c.Exists("tests") {
		return nil
	}
	var tests TestCases
	if err := c.Unmarshal("tests", &tests); err != nil {
		return errors.Wrap(err, "unable to unmarshal tests")
	}
	if err := tests.Asserts(ruler, dto.NewDecoder()); err != nil {
		return errors.Wrap(err, "tests failed")
	}
	return nil
}

func runSchemaValidation(ruler rule.Ruler, c *koanf.Koanf) error {
	if !c.Exists("def") {
		return nil
	}
	var schemaStruct map[string]interface{}
	if err := c.Unmarshal("def", &schemaStruct); err != nil {
		return errors.Wrap(err, "unable to unmarshal def")
	}
	schema := gojsonschema.NewGoLoader(schemaStruct)
	if err := ruler.ValidateWithSchema(schema); err != nil {
		return errors.Wrap(err, "def failed")
	}
	return nil
}
