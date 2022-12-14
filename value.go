package rule

import (
	"context"
	"fmt"
	"io"

	"github.com/GGXXLL/rule/dto"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/knadh/koanf"
	"github.com/xeipuuv/gojsonschema"
)

type KeyValue struct {
	Key   string
	Value []byte
	Type  EventType
	Err   error
}

type KvWatchChan <-chan *KeyValue

type Driver interface {
	One(ctx context.Context, key string) ([]byte, error)
	All(ctx context.Context) ([]*KeyValue, error)
	Watch(ctx context.Context) KvWatchChan
}

type RulerOption func()

type Ruler interface {
	Unmarshal(reader *koanf.Koanf) error
	Calculate(payload interface{}) (dto.Data, error)
	Compile() error
	ValidateWithSchema(schema gojsonschema.JSONLoader) error
}

type CustomRuler interface {
	Ruler
	CompileWithFunc(compileFunc CompileFunc) error
}

type Repository interface {
	// GetRuler returns the Ruler
	GetRuler(ruleName string) Ruler
	// GetRaw returns the original value
	GetRaw(ruleName string) []byte
	// Watch start watch changes
	Watch(ctx context.Context) error
	// Count returns the number of cached rules
	Count() int
}

func Calculate(rules Ruler, env interface{}) (dto.Data, error) {
	if _, ok := env.(expr.Option); ok {
		return nil, fmt.Errorf("misused expr.Eval: second argument (env) should be passed without expr.Env")
	}
	return rules.Calculate(env)
}

type NewRulerFunc func(reader io.Reader) (Ruler, error)
type CompileFunc func(string) (*vm.Program, error)
