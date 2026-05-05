package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"confkit"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSSecretsManagerSource struct {
	secretName  string
	region      string
	client      *secretsmanager.Client
	cache       map[string]any
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	lastCacheAt time.Time
	ctx         context.Context
}

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
		ctx:        ctx,
	}, nil
}

func (a *AWSSecretsManagerSource) Name() string {
	return "aws-secrets-manager"
}

func (a *AWSSecretsManagerSource) Lookup(field *confkit.FieldInfo) (any, bool, error) {
	if err := a.ensureCached(); err != nil {
		return "", false, err
	}

	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	value, ok := a.cache[field.Path]
	return value, ok, nil
}

func (a *AWSSecretsManagerSource) ensureCached() error {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	if time.Since(a.lastCacheAt) < a.cacheTTL && a.lastCacheAt != (time.Time{}) {
		return nil
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: awssdk.String(a.secretName),
	}

	result, err := a.client.GetSecretValue(a.ctx, input)
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

func FromAWSSecretsManager(secretName string) confkit.Source {
	return FromAWSSecretsManagerWithRegion(secretName, "")
}

func FromAWSSecretsManagerWithRegion(secretName string, region string) confkit.Source {
	return FromAWSSecretsManagerWithOptions(secretName, region, 5*time.Minute)
}

func FromAWSSecretsManagerWithOptions(secretName string, region string, cacheTTL time.Duration) confkit.Source {
	src, err := NewAWSSecretsManagerSource(secretName, region, cacheTTL)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
