package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MimoJanra/confkit"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// AWSSecretsManagerSource reads configuration values from a single AWS Secrets Manager
// secret.
//
// The secret is fetched once and cached for the configured TTL. Its value is expected to be
// a JSON object whose keys are field paths; if it is not valid JSON, the whole payload is
// stored under the key "_raw", as is a binary secret.
type AWSSecretsManagerSource struct {
	secretName  string
	region      string
	client      *secretsmanager.Client
	cache       map[string]any
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	lastCacheAt time.Time
}

// NewAWSSecretsManagerSource returns a source reading secretName. An empty region keeps the
// region from the ambient AWS configuration.
//
// Credentials are resolved by the AWS SDK's default chain and are not verified here, so an
// authentication problem surfaces on the first lookup.
func NewAWSSecretsManagerSource(secretName string, region string, cacheTTL time.Duration) (*AWSSecretsManagerSource, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	if region != "" {
		cfg.Region = region
	}

	return &AWSSecretsManagerSource{
		secretName: secretName,
		region:     region,
		client:     secretsmanager.NewFromConfig(cfg),
		cache:      make(map[string]any),
		cacheTTL:   cacheTTL,
	}, nil
}

// Name returns "aws-secrets-manager".
func (a *AWSSecretsManagerSource) Name() string {
	return "aws-secrets-manager"
}

// Lookup refreshes the secret if the cache has expired, then returns the entry whose key
// equals the field's dotted path. A key that is absent means not found rather than an error.
func (a *AWSSecretsManagerSource) Lookup(ctx context.Context, field *confkit.FieldInfo) (any, bool, error) {
	if err := a.ensureCached(ctx); err != nil {
		return "", false, err
	}

	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	value, ok := a.cache[field.Path]
	return value, ok, nil
}

func (a *AWSSecretsManagerSource) ensureCached(ctx context.Context) error {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	if time.Since(a.lastCacheAt) < a.cacheTTL && a.lastCacheAt != (time.Time{}) {
		return nil
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: awssdk.String(a.secretName),
	}

	result, err := a.client.GetSecretValue(ctx, input)
	if err != nil {
		return fmt.Errorf("cannot fetch secret %s: %w", a.secretName, err)
	}

	a.cache = make(map[string]any)

	if result.SecretString != nil {
		if err := json.Unmarshal([]byte(*result.SecretString), &a.cache); err != nil {
			a.cache["_raw"] = *result.SecretString
		}
	} else if result.SecretBinary != nil {
		a.cache["_raw"] = string(result.SecretBinary)
	}

	a.lastCacheAt = time.Now()
	return nil
}

// FromAWSSecretsManager reads secretName using the ambient region and a five-minute cache.
func FromAWSSecretsManager(secretName string) confkit.Source {
	return FromAWSSecretsManagerWithRegion(secretName, "")
}

// FromAWSSecretsManagerWithRegion reads secretName from a specific region with a five-minute
// cache.
func FromAWSSecretsManagerWithRegion(secretName string, region string) confkit.Source {
	return FromAWSSecretsManagerWithOptions(secretName, region, 5*time.Minute)
}

// FromAWSSecretsManagerWithOptions reads secretName from a specific region, caching it for
// cacheTTL.
//
// A configuration failure is not reported here: the returned Source fails every lookup, so
// the problem appears in the load's ErrorReport.
func FromAWSSecretsManagerWithOptions(secretName string, region string, cacheTTL time.Duration) confkit.Source {
	src, err := NewAWSSecretsManagerSource(secretName, region, cacheTTL)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
