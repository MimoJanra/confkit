package consul

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/hashicorp/consul/api"
)

type ConsulSource struct {
	client       *api.Client
	prefix       string
	token        string
	datacenter   string
	waitDuration time.Duration
}

func NewConsulSource(addr string, token string, datacenter string) (*ConsulSource, error) {
	config := api.DefaultConfig()
	config.Address = addr
	if token != "" {
		config.Token = token
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create Consul client: %w", err)
	}

	return &ConsulSource{
		client:       client,
		prefix:       "myapp",
		token:        token,
		datacenter:   datacenter,
		waitDuration: 5 * time.Minute,
	}, nil
}

func (c *ConsulSource) Name() string {
	return "consul"
}

func (c *ConsulSource) Lookup(_ context.Context, field *confkit.FieldInfo) (any, bool, error) {
	key := c.buildKey(field.Path)

	opts := &api.QueryOptions{
		Datacenter: c.datacenter,
	}

	kv, _, err := c.client.KV().Get(key, opts)
	if err != nil {
		return "", false, fmt.Errorf("cannot read Consul key %s: %w", key, err)
	}

	if kv == nil {
		return "", false, nil
	}

	return string(kv.Value), true, nil
}

func (c *ConsulSource) buildKey(fieldPath string) string {
	return fmt.Sprintf("%s/%s", c.prefix, strings.ToLower(fieldPath))
}

func (c *ConsulSource) SetPrefix(prefix string) {
	c.prefix = prefix
}

func (c *ConsulSource) SetWaitDuration(duration time.Duration) {
	c.waitDuration = duration
}

func FromConsul(addr string) confkit.Source {
	return FromConsulWithOptions(addr, "", "")
}

func FromConsulWithToken(addr string, token string) confkit.Source {
	return FromConsulWithOptions(addr, token, "")
}

func FromConsulWithOptions(addr string, token string, datacenter string) confkit.Source {
	src, err := NewConsulSource(addr, token, datacenter)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
