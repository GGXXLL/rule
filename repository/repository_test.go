package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"

	"github.com/GGXXLL/rule"
	"github.com/GGXXLL/rule/dto"
)

type mockDriver struct {
	ch chan *rule.KeyValue
}

func (r *mockDriver) One(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (r *mockDriver) All(ctx context.Context) ([]*rule.KeyValue, error) {
	return []*rule.KeyValue{
		{Key: "a", Value: []byte(`
style: advanced
rule:
  - if: name == "a"
    then:
      age: 1
`)},
		{Key: "b", Value: []byte(`
style: advanced
rule:
  - if: Name == "b"
    then:
      age: 2
`)},
		{Key: "c", Value: []byte(`
style: advanced
rule:
  - if: Name == "c"
    then:
      age: 3
`)},
		{Key: "d", Value: []byte(`
style: advanced
rule:
  - if: Name == "d"
    then:
      age: 4
`)},
		{Key: "e", Value: []byte(`
style: advanced
rule:
  - if: Foo == "e"
    then:
      age: 5
`)},
		{Key: "f", Value: []byte(`
style: advanced
rule:
  - if: Name == "f"
    then:
      age: 6
`)},
	}, nil

}

func (r *mockDriver) Watch(ctx context.Context) rule.KvWatchChan {
	<-ctx.Done()
	return nil
}

func TestRepository(t *testing.T) {
	type payload struct {
		Name string
	}
	rp, _ := regexp.Compile("[a-e]")
	cf := func(s string) (*vm.Program, error) {
		return expr.Compile(s, expr.Env(payload{}), expr.AllowUndefinedVariables())
	}

	repo, _ := NewRepository(&mockDriver{},
		WithLogger(log.NewNopLogger()),
		WithRegex(rp),
		WithCompileFuncMap(map[string]rule.CompileFunc{
			"c": cf,
			"d": cf,
			"e": cf,
		}),
	)

	// a-e are loaded
	assert.Equal(t, 5, repo.Count())

	ctx, cancel := context.WithCancel(context.Background())
	go repo.Watch(ctx)

	cases := []struct {
		key     string
		payload interface{}
		want    int
	}{
		{key: "a", payload: dto.Payload{"name": "a"}, want: 1},
		{key: "b", payload: payload{Name: "b"}, want: 2},
		{key: "c", payload: dto.Payload{"Name": "c"}, want: 3},
		{key: "d", payload: payload{Name: "d"}, want: 4},
		{key: "e", payload: struct{ Foo string }{Foo: "e"}, want: 5},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			r := repo.GetRuler(c.key)
			d, err := r.Calculate(c.payload)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, c.want, d["age"])
		})
	}

	cancel()
}
