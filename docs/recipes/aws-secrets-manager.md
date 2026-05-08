---
layout: default
title: "Recipe: AWS Secrets Manager"
---

# Recipe: AWS Secrets Manager

Load secrets from AWS Secrets Manager.

## Installation

```bash
go get github.com/MimoJanra/confkit/aws@latest
```

## Code

```go
package main

import (
    "context"
    "log"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/aws"
)

type Config struct {
    Database struct {
        Host     string `validate:"required"`
        Port     int    `default:"5432"`
        Name     string `validate:"required"`
        User     string `validate:"required"`
        Password string `validate:"required" secret:"true"`
    } `prefix:"DB_"`
}

func main() {
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
        aws.FromAWSSecretsManager("prod/database"),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("Database: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
```

## AWS Secrets Manager Setup

### Create Secret

```bash
aws secretsmanager create-secret \
  --name prod/database \
  --description "Production database credentials" \
  --secret-string '{
    "DB_HOST": "postgres.internal",
    "DB_PORT": "5432",
    "DB_NAME": "myapp_db",
    "DB_USER": "appuser",
    "DB_PASSWORD": "secret-password"
  }'
```

## Basic Usage

### With Default Region

```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManager("prod/database"),
)
```

Uses `AWS_REGION` environment variable or default region.

### With Specific Region

```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManagerWithRegion("prod/database", "us-east-1"),
)
```

## Multi-Region

For high availability:

```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManagerMultiRegion("prod/database", []string{
        "us-east-1",
        "us-west-2",
    }),
)
```

Tries the first region, falls back to others on failure.

## Caching and TTL

Reduce API calls with caching:

```go
// Cache for 5 minutes
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManagerWithOptions("prod/database", "us-east-1", 5*time.Minute),
)
```

## IAM Policy

Grant your application permission to read secrets:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/database-*"
      ]
    }
  ]
}
```

## EC2 Instance

```bash
# Set IAM role for EC2 instance
aws ec2 create-iam-instance-profile --iam-instance-profile-name myapp-role

# Then in app code
cfg, err := confkit.Load[Config](
    aws.FromAWSSecretsManager("prod/database"),
    // Uses IAM role from EC2 metadata
)
```

## ECS Task

```json
{
  "name": "myapp",
  "image": "myapp:latest",
  "taskRoleArn": "arn:aws:iam::123456789012:role/myapp-role",
  "environment": [
    {
      "name": "AWS_REGION",
      "value": "us-east-1"
    }
  ]
}
```

## Lambda Function

```go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/aws"
)

type Config struct {
    Database struct {
        Host     string
        User     string
        Password string `secret:"true"`
    } `prefix:"DB_"`
}

var config Config

func init() {
    var err error
    config, err = confkit.Load[Config](
        aws.FromAWSSecretsManager("prod/database"),
    )
    if err != nil {
        panic(err)
    }
}

func handler(ctx context.Context, event interface{}) error {
    // Use config
    return nil
}

func main() {
    lambda.Start(handler)
}
```

## Combining with Other Sources

The first source to provide a value wins; later sources fill in only unset fields:

```go
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),                           // Highest priority — checked first
    aws.FromAWSSecretsManager("prod/database"),  // Secrets fill in what env did not set
    confkit.FromYAML("config.yaml"),              // Base fallback
)
```

## Rotating Secrets

Update secret in AWS:

```bash
aws secretsmanager update-secret \
  --secret-id prod/database \
  --secret-string '{
    "DB_HOST": "new-host.internal",
    "DB_PASSWORD": "new-password"
  }'
```

For automatic rotation, enable rotation in Secrets Manager console.

## Real-World Example

```go
package main

import (
    "log"
    "time"
    "github.com/MimoJanra/confkit"
    "github.com/MimoJanra/confkit/aws"
)

type Config struct {
    App struct {
        Name    string `default:"myapp"`
        Version string `validate:"required"`
    }
    Database struct {
        Host     string `validate:"required"`
        Port     int    `default:"5432"`
        Name     string `validate:"required"`
        User     string `validate:"required"`
        Password string `secret:"true" validate:"required"`
    } `prefix:"DB_"`
}

func main() {
    // Load from Secrets Manager with 10-minute cache
    cfg, err := confkit.Load[Config](
        confkit.FromEnv(),
        aws.FromAWSSecretsManagerWithOptions("prod/database", "us-east-1", 10*time.Minute),
    )
    if err != nil {
        log.Fatal(confkit.Explain(err))
    }
    
    log.Printf("App: %s v%s", cfg.App.Name, cfg.App.Version)
    log.Printf("Connected to %s@%s", cfg.Database.User, cfg.Database.Host)
}
```

## AWS SSM Parameter Store

Alternative to Secrets Manager:

```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSSMParameterStore("/prod/myapp"),
)
```

Or with multi-region:

```go
cfg, err := confkit.Load[Config](
    aws.FromAWSSSMParameterStoreMultiRegion("/prod/myapp", []string{
        "us-east-1",
        "us-west-2",
    }),
)
```

## Monitoring

CloudWatch metrics:

```bash
# List API calls to secrets
aws cloudtrail lookup-events --lookup-attributes AttributeKey=EventName,AttributeValue=GetSecretValue
```

## Best Practices

1. **Use specific secret names**
   ```go
   aws.FromAWSSecretsManager("prod/database")
   ```

2. **Cache secrets with TTL**
   ```go
   aws.FromAWSSecretsManagerWithOptions(name, region, 5*time.Minute)
   ```

3. **Enable encryption**
   ```bash
   aws secretsmanager create-secret --kms-key-id alias/aws/secretsmanager ...
   ```

4. **Use IAM roles**
   - EC2: Attach role to instance
   - ECS: Use taskRoleArn
   - Lambda: Use execution role

5. **Mark secrets in code**
   ```go
   Password string `secret:"true"`
   ```

## Troubleshooting

### Access denied

```
AccessDeniedException: User: arn:aws:iam::123456789012:user/app is not authorized
```

Check IAM policy allows `secretsmanager:GetSecretValue`.

### Secret not found

```
ResourceNotFoundException: Secrets Manager can't find the specified secret
```

Verify secret exists:

```bash
aws secretsmanager describe-secret --secret-id prod/database
```

### Wrong region

```
RequestError: send request failed
```

Ensure region is correct:

```bash
export AWS_REGION=us-east-1
go run main.go
```

## See Also

- **[Vault](./vault.md)** — HashiCorp Vault alternative
- **[Sources](../docs/sources.md)** — All configuration sources
- **[Secret Redaction](../docs/secret-redaction.md)** — Protecting secrets
