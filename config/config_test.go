package config

import (
	"testing"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/assert"
)

func TestKoanfAdapter_Route(t *testing.T) {
	t.Parallel()
	ka := prepareJSONTestSubject(t)
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
}

func TestKoanfAdapter_Bool(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.True(t, k.Bool("bool"))
}

func TestKoanfAdapter_String(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, "string", k.String("string"))
}

func TestKoanfAdapter_Strings(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, []string{"foo", "bar"}, k.Strings("strings"))
}

func TestKoanfAdapter_Float64(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, 1.0, k.Float64("float"))
}

func TestKoanfAdapter_Get(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, 1.0, k.Get("float"))
}

func TestKoanfAdapter_Duration(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, time.Second, k.Duration("duration_string"))
}

func TestKoanfAdapter_Unmarshal_Json(t *testing.T) {
	t.Parallel()
	ka := prepareJSONTestSubject(t)
	var target string
	err := ka.Unmarshal("foo.bar", &target)
	assert.NoError(t, err)
	assert.Equal(t, "baz", target)

	var r Duration
	err = ka.Unmarshal("duration_string", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Second})

	err = ka.Unmarshal("duration_number", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Nanosecond})
}

func TestKoanfAdapter_Unmarshal_Yaml(t *testing.T) {
	t.Parallel()
	ka := prepareYamlTestSubject(t)
	var target string
	err := ka.Unmarshal("foo.bar", &target)
	assert.NoError(t, err)
	assert.Equal(t, "baz", target)

	var r Duration
	err = ka.Unmarshal("duration_string", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Second})

	err = ka.Unmarshal("duration_number", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Nanosecond})
}

func TestMapAdapter_Route(t *testing.T) {
	t.Parallel()
	m := MapAdapter(
		map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		},
	)
	assert.Equal(t, MapAdapter(map[string]interface{}{
		"bar": "baz",
	}), m.Route("foo"))
	assert.Panics(t, func() {
		m.Route("foo2")
	})
}

func TestMapAdapter_Unmarshal(t *testing.T) {
	t.Parallel()
	m := MapAdapter(
		map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		},
	)
	var target map[string]interface{}
	err := m.Unmarshal("foo", &target)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"bar": "baz",
	}, target)
}

func TestUpgrade(t *testing.T) {
	var m MapAdapter = map[string]interface{}{"foo": "bar"}
	upgraded := WithAccessor(m)

	assert.Equal(t, float64(0), upgraded.Float64("foo"))
	assert.Equal(t, 0, upgraded.Int("foo"))
	assert.Equal(t, "bar", upgraded.String("foo"))
	assert.Equal(t, false, upgraded.Bool("foo"))
	assert.Equal(t, "bar", upgraded.Get("foo"))
	assert.Equal(t, []string{"bar"}, upgraded.Strings("foo"))
	assert.Equal(t, time.Duration(0), upgraded.Duration("foo"))
}

func prepareJSONTestSubject(t *testing.T) *KoanfAdapter {
	k := koanf.New(".")
	if err := k.Load(file.Provider("testdata/mock.json"), json.Parser()); err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	ka := KoanfAdapter{K: k}
	return &ka
}

func prepareYamlTestSubject(t *testing.T) *KoanfAdapter {
	k := koanf.New(".")
	if err := k.Load(file.Provider("testdata/mock.yaml"), yaml.Parser()); err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	ka := KoanfAdapter{K: k}
	return &ka
}
