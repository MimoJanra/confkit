// Package aws provides confkit Sources for AWS Systems Manager and Secrets Manager.
//
// Import and use with Load:
//
//	import "github.com/MimoJanra/confkit/aws"
//
//	// AWS Systems Manager Parameter Store
//	cfg, err := confkit.Load[Config](
//	    aws.FromAWSSSMParameterStore("/prod/myapp"),
//	)
//
//	// AWS Secrets Manager
//	cfg, err := confkit.Load[Config](
//	    aws.FromAWSSecretsManager("prod/app-secrets"),
//	)
//
// # AWS Systems Manager Parameter Store
//
// Use for configuration parameters:
//
//	// Load from parameter hierarchy
//	src := aws.FromAWSSSMParameterStore("/prod/myapp")
//
//	// With custom TTL (default 1 hour)
//	src := aws.FromAWSSSMParameterStoreWithTTL("/prod/myapp", 30*time.Minute)
//
// Config fields are mapped to parameter paths using dot notation:
//
//	type Config struct {
//	    Host string
//	    Port int
//	}
//
// With prefix "/prod/myapp", parameters are:
// /prod/myapp/host → Config.Host
// /prod/myapp/port → Config.Port
//
// # AWS Secrets Manager
//
// Use for sensitive data (passwords, API keys, etc.):
//
//	// Basic usage
//	src := aws.FromAWSSecretsManager("prod/app-secrets")
//
//	// Specify region
//	src := aws.FromAWSSecretsManagerWithRegion("prod/app-secrets", "us-west-2")
//
//	// Multi-region failover
//	src := aws.FromAWSSecretsManagerMultiRegion(
//	    "prod/app-secrets",
//	    []string{"us-east-1", "us-west-2"},
//	)
//
// Secrets Manager expects JSON secrets:
//
//	{
//	    "api_key": "sk_live_xxx",
//	    "password": "secret_password"
//	}
//
// # Secrets
//
// Always mark sensitive fields with secret:"true":
//
//	type Config struct {
//	    APIKey   string `secret:"true"`
//	    Password string `secret:"true"`
//	}
//
// Secrets are automatically redacted in error messages and logs.
//
// # Multi-Region Failover
//
// For high availability, use multi-region sources:
//
//	// Parameter Store with failover
//	src := aws.FromAWSSSMParameterStoreMultiRegion(
//	    "/prod/myapp",
//	    []string{"us-east-1", "us-west-2"},
//	)
//
//	// Secrets Manager with failover
//	src := aws.FromAWSSecretsManagerMultiRegion(
//	    "prod/app-secrets",
//	    []string{"us-east-1", "us-west-2"},
//	)
//
// The source will try regions in order and use the first available.
//
// # Caching
//
// Secrets are cached to reduce API calls. Cache TTL can be customized:
//
//	src := aws.FromAWSSecretsManagerWithOptions(
//	    "prod/app-secrets",
//	    "us-east-1",
//	    5 * time.Minute, // Custom TTL
//	)
//
// # IAM Permissions
//
// Required IAM permissions:
//
// For Systems Manager Parameter Store:
// • ssm:GetParameter
// • ssm:GetParameters
//
// For Secrets Manager:
// • secretsmanager:GetSecretValue
//
// Example IAM policy:
//
//	{
//	    "Version": "2012-10-17",
//	    "Statement": [
//	        {
//	            "Effect": "Allow",
//	            "Action": [
//	                "ssm:GetParameter",
//	                "secretsmanager:GetSecretValue"
//	            ],
//	            "Resource": [
//	                "arn:aws:ssm:*:*:parameter/prod/myapp/*",
//	                "arn:aws:secretsmanager:*:*:secret:prod/app-secrets*"
//	            ]
//	        }
//	    ]
//	}
package aws
