package entity

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/GGXXLL/rule"

	"github.com/GGXXLL/rule/dto"
	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
)

type Given struct {
	Method string `json:"method" yaml:"method"`
	URL    string `json:"url" yaml:"url"`
	Body   string `json:"body" yaml:"url"`
}

type TestCase struct {
	Given  Given  `json:"given" yaml:"given"`
	Expect string `json:"expect" yaml:"expect"`
}

func (t *TestCase) applyDefaults() {
	if t.Given.URL == "" {
		t.Given.URL = "http://baidu.com"
	}
	if t.Given.Method == "" {
		t.Given.Method = http.MethodGet
	}
	if t.Expect == "" {
		t.Expect = "true"
	}
}

type Decoder interface {
	Decode(payload interface{}, r *http.Request) (err error)
}

func (t TestCase) Asserts(ruler rule.Ruler, decoder Decoder) error {
	t.applyDefaults()
	req, err := http.NewRequest(t.Given.Method, t.Given.URL, strings.NewReader(t.Given.Body))
	if err != nil {
		return errors.Wrap(err, "unable to parse \"given\"")
	}

	var payload = make(dto.Payload)
	err = decoder.Decode(payload, req)
	if err != nil {
		return errors.Wrapf(err, "unable to decode querystring: %s", req.URL.RawQuery)
	}

	data, err := ruler.Calculate(payload)
	if err != nil {
		return errors.Wrap(err, "unable to calculate payload")
	}

	output, err := expr.Eval(t.Expect, data)
	if err != nil {
		return errors.Wrap(err, "fails to compile \"expect\"")
	}
	pass, ok := output.(bool)
	if !ok {
		return fmt.Errorf("\"expect\" should return a boolean, got %T", output)
	}
	if !pass {
		return fmt.Errorf("given %s, expects %s to be true, but it is false", t.Given, t.Expect)
	}
	return nil
}

type TestCases []TestCase

func (t TestCases) Asserts(ruler rule.Ruler, decoder Decoder) error {
	for i := range t {
		err := t[i].Asserts(ruler, decoder)
		if err != nil {
			return errors.Wrapf(err, "no.%d", i)
		}
	}
	return nil
}
