package confkit

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdSource struct {
	client  *clientv3.Client
	prefix  string
	timeout int
}

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

func (e *EtcdSource) Name() string {
	return "etcd"
}

func (e *EtcdSource) Lookup(field *FieldInfo) (any, bool, error) {
	key := e.buildKey(field.Path)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.timeout)*time.Second)
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

func (e *EtcdSource) SetTimeout(seconds int) {
	e.timeout = seconds
}

func (e *EtcdSource) Close() error {
	return e.client.Close()
}

func FromEtcd(endpoints []string) Source {
	return FromEtcdWithPrefix(endpoints, "/myapp/")
}

func FromEtcdWithPrefix(endpoints []string, prefix string) Source {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	src, err := NewEtcdSource(endpoints, prefix)
	if err != nil {
		return &errorSource{err: err}
	}
	return src
}
