package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MimoJanra/confkit"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// AWSSSMSource reads configuration values from AWS SSM Parameter Store.
//
// All parameters below the path prefix are fetched recursively in one pass and cached for
// the configured TTL, so a load performs a handful of API calls rather than one per
// field. SecureString parameters are decrypted. Field paths become parameter paths:
// DB.Host under prefix "/myapp/" reads "/myapp/db/host".
type AWSSSMSource struct {
	pathPrefix  string
	client      *ssm.Client
	cache       map[string]string
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	lastCacheAt time.Time
}

// NewAWSSSMSource returns a source using the region from the ambient AWS configuration.
func NewAWSSSMSource(pathPrefix string, cacheTTL time.Duration) (*AWSSSMSource, error) {
	return NewAWSSSMSourceWithRegion(pathPrefix, cacheTTL, "")
}

// NewAWSSSMSourceWithRegion returns a source pinned to region. An empty region falls back
// to the ambient AWS configuration.
//
// Credentials are resolved by the AWS SDK's default chain and are not verified here, so an
// authentication problem surfaces on the first lookup.
func NewAWSSSMSourceWithRegion(pathPrefix string, cacheTTL time.Duration, region string) (*AWSSSMSource, error) {
	opts := []func(*config.LoadOptions) error{}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot load AWS config: %w", err)
	}

	return &AWSSSMSource{
		pathPrefix: pathPrefix,
		client:     ssm.NewFromConfig(cfg),
		cache:      make(map[string]string),
		cacheTTL:   cacheTTL,
	}, nil
}

// Name returns "aws-ssm".
func (a *AWSSSMSource) Name() string {
	return "aws-ssm"
}

// Lookup refreshes the parameter cache if it has expired, then returns the value for
// field. A parameter that is absent means not found rather than an error.
func (a *AWSSSMSource) Lookup(ctx context.Context, field *confkit.FieldInfo) (any, bool, error) {
	if err := a.ensureCached(ctx); err != nil {
		return "", false, err
	}

	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	paramPath := a.fieldPathToParameterPath(field.Path)
	value, ok := a.cache[paramPath]
	return value, ok, nil
}

func (a *AWSSSMSource) ensureCached(ctx context.Context) error {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	if time.Since(a.lastCacheAt) < a.cacheTTL && a.lastCacheAt != (time.Time{}) {
		return nil
	}

	paginator := ssm.NewGetParametersByPathPaginator(a.client, &ssm.GetParametersByPathInput{
		Path:           awssdk.String(a.pathPrefix),
		Recursive:      awssdk.Bool(true),
		WithDecryption: awssdk.Bool(true),
	})

	a.cache = make(map[string]string)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("cannot fetch SSM parameters: %w", err)
		}

		for _, param := range page.Parameters {
			if param.Name != nil && param.Value != nil {
				a.cache[strings.ToLower(*param.Name)] = *param.Value
			}
		}
	}

	a.lastCacheAt = time.Now()
	return nil
}

func (a *AWSSSMSource) fieldPathToParameterPath(fieldPath string) string {
	parts := strings.Split(fieldPath, ".")
	paramParts := make([]string, len(parts))
	for i, part := range parts {
		paramParts[i] = strings.ToLower(part)
	}
	return strings.ToLower(a.pathPrefix) + strings.Join(paramParts, "/")
}

func (a *AWSSSMSource) parameterPathToFieldPath(paramPath string) string {
	prefix := strings.ToLower(a.pathPrefix)
	if !strings.HasPrefix(paramPath, prefix) {
		return ""
	}

	relativePath := strings.TrimPrefix(paramPath, prefix)
	relativePath = strings.TrimPrefix(relativePath, "/")

	parts := strings.Split(relativePath, "/")
	return strings.Join(parts, ".")
}

// FromAWSSSMParameterStore reads parameters under pathPrefix with a five-minute cache.
func FromAWSSSMParameterStore(pathPrefix string) confkit.Source {
	return FromAWSSSMParameterStoreWithTTL(pathPrefix, 5*time.Minute)
}

// FromAWSSSMParameterStoreWithTTL reads parameters under pathPrefix, caching them for
// cacheTTL and appending a trailing slash to the prefix if absent.
//
// A configuration failure is not reported here: the returned Source fails every lookup, so
// the problem appears in the load's ErrorReport.
func FromAWSSSMParameterStoreWithTTL(pathPrefix string, cacheTTL time.Duration) confkit.Source {
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}

	src, err := NewAWSSSMSource(pathPrefix, cacheTTL)
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
