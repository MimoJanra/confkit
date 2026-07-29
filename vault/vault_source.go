package vault

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MimoJanra/confkit"
	"github.com/hashicorp/vault/api"
)

// VaultAuth obtains a Vault client token. Implementations are provided by
// VaultTokenAuth, VaultAppRoleAuth and VaultKubernetesAuth.
type VaultAuth interface {
	Authenticate(ctx context.Context, client *api.Client) (token string, err error)
}

// VaultSource reads secrets from HashiCorp Vault's KV engine.
//
// The token obtained from the configured VaultAuth is cached and reused across lookups,
// and re-authentication is performed when it expires.
type VaultSource struct {
	addr       string
	auth       VaultAuth
	kvVersion  int
	pathPrefix string
	client     *api.Client
	token      string
	tokenMu    sync.RWMutex
	lastAuth   time.Time
}

// NewVaultSource returns a source reading from the Vault server at addr under
// pathPrefix, authenticating with auth. kvVersion selects the KV engine version, 1 or 2,
// which determines how the secret payload is nested.
func NewVaultSource(addr string, auth VaultAuth, kvVersion int, pathPrefix string) (*VaultSource, error) {
	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create Vault client: %w", err)
	}

	return &VaultSource{
		addr:       addr,
		auth:       auth,
		kvVersion:  kvVersion,
		pathPrefix: pathPrefix,
		client:     client,
	}, nil
}

// Name returns "vault".
func (v *VaultSource) Name() string {
	return "vault"
}

// Lookup authenticates if necessary, then reads the secret for field and returns the
// entry matching the field's name. A missing secret or key means not found rather than
// an error.
func (v *VaultSource) Lookup(ctx context.Context, field *confkit.FieldInfo) (any, bool, error) {
	token, err := v.ensureAuthed(ctx)
	if err != nil {
		return "", false, err
	}

	v.client.SetToken(token)

	path := v.buildPath(field.Path)
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return "", false, fmt.Errorf("cannot read Vault secret %s: %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return "", false, nil
	}

	if data, ok := secret.Data["data"].(map[string]any); ok && v.kvVersion == 2 {
		value, exists := data[field.Name]
		return value, exists, nil
	}

	value, exists := secret.Data[field.Name]
	return value, exists, nil
}

func (v *VaultSource) ensureAuthed(ctx context.Context) (string, error) {
	v.tokenMu.RLock()
	if v.token != "" && time.Since(v.lastAuth) < 1*time.Hour {
		defer v.tokenMu.RUnlock()
		return v.token, nil
	}
	v.tokenMu.RUnlock()

	token, err := v.auth.Authenticate(ctx, v.client)
	if err != nil {
		return "", err
	}

	v.tokenMu.Lock()
	v.token = token
	v.lastAuth = time.Now()
	v.tokenMu.Unlock()

	return token, nil
}

func (v *VaultSource) buildPath(fieldPath string) string {
	prefix := v.pathPrefix
	if prefix == "" {
		prefix = "myapp"
	}
	if v.kvVersion == 2 {
		return fmt.Sprintf("secret/data/%s/%s", prefix, fieldPath)
	}
	return fmt.Sprintf("secret/%s/%s", prefix, fieldPath)
}

type vaultTokenAuth struct {
	token string
}

// VaultTokenAuth authenticates with a token supplied directly, as in development or when
// a token is already available from the environment.
func VaultTokenAuth(token string) VaultAuth {
	return &vaultTokenAuth{token: token}
}

func (a *vaultTokenAuth) Authenticate(_ context.Context, _ *api.Client) (string, error) {
	return a.token, nil
}

type vaultAppRoleAuth struct {
	roleID   string
	secretID string
}

// VaultAppRoleAuth authenticates through the AppRole backend at auth/approle/login,
// exchanging a role ID and secret ID for a token.
func VaultAppRoleAuth(roleID, secretID string) VaultAuth {
	return &vaultAppRoleAuth{roleID: roleID, secretID: secretID}
}

func (a *vaultAppRoleAuth) Authenticate(ctx context.Context, client *api.Client) (string, error) {
	path := "auth/approle/login"
	data := map[string]any{
		"role_id":   a.roleID,
		"secret_id": a.secretID,
	}

	secret, err := client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return "", fmt.Errorf("approle authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return "", fmt.Errorf("no auth in Vault response")
	}

	return secret.Auth.ClientToken, nil
}

type vaultKubernetesAuth struct {
	role string
	jwt  string
}

// VaultKubernetesAuth authenticates through the Kubernetes backend at
// auth/kubernetes/login, exchanging a role and a service-account JWT for a token. The
// JWT is normally read from
// /var/run/secrets/kubernetes.io/serviceaccount/token.
func VaultKubernetesAuth(role, jwt string) VaultAuth {
	return &vaultKubernetesAuth{role: role, jwt: jwt}
}

func (a *vaultKubernetesAuth) Authenticate(ctx context.Context, client *api.Client) (string, error) {
	path := "auth/kubernetes/login"
	data := map[string]any{
		"role": a.role,
		"jwt":  a.jwt,
	}

	secret, err := client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return "", fmt.Errorf("kubernetes authentication failed: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return "", fmt.Errorf("no auth in Vault response")
	}

	return secret.Auth.ClientToken, nil
}

// FromVault reads secrets from Vault's KV version 2 engine.
func FromVault(addr string, auth VaultAuth, pathPrefix string) confkit.Source {
	return FromVaultWithKVVersion(addr, auth, 2, pathPrefix)
}

// FromVaultWithKVVersion reads secrets from the given KV engine version; anything other
// than 1 or 2 is treated as 2.
//
// A client-creation failure is not reported here: the returned Source fails every lookup,
// so the problem appears in the load's ErrorReport.
func FromVaultWithKVVersion(addr string, auth VaultAuth, kvVersion int, pathPrefix string) confkit.Source {
	if kvVersion != 1 && kvVersion != 2 {
		kvVersion = 2
	}

	src, err := NewVaultSource(addr, auth, kvVersion, pathPrefix)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
