package client

import (
	"github.com/GGXXLL/rule/contract"
)

type Tenanter interface {
	Payload(pl interface{}) (contract.ConfigAccessor, error)
}

type Engine interface {
	Of(ruleName string) Tenanter
}
