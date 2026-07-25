package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type Resolver interface {
	Resolve(ctx context.Context) (string, error)
}

type ResolverFunc func(context.Context) (string, error)

func (f ResolverFunc) Resolve(ctx context.Context) (string, error) { return f(ctx) }

type SourceConfig struct {
	Direct string
	AWS    *AWSConfig
	GCP    *GCPConfig
	Azure  *AzureConfig
}

type AWSConfig struct {
	Region       string
	SecretID     string
	VersionID    string
	VersionStage string
	JSONKey      string
}

type GCPConfig struct {
	ProjectID string
	SecretID  string
	Version   string
	JSONKey   string
}

type AzureConfig struct {
	VaultURL   string
	SecretName string
	Version    string
	JSONKey    string
}

func NewResolver(cfg SourceConfig) (Resolver, string, error) {
	sources := configuredSources(cfg)
	if len(sources) == 0 {
		return nil, "none", fmt.Errorf("missing OpenAI admin API key source: set admin_api_key or configure exactly one of aws_secrets_manager, gcp_secret_manager, or azure_key_vault")
	}
	if len(sources) > 1 {
		return nil, "multiple", fmt.Errorf("ambiguous OpenAI admin API key source: configure exactly one source, got %s", strings.Join(sources, ", "))
	}

	switch sources[0] {
	case "configuration":
		value := strings.TrimSpace(cfg.Direct)
		return ResolverFunc(func(context.Context) (string, error) { return value, nil }), "configuration", nil
	case "aws_secrets_manager":
		return NewAWSResolver(*cfg.AWS), "aws_secrets_manager", nil
	case "gcp_secret_manager":
		return NewGCPResolver(*cfg.GCP), "gcp_secret_manager", nil
	case "azure_key_vault":
		return NewAzureResolver(*cfg.Azure), "azure_key_vault", nil
	default:
		return nil, "unknown", fmt.Errorf("unsupported OpenAI admin API key source %q", sources[0])
	}
}

func configuredSources(cfg SourceConfig) []string {
	var sources []string
	if strings.TrimSpace(cfg.Direct) != "" {
		sources = append(sources, "configuration")
	}
	if cfg.AWS != nil {
		sources = append(sources, "aws_secrets_manager")
	}
	if cfg.GCP != nil {
		sources = append(sources, "gcp_secret_manager")
	}
	if cfg.Azure != nil {
		sources = append(sources, "azure_key_vault")
	}
	return sources
}

func extractSecretValue(raw, jsonKey string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("secret value is empty")
	}
	if strings.TrimSpace(jsonKey) == "" {
		return value, nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return "", fmt.Errorf("secret value is not valid JSON for json_key %q: %w", jsonKey, err)
	}
	current := decoded
	for _, part := range strings.Split(jsonKey, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", fmt.Errorf("json_key %q contains an empty path segment", jsonKey)
		}
		object, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("json_key %q did not resolve to a string value", jsonKey)
		}
		var exists bool
		current, exists = object[part]
		if !exists {
			return "", fmt.Errorf("json_key %q was not found in secret value", jsonKey)
		}
	}
	resolved, ok := current.(string)
	if !ok || strings.TrimSpace(resolved) == "" {
		return "", fmt.Errorf("json_key %q did not resolve to a non-empty string", jsonKey)
	}
	return strings.TrimSpace(resolved), nil
}

func bytesSecretValue(data []byte, jsonKey string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("secret binary value is empty")
	}
	decoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(decoded, data)
	return extractSecretValue(string(decoded), jsonKey)
}
