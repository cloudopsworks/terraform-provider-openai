package secrets

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type gcpSecretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...any) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

type gcpClientAdapter struct{ *secretmanager.Client }

func (a gcpClientAdapter) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, _ ...any) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return a.Client.AccessSecretVersion(ctx, req)
}

type GCPResolver struct {
	cfg       GCPConfig
	newClient func(context.Context, GCPConfig) (gcpSecretManagerClient, error)
}

func NewGCPResolver(cfg GCPConfig) *GCPResolver {
	return &GCPResolver{cfg: cfg, newClient: newGCPSecretManagerClient}
}

func NewGCPResolverWithClient(cfg GCPConfig, cl gcpSecretManagerClient) *GCPResolver {
	return &GCPResolver{cfg: cfg, newClient: func(context.Context, GCPConfig) (gcpSecretManagerClient, error) { return cl, nil }}
}

func (r *GCPResolver) Resolve(ctx context.Context) (string, error) {
	if r.cfg.ProjectID == "" {
		return "", fmt.Errorf("gcp_secret_manager.project_id is required")
	}
	if r.cfg.SecretID == "" {
		return "", fmt.Errorf("gcp_secret_manager.secret_id is required")
	}
	version := r.cfg.Version
	if version == "" {
		version = "latest"
	}
	cl, err := r.newClient(ctx, r.cfg)
	if err != nil {
		return "", fmt.Errorf("create GCP Secret Manager client: %w", err)
	}
	defer cl.Close()
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", r.cfg.ProjectID, r.cfg.SecretID, version)
	out, err := cl.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{Name: name})
	if err != nil {
		return "", fmt.Errorf("read GCP Secret Manager secret: %w", err)
	}
	if out == nil || out.Payload == nil {
		return "", fmt.Errorf("GCP Secret Manager secret payload is empty")
	}
	return extractSecretValue(string(out.Payload.Data), r.cfg.JSONKey)
}

func newGCPSecretManagerClient(ctx context.Context, _ GCPConfig) (gcpSecretManagerClient, error) {
	cl, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return gcpClientAdapter{Client: cl}, nil
}
