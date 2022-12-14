package client

import (
	"fmt"

	"github.com/GGXXLL/rule"
	"github.com/GGXXLL/rule/config"
	"github.com/GGXXLL/rule/contract"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/pkg/errors"
)

type ofRule struct {
	d        *ruleEngine
	ruleName string
}

func (r *ofRule) Payload(pl interface{}) (contract.ConfigAccessor, error) {
	ruler := r.d.repository.GetRuler(r.ruleName)
	if ruler == nil {
		return nil, fmt.Errorf("no suitable configuration found for %s", r.ruleName)
	}

	calculated, err := rule.Calculate(ruler, pl)
	if err != nil {
		return nil, err
	}

	c, err := config.NewConfig(config.WithProviderLayer(confmap.Provider(calculated, "."), nil))
	if err != nil {
		return nil, errors.Wrap(err, "cannot load from map")
	}
	return c, nil
}
