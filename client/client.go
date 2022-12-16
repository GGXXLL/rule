package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/GGXXLL/rule"
	"github.com/GGXXLL/rule/repository"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Option func(*ruleEngine)

type ruleEngine struct {
	logger     log.Logger
	repository rule.Repository
}

// WithRepository replace the rule.Repository
func WithRepository(repository rule.Repository) Option {
	return func(c *ruleEngine) {
		c.repository = repository
	}
}

// WithLogger replace the log.Logger
func WithLogger(logger log.Logger) Option {
	return func(c *ruleEngine) {
		c.logger = logger
	}
}

// DefaultRuleEngine returns Engine with rule.Driver and log.Logger.
// It will auto init rule.Repository and call Watch method.
func DefaultRuleEngine(driver rule.Driver, logger log.Logger) (Engine, func(), error) {
	repo, err := repository.NewRepository(driver, repository.WithLogger(logger))
	if err != nil {
		return nil, nil, err
	}
	engine, err := NewRuleEngine(
		WithLogger(logger),
		WithRepository(repo),
	)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := repo.Watch(ctx)
		if err != nil {
			_ = level.Error(logger).Log("msg", fmt.Errorf("repository watch error %w", err))
		}
	}()

	return engine, func() {
		cancel()
	}, nil
}

// NewRuleEngine returns Engine with Option.
func NewRuleEngine(opt ...Option) (Engine, error) {
	c := &ruleEngine{
		logger: log.NewJSONLogger(os.Stdout),
	}
	for _, o := range opt {
		o(c)
	}
	if c.repository == nil {
		return nil, errors.New("repository is nil")
	}
	return c, nil
}

func (d *ruleEngine) Of(ruleName string) Tenanter {
	return &ofRule{
		ruleName: ruleName,
		d:        d,
	}
}
