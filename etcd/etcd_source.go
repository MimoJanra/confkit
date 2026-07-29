package etcd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MimoJanra/confkit"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdSource reads configuration values from an etcd v3 key-value store.
//
// Keys are the field's dotted path lower-cased and joined to the prefix, so DB.Host
// under prefix "/myapp/" reads "/myapp/db.host".
type EtcdSource struct {
	client  *clientv3.Client
	prefix  string
	timeout int
}

// NewEtcdSource dials endpoints and returns a source reading keys under prefix. The
// per-lookup timeout defaults to five seconds; change it with SetTimeout. Close the
// source when finished to release the client.
func NewEtcdSource(endpoints []string, prefix string) (*EtcdSource, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create etcd client: %w", err)
	}

	return &EtcdSource{
		client:  client,
		prefix:  prefix,
		timeout: 5,
	}, nil
}

// Name returns "etcd".
func (e *EtcdSource) Name() string {
	return "etcd"
}

// Lookup fetches the key for field, bounded by the configured timeout. A key that does
// not exist means not found rather than an error.
func (e *EtcdSource) Lookup(ctx context.Context, field *confkit.FieldInfo) (any, bool, error) {
	key := e.buildKey(field.Path)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(e.timeout)*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return "", false, fmt.Errorf("cannot read etcd key %s: %w", key, err)
	}

	if resp.Count == 0 {
		return "", false, nil
	}

	if len(resp.Kvs) > 0 {
		return string(resp.Kvs[0].Value), true, nil
	}

	return "", false, nil
}

func (e *EtcdSource) buildKey(fieldPath string) string {
	return fmt.Sprintf("%s%s", e.prefix, strings.ToLower(fieldPath))
}

// SetTimeout sets the per-lookup timeout in seconds. Call it before loading, as it is
// not safe to change concurrently with Lookup.
func (e *EtcdSource) SetTimeout(seconds int) {
	e.timeout = seconds
}

// Close releases the underlying etcd client.
func (e *EtcdSource) Close() error {
	return e.client.Close()
}

// FromEtcd reads keys under the default "/myapp/" prefix. Use FromEtcdWithPrefix to
// choose your own.
func FromEtcd(endpoints []string) confkit.Source {
	return FromEtcdWithPrefix(endpoints, "/myapp/")
}

// FromEtcdWithPrefix reads keys under prefix, appending a trailing slash if absent.
//
// A connection failure is not reported here: the returned Source fails every lookup, so
// the problem appears in the load's ErrorReport. Note that the returned Source cannot be
// closed; use NewEtcdSource when you need to release the client.
func FromEtcdWithPrefix(endpoints []string, prefix string) confkit.Source {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	src, err := NewEtcdSource(endpoints, prefix)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
