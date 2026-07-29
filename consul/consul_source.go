package consul

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/hashicorp/consul/api"
)

// ConsulSource reads configuration values from the Consul key-value store.
//
// Keys are the field's dotted path lower-cased and joined to the prefix with a slash, so
// DB.Host under prefix "myapp" reads "myapp/db.host".
type ConsulSource struct {
	client       *api.Client
	prefix       string
	token        string
	datacenter   string
	waitDuration time.Duration
}

// NewConsulSource returns a source talking to the Consul agent at addr. An empty token
// or datacenter leaves the client default in place. The key prefix defaults to "myapp";
// change it with SetPrefix.
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

// Name returns "consul".
func (c *ConsulSource) Name() string {
	return "consul"
}

// Lookup fetches the key for field from the KV store. A missing key means not found
// rather than an error.
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

// SetPrefix sets the key prefix. Call it before loading, as it is not safe to change
// concurrently with Lookup.
func (c *ConsulSource) SetPrefix(prefix string) {
	c.prefix = prefix
}

// SetWaitDuration sets the blocking-query wait duration used for change watching.
func (c *ConsulSource) SetWaitDuration(duration time.Duration) {
	c.waitDuration = duration
}

// FromConsul reads from the Consul agent at addr with no ACL token and the agent's own
// datacenter.
func FromConsul(addr string) confkit.Source {
	return FromConsulWithOptions(addr, "", "")
}

// FromConsulWithToken reads from Consul using an ACL token.
func FromConsulWithToken(addr string, token string) confkit.Source {
	return FromConsulWithOptions(addr, token, "")
}

// FromConsulWithOptions reads from Consul with an optional ACL token and datacenter;
// empty values keep the client defaults.
//
// A client-creation failure is not reported here: the returned Source fails every
// lookup, so the problem appears in the load's ErrorReport.
func FromConsulWithOptions(addr string, token string, datacenter string) confkit.Source {
	src, err := NewConsulSource(addr, token, datacenter)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
