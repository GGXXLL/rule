package repository

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/GGXXLL/rule/internal/entity"
	"github.com/GGXXLL/rule/msg"

	"github.com/GGXXLL/rule"
	"github.com/GGXXLL/rule/contract"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
)

// defaultRepository 专门为客户端提供的 defaultRepository，不具备自举性，可以只监听需要的规则
type defaultRepository struct {
	driver     rule.Driver
	logger     log.Logger
	containers map[string]*Container
	rwLock     sync.RWMutex
	regexp     *regexp.Regexp

	customNewRuleFuncMap map[string]rule.NewRulerFunc
	customCompileFuncMap map[string]rule.CompileFunc
	customNewRuleFunc    rule.NewRulerFunc
	customCompileFunc    rule.CompileFunc

	dispatcher contract.Dispatcher
}

type Option func(r *defaultRepository)

func WithRegex(regexp *regexp.Regexp) Option {
	return func(r *defaultRepository) {
		r.regexp = regexp
	}
}

func WithDispatcher(d contract.Dispatcher) Option {
	return func(r *defaultRepository) {
		r.dispatcher = d
	}
}

func WithCompileFunc(f rule.CompileFunc) Option {
	return func(r *defaultRepository) {
		r.customCompileFunc = f
	}
}

func WithCompileFuncMap(m map[string]rule.CompileFunc) Option {
	return func(r *defaultRepository) {
		r.customCompileFuncMap = m
	}
}

func WithRuleFunc(f rule.NewRulerFunc) Option {
	return func(r *defaultRepository) {
		r.customNewRuleFunc = f
	}
}

// WithRuleFuncMap replace the rule.Ruler implement
func WithRuleFuncMap(m map[string]rule.NewRulerFunc) Option {
	return func(r *defaultRepository) {
		r.customNewRuleFuncMap = m
	}
}

func WithLogger(l log.Logger) Option {
	return func(r *defaultRepository) {
		r.logger = l
	}
}

func NewRepository(driver rule.Driver, opts ...Option) (rule.Repository, error) {
	var repo = &defaultRepository{
		driver:     driver,
		logger:     log.NewJSONLogger(os.Stdout),
		containers: make(map[string]*Container),
		rwLock:     sync.RWMutex{},
	}

	for _, opt := range opts {
		opt(repo)
	}

	// 第一次拉取配置
	items, err := driver.All(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, msg.ErrorRules)
	}

	for _, item := range items {
		if repo.regexp != nil && !repo.regexp.MatchString(item.Key) {
			continue
		}

		c := Container{KV: item}

		c.RuleSet, err = repo.generateRuler(&c)
		if err != nil {
			_ = level.Error(repo.logger).Log("msg", fmt.Sprintf("%s generate rule error", item.Key), "err", err)
			continue
		}

		repo.containers[item.Key] = &c
		if repo.dispatcher != nil {
			_ = repo.dispatcher.Dispatch(context.Background(), item.Type, c)
		}
	}

	_ = level.Info(repo.logger).Log("msg", fmt.Sprintf("%d rules have been added", len(repo.containers)))

	return repo, nil
}

func (r *defaultRepository) getCustomNewRuleFunc(name string) rule.NewRulerFunc {
	if f, ok := r.customNewRuleFuncMap[name]; ok {
		return f
	}
	if r.customNewRuleFunc != nil {
		return r.customNewRuleFunc
	}
	return nil
}

func (r *defaultRepository) getCustomCompileFunc(name string) rule.CompileFunc {
	if f, ok := r.customCompileFuncMap[name]; ok {
		return f
	}
	if r.customCompileFunc != nil {
		return r.customCompileFunc
	}
	return nil
}

func (r *defaultRepository) generateRuler(c *Container) (ruler rule.Ruler, err error) {
	reader := bytes.NewReader(c.KV.Value)
	if customNewRuleFunc := r.getCustomNewRuleFunc(c.KV.Key); customNewRuleFunc != nil {
		ruler, err = customNewRuleFunc(reader)
		if err != nil {
			return nil, errors.New("invalid custom NewRuleFunc")
		}
	} else if customCompileFunc := r.getCustomCompileFunc(c.KV.Key); customCompileFunc != nil {
		ruler, err = entity.NewCustomRules(reader, customCompileFunc)
		if err != nil {
			return nil, errors.New("invalid custom CompileFunc")
		}
	} else {
		ruler, err = entity.NewRules(reader)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (r *defaultRepository) GetRuler(ruleName string) rule.Ruler {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	if c, ok := r.containers[ruleName]; ok {
		return c.RuleSet
	}
	return nil
}

func (r *defaultRepository) GetRaw(ruleName string) []byte {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	if c, ok := r.containers[ruleName]; ok {
		return c.KV.Value
	}
	return nil
}

func (r *defaultRepository) Watch(ctx context.Context) (err error) {
	ch := r.driver.Watch(ctx)
	for {
		select {
		case kv := <-ch:
			if kv.Err != nil {
				return kv.Err
			}
			// 匹配正则监听
			if r.regexp != nil && !r.regexp.MatchString(kv.Key) {
				continue
			}
			c := Container{KV: kv}
			if kv.Type == rule.EventTypeDelete || len(kv.Value) == 0 {
				r.deleteRuleSetByDbKey(kv.Key)
				if r.dispatcher != nil {
					_ = r.dispatcher.Dispatch(ctx, kv.Type, Container{KV: kv})
				}
				continue
			}
			c.RuleSet, err = r.generateRuler(&c)
			if err != nil {
				_ = level.Error(r.logger).Log("msg", fmt.Sprintf("%s generate rule error", kv.Key), "err", err)
				continue
			}
			r.updateRuleSet(&c)
			if r.dispatcher != nil {
				_ = r.dispatcher.Dispatch(ctx, kv.Type, c)
			}
			_ = level.Info(r.logger).Log("msg", fmt.Sprintf("配置已更新 %s", kv.Key))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *defaultRepository) Count() int {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	return len(r.containers)
}

func (r *defaultRepository) updateRuleSet(c *Container) bool {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	if _, ok := r.containers[c.KV.Key]; ok {
		r.containers[c.KV.Key] = c
		return true
	}
	return false
}

func (r *defaultRepository) deleteRuleSetByDbKey(key string) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	delete(r.containers, key)
}
