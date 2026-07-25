package secrets

import (
	"context"
	"errors"
	"testing"

	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func TestNewResolverExactlyOneSource(t *testing.T) {
	ctx := context.Background()
	resolver, source, err := NewResolver(SourceConfig{Direct: " sk-admin "})
	if err != nil {
		t.Fatalf("NewResolver direct: %v", err)
	}
	if source != "configuration" {
		t.Fatalf("source = %q", source)
	}
	value, err := resolver.Resolve(ctx)
	if err != nil || value != "sk-admin" {
		t.Fatalf("Resolve() = %q, %v", value, err)
	}

	if _, _, err := NewResolver(SourceConfig{}); err == nil {
		t.Fatal("expected missing source error")
	}
	if _, _, err := NewResolver(SourceConfig{Direct: "sk", AWS: &AWSConfig{SecretID: "s"}}); err == nil {
		t.Fatal("expected multiple source error")
	}
}

func TestExtractSecretValue(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		jsonKey string
		want    string
		wantErr bool
	}{
		{name: "raw", raw: " sk-raw ", want: "sk-raw"},
		{name: "json top", raw: `{"admin_api_key":"sk-json"}`, jsonKey: "admin_api_key", want: "sk-json"},
		{name: "json nested", raw: `{"openai":{"admin":"sk-nested"}}`, jsonKey: "openai.admin", want: "sk-nested"},
		{name: "missing", raw: `{}`, jsonKey: "admin_api_key", wantErr: true},
		{name: "not string", raw: `{"admin_api_key":123}`, jsonKey: "admin_api_key", wantErr: true},
		{name: "empty", raw: " ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSecretValue(tt.raw, tt.jsonKey)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("extractSecretValue() = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

type fakeAWSClient struct {
	out *secretsmanager.GetSecretValueOutput
	err error
	in  *secretsmanager.GetSecretValueInput
}

func (f *fakeAWSClient) GetSecretValue(ctx context.Context, in *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	f.in = in
	return f.out, f.err
}

func TestAWSResolverWithMockedClient(t *testing.T) {
	secret := `{"key":"sk-aws"}`
	version := "v1"
	client := &fakeAWSClient{out: &secretsmanager.GetSecretValueOutput{SecretString: &secret}}
	resolver := NewAWSResolverWithClient(AWSConfig{SecretID: "secret/openai", VersionID: version, JSONKey: "key"}, client)
	got, err := resolver.Resolve(context.Background())
	if err != nil || got != "sk-aws" {
		t.Fatalf("Resolve() = %q, %v", got, err)
	}
	if client.in == nil || *client.in.SecretId != "secret/openai" || *client.in.VersionId != version {
		t.Fatalf("unexpected AWS input: %#v", client.in)
	}
}

func TestAWSResolverErrors(t *testing.T) {
	if _, err := NewAWSResolverWithClient(AWSConfig{}, &fakeAWSClient{}).Resolve(context.Background()); err == nil {
		t.Fatal("expected missing secret_id error")
	}
	boom := errors.New("boom")
	if _, err := NewAWSResolverWithClient(AWSConfig{SecretID: "s"}, &fakeAWSClient{err: boom}).Resolve(context.Background()); err == nil {
		t.Fatal("expected SDK error")
	}
}

type fakeGCPClient struct {
	out    *secretmanagerpb.AccessSecretVersionResponse
	err    error
	closed bool
	name   string
}

func (f *fakeGCPClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, _ ...any) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	f.name = req.Name
	return f.out, f.err
}
func (f *fakeGCPClient) Close() error { f.closed = true; return nil }

func TestGCPResolverWithMockedClient(t *testing.T) {
	client := &fakeGCPClient{out: &secretmanagerpb.AccessSecretVersionResponse{Payload: &secretmanagerpb.SecretPayload{Data: []byte(`{"key":"sk-gcp"}`)}}}
	resolver := NewGCPResolverWithClient(GCPConfig{ProjectID: "p", SecretID: "openai", JSONKey: "key"}, client)
	got, err := resolver.Resolve(context.Background())
	if err != nil || got != "sk-gcp" {
		t.Fatalf("Resolve() = %q, %v", got, err)
	}
	if client.name != "projects/p/secrets/openai/versions/latest" || !client.closed {
		t.Fatalf("unexpected GCP request: name=%q closed=%v", client.name, client.closed)
	}
}

func TestGCPResolverErrors(t *testing.T) {
	if _, err := NewGCPResolverWithClient(GCPConfig{SecretID: "s"}, &fakeGCPClient{}).Resolve(context.Background()); err == nil {
		t.Fatal("expected missing project_id")
	}
	if _, err := NewGCPResolverWithClient(GCPConfig{ProjectID: "p"}, &fakeGCPClient{}).Resolve(context.Background()); err == nil {
		t.Fatal("expected missing secret_id")
	}
}

type fakeAzureClient struct {
	out     azsecrets.GetSecretResponse
	err     error
	name    string
	version string
}

func (f *fakeAzureClient) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	f.name = name
	f.version = version
	return f.out, f.err
}

func TestAzureResolverWithMockedClient(t *testing.T) {
	value := `{"key":"sk-azure"}`
	client := &fakeAzureClient{out: azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &value}}}
	resolver := NewAzureResolverWithClient(AzureConfig{VaultURL: "https://vault.example", SecretName: "openai", Version: "42", JSONKey: "key"}, client)
	got, err := resolver.Resolve(context.Background())
	if err != nil || got != "sk-azure" {
		t.Fatalf("Resolve() = %q, %v", got, err)
	}
	if client.name != "openai" || client.version != "42" {
		t.Fatalf("unexpected Azure request: %q %q", client.name, client.version)
	}
}

