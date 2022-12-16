package driver

import (
	"context"
	"regexp"
	"sync"

	"github.com/GGXXLL/rule"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdDriver struct {
	client *clientv3.Client
	prefix string
	regexp *regexp.Regexp
	rev    int64

	ch chan *rule.KeyValue

	once sync.Once

	limit int64
}

type Option func(driver *EtcdDriver)

// WithPrefix specifies the path prefix to listen.
func WithPrefix(p string) Option {
	return func(driver *EtcdDriver) {
		driver.prefix = p
	}
}

// WithLimit set the clientv3.WithLimit.
func WithLimit(limit int64) Option {
	return func(driver *EtcdDriver) {
		driver.limit = limit
	}
}

// WithRegex filter the path to listen by regexp.
func WithRegex(regexp *regexp.Regexp) Option {
	return func(r *EtcdDriver) {
		r.regexp = regexp
	}
}

func NewEtcdDriver(client *clientv3.Client, opts ...Option) *EtcdDriver {
	d := &EtcdDriver{
		client: client,
		rev:    0,
		ch:     make(chan *rule.KeyValue),
		limit:  1000,
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.prefix == "" {
		panic("etcd driver must define prefix")
	}

	return d
}

func (r *EtcdDriver) One(ctx context.Context, key string) ([]byte, error) {
	resp, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		return ev.Value, nil
	}
	return nil, nil
}

// All returns all kvs that conform to the specified prefix and regular expression.
func (r *EtcdDriver) All(ctx context.Context) ([]*rule.KeyValue, error) {
	kvs := make([]*rule.KeyValue, 0)
	key := r.prefix
	for {
		resp, err := r.client.Get(ctx, key, clientv3.WithRange(clientv3.GetPrefixRangeEnd(r.prefix)), clientv3.WithLimit(r.limit))
		if err != nil {
			return nil, errors.Wrapf(err, "prefix not found %s", r.prefix)
		}
		if r.rev == 0 {
			r.rev = resp.Header.Revision
		}
		for _, ev := range resp.Kvs {
			if r.regexp != nil && !r.regexp.Match(ev.Key) {
				continue
			}
			kvs = append(kvs, &rule.KeyValue{
				Key:   string(ev.Key),
				Value: ev.Value,
			})
		}
		if !resp.More {
			return kvs, nil
		}
		// move to next key
		key = string(append(resp.Kvs[len(resp.Kvs)-1].Key, 0))
	}
}

func (r *EtcdDriver) Watch(ctx context.Context) rule.KvWatchChan {
	r.once.Do(func() {
		go r.watch(ctx)
	})
	return r.ch
}

func (r *EtcdDriver) watch(ctx context.Context) {
	defer close(r.ch)
	rch := r.client.Watch(ctx, r.prefix, clientv3.WithPrefix(), clientv3.WithRev(r.rev))
	for {
		select {
		case wr := <-rch:
			if wr.Err() != nil {
				r.ch <- &rule.KeyValue{
					Err: wr.Err(),
				}
				return
			}
			for _, ev := range wr.Events {
				if r.regexp != nil && !r.regexp.Match(ev.Kv.Key) {
					continue
				}
				kv := &rule.KeyValue{Key: string(ev.Kv.Key), Value: ev.Kv.Value}
				if ev.Type == clientv3.EventTypeDelete {
					kv.Type = rule.EventTypeDelete
				}
				r.ch <- kv
			}
		case <-ctx.Done():
			return
		}
	}
}
