package rule

import (
	"testing"

	"github.com/antonmedv/expr"
)

func TestEval(t *testing.T) {
	r, err := expr.Eval("name == 1", map[string]interface{}{
		"name": 1,
	})
	t.Log(r, err)
}
