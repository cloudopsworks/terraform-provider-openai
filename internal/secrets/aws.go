package secrets

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type awsSecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type AWSResolver struct {
	cfg       AWSConfig
	newClient func(context.Context, AWSConfig) (awsSecretsManagerClient, error)
}

func NewAWSResolver(cfg AWSConfig) *AWSResolver {
	return &AWSResolver{cfg: cfg, newClient: newAWSSecretsManagerClient}
}

func NewAWSResolverWithClient(cfg AWSConfig, cl awsSecretsManagerClient) *AWSResolver {
	return &AWSResolver{cfg: cfg, newClient: func(context.Context, AWSConfig) (awsSecretsManagerClient, error) { return cl, nil }}
}

func (r *AWSResolver) Resolve(ctx context.Context) (string, error) {
	if r.cfg.SecretID == "" {
		return "", fmt.Errorf("aws_secrets_manager.secret_id is required")
	}
	cl, err := r.newClient(ctx, r.cfg)
	if err != nil {
		return "", fmt.Errorf("create AWS Secrets Manager client: %w", err)
	}
	input := &secretsmanager.GetSecretValueInput{SecretId: &r.cfg.SecretID}
	if r.cfg.VersionID != "" {
		input.VersionId = &r.cfg.VersionID
	}
	if r.cfg.VersionStage != "" {
		input.VersionStage = &r.cfg.VersionStage
	}
	out, err := cl.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("read AWS Secrets Manager secret: %w", err)
	}
	if out.SecretString != nil {
		return extractSecretValue(*out.SecretString, r.cfg.JSONKey)
	}
	return bytesSecretValue(out.SecretBinary, r.cfg.JSONKey)
}

func newAWSSecretsManagerClient(ctx context.Context, cfg AWSConfig) (awsSecretsManagerClient, error) {
	options := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		options = append(options, awsconfig.WithRegion(cfg.Region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, err
	}
	return secretsmanager.NewFromConfig(awsCfg), nil
}
