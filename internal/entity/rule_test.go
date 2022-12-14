package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRules(t *testing.T) {
	cases := []struct {
		name    string
		rule    string
		asserts func(t *testing.T, err error)
	}{
		{
			"valid",
			`
style: advanced
rule:
  - if: name > 0
    then:
      sms: 1
`,
			func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			"with success tests",
			`
style: advanced
tests:
  - given: 
      url: http://monetization.tagtic.cn?name=foo
    expect: sms == 1
rule:
  - if: name == "foo"
    then:
      sms: 1
`,
			func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			"with failed tests",
			`
style: advanced
tests:
  - given: 
      url: http://monetization.tagtic.cn?name=bar
    expect: sms == 1
rule:
  - if: name == "foo"
    then:
      sms: 1
`,
			func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			"with success schema",
			`
def:
  type: object
  required:
    - foo
  properties:
    foo:
      type: string
style: switch
by: name
rule:
  - case: foo
    style: basic
    rule:
      foo: bar
  - case: bar
    style: advanced
    rule:
      - if: false
        child:
          style: basic
          rule:
            foo: baz
      - if: true
        then:
          foo: qux
default:
  style: basic
  rule:
    foo: bar`,
			func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			"with failed schema",
			`
def:
  type: object
  required:
    - foo
  properties:
    foo:
      type: string
style: switch
by: name
rule:
  - case: foo
    style: basic
    rule:
      foo: bar
  - case: bar
    style: advanced
    rule:
      - if: false
        child:
          style: basic
          rule:
            foot: baz
      - if: true
        then:
          foo: qux
default:
  style: basic
  rule:
    foo: bar`,
			func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateRules(strings.NewReader(c.rule))
			c.asserts(t, err)
		})
	}
}