func TestAzureResolverErrors(t *testing.T) {
	if _, err := NewAzureResolverWithClient(AzureConfig{SecretName: "s"}, &fakeAzureClient{}).Resolve(context.Background()); err == nil {
		t.Fatal("expected missing vault_url")
	}
	if _, err := NewAzureResolverWithClient(AzureConfig{VaultURL: "https://vault.example"}, &fakeAzureClient{}).Resolve(context.Background()); err == nil {
		t.Fatal("expected missing secret_name")
	}
}

func TestResolverConstructorsAndBinarySecret(t *testing.T) {
	if NewAWSResolver(AWSConfig{SecretID: "s"}) == nil {
		t.Fatal("NewAWSResolver returned nil")
	}
	if NewGCPResolver(GCPConfig{ProjectID: "p", SecretID: "s"}) == nil {
		t.Fatal("NewGCPResolver returned nil")
	}
	if NewAzureResolver(AzureConfig{VaultURL: "https://vault.example", SecretName: "s"}) == nil {
		t.Fatal("NewAzureResolver returned nil")
	}
	got, err := bytesSecretValue([]byte("secret"), "")
	if err != nil || got == "" {
		t.Fatalf("bytesSecretValue() = %q, %v", got, err)
	}
	if _, err := bytesSecretValue(nil, ""); err == nil {
		t.Fatal("expected empty binary error")
	}
}

func TestNewResolverCloudSourceNames(t *testing.T) {
	cases := []struct {
		name string
		cfg  SourceConfig
		want string
	}{
		{name: "aws", cfg: SourceConfig{AWS: &AWSConfig{SecretID: "s"}}, want: "aws_secrets_manager"},
		{name: "gcp", cfg: SourceConfig{GCP: &GCPConfig{ProjectID: "p", SecretID: "s"}}, want: "gcp_secret_manager"},
		{name: "azure", cfg: SourceConfig{Azure: &AzureConfig{VaultURL: "https://vault.example", SecretName: "s"}}, want: "azure_key_vault"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resolver, source, err := NewResolver(tt.cfg)
			if err != nil {
				t.Fatalf("NewResolver() error = %v", err)
			}
			if resolver == nil || source != tt.want {
				t.Fatalf("resolver/source = %#v/%q", resolver, source)
			}
		})
	}
}
