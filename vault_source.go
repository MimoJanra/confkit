package confkit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

type VaultAuth interface {
	Authenticate(ctx context.Context, client *api.Client) (token string, err error)
}

type VaultSource struct {
	addr      string
	auth      VaultAuth
	kvVersion int
	client    *api.Client
	token     string
	tokenMu   sync.RWMutex
	lastAuth  time.Time
}

func NewVaultSource(addr string, auth VaultAuth, kvVersion int) (*VaultSource, error) {
	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create Vault client: %w", err)
	}

	return &VaultSource{
		addr:      addr,
		auth:      auth,
		kvVersion: kvVersion,
		client:    client,
	}, nil
}

func (v *VaultSource) Name() string {
	return "vault"
}

func (v *VaultSource) Lookup(field *FieldInfo) (any, bool, error) {
	token, err := v.ensureAuthed(context.Background())
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
	if v.kvVersion == 2 {
		return fmt.Sprintf("secret/data/myapp/%s", fieldPath)
	}
	return fmt.Sprintf("secret/myapp/%s", fieldPath)
}

type vaultTokenAuth struct {
	token string
}

func VaultTokenAuth(token string) VaultAuth {
	return &vaultTokenAuth{token: token}
}

func (a *vaultTokenAuth) Authenticate(ctx context.Context, client *api.Client) (string, error) {
	return a.token, nil
}

type vaultAppRoleAuth struct {
	roleID   string
	secretID string
}

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
		return "", fmt.Errorf("AppRole authentication failed: %w", err)
	}

	if secret.Auth == nil {
		return "", fmt.Errorf("no auth in Vault response")
	}

	return secret.Auth.ClientToken, nil
}

type vaultKubernetesAuth struct {
	role string
	jwt  string
}

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
		return "", fmt.Errorf("Kubernetes authentication failed: %w", err)
	}

	if secret.Auth == nil {
		return "", fmt.Errorf("no auth in Vault response")
	}

	return secret.Auth.ClientToken, nil
}

func FromVault(addr string, auth VaultAuth) Source {
	return FromVaultWithKVVersion(addr, auth, 2)
}

func FromVaultWithKVVersion(addr string, auth VaultAuth, kvVersion int) Source {
	if kvVersion != 1 && kvVersion != 2 {
		kvVersion = 2
	}

	src, err := NewVaultSource(addr, auth, kvVersion)
	if err != nil {
		return &errorSource{err: err}
	}
	return src
}
