package repository

import (
	"github.com/GGXXLL/rule"
)

type Container struct {
	RuleSet rule.Ruler
	KV      *rule.KeyValue
}
