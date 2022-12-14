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

func WithRepository(repository rule.Repository) Option {
	return func(c *ruleEngine) {
		c.repository = repository
	}
}

func WithLogger(logger log.Logger) Option {
	return func(c *ruleEngine) {
		c.logger = logger
	}
}

func DefaultRuleEngine(driver rule.Driver, logger log.Logger) (Engine, func(), error) {
	repo, err := repository.NewRepository(driver, repository.WithLogger(logger))
	if err != nil {
		return nil, nil, err
	}
	return NewRuleEngine(
		WithLogger(logger),
		WithRepository(repo),
	)
}

func NewRuleEngine(opt ...Option) (Engine, func(), error) {
	c := &ruleEngine{
		logger: log.NewJSONLogger(os.Stdout),
	}
	for _, o := range opt {
		o(c)
	}
	if c.repository == nil {
		return nil, nil, errors.New("repository is nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := c.repository.Watch(ctx)
		if err != nil {
			_ = level.Error(c.logger).Log("msg", fmt.Errorf("repository watch error %w", err))
		}
	}()
	return c, func() {
		cancel()
	}, nil
}

func (d *ruleEngine) Of(ruleName string) Tenanter {
	return &ofRule{
		ruleName: ruleName,
		d:        d,
	}
}
