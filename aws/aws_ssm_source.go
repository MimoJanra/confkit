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

type AWSSSMSource struct {
	pathPrefix  string
	client      *ssm.Client
	cache       map[string]string
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	lastCacheAt time.Time
	ctx         context.Context
}

func NewAWSSSMSource(pathPrefix string, cacheTTL time.Duration) (*AWSSSMSource, error) {
	return NewAWSSSMSourceWithRegion(pathPrefix, cacheTTL, "")
}

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
		ctx:        context.Background(),
	}, nil
}

func (a *AWSSSMSource) Name() string {
	return "aws-ssm"
}

func (a *AWSSSMSource) Lookup(field *confkit.FieldInfo) (any, bool, error) {
	if err := a.ensureCached(); err != nil {
		return "", false, err
	}

	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	paramPath := a.fieldPathToParameterPath(field.Path)
	value, ok := a.cache[paramPath]
	return value, ok, nil
}

func (a *AWSSSMSource) ensureCached() error {
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
		page, err := paginator.NextPage(a.ctx)
		if err != nil {
			return fmt.Errorf("cannot fetch SSM parameters: %w", err)
		}

		for _, param := range page.Parameters {
			if param.Name != nil && param.Value != nil {
				a.cache[*param.Name] = *param.Value
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
	return a.pathPrefix + strings.Join(paramParts, "/")
}

func (a *AWSSSMSource) parameterPathToFieldPath(paramPath string) string {
	if !strings.HasPrefix(paramPath, a.pathPrefix) {
		return ""
	}

	relativePath := strings.TrimPrefix(paramPath, a.pathPrefix)
	relativePath = strings.TrimPrefix(relativePath, "/")

	parts := strings.Split(relativePath, "/")
	return strings.Join(parts, ".")
}

func FromAWSSSMParameterStore(pathPrefix string) confkit.Source {
	return FromAWSSSMParameterStoreWithTTL(pathPrefix, 5*time.Minute)
}

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
