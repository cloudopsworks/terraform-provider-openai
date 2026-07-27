package client

import (
	"context"
	"fmt"
)

type Project struct {
	ID            string
	Name          string
	Status        string
	ExternalKeyID string
	CreatedAt     int64
	ArchivedAt    int64
}

type ProjectCreateRequest struct {
	Name          string
	ExternalKeyID string
	Geography     string
}

type ProjectUpdateRequest struct {
	Name          string
	ExternalKeyID string
	Geography     string
}

type ProjectListRequest struct {
	After           string
	Limit           int64
	IncludeArchived bool
}

type ProjectListResponse struct {
	Items   []Project
	HasMore bool
	LastID  string
}

type ServiceAccount struct {
	ID        string
	Name      string
	Role      string
	CreatedAt int64
}

type ServiceAccountCreateRequest struct {
	Name string
	Role string
}

type ServiceAccountUpdateRequest struct {
	Name string
	Role string
}

type ServiceAccountListRequest struct {
	After string
	Limit int64
}

type ServiceAccountListResponse struct {
	Items   []ServiceAccount
	HasMore bool
	LastID  string
}

type APIKey struct {
	ID                 string
	Name               string
	RedactedValue      string
	OwnerType          string
	OwnerID            string
	OwnerName          string
	OwnerProjectAccess string
	CreatedAt          int64
	LastUsedAt         int64
}

type AdminAPIKey struct {
	ID            string
	Name          string
	Value         string
	RedactedValue string
	OwnerType     string
	OwnerID       string
	OwnerName     string
	CreatedAt     int64
	ExpiresAt     int64
	LastUsedAt    int64
}

type ServiceAccountAPIKeyCreateRequest struct {
	Name   string
	Scopes []string
}

type ServiceAccountAPIKeyCreateResponse struct {
	ID        string
	Name      string
	Value     string
	CreatedAt int64
}

type AdminAPIKeyCreateRequest struct {
	Name string
}

type AdminAPIKeyListRequest struct {
	After string
	Limit int64
	Order string
}

type AdminAPIKeyListResponse struct {
	Items   []AdminAPIKey
	HasMore bool
	LastID  string
}

type AdminClient interface {
	CreateProject(ctx context.Context, req ProjectCreateRequest) (*Project, error)
	GetProject(ctx context.Context, id string) (*Project, error)
	ListProjects(ctx context.Context, req ProjectListRequest) (*ProjectListResponse, error)
	UpdateProject(ctx context.Context, id string, req ProjectUpdateRequest) (*Project, error)
	ArchiveProject(ctx context.Context, id string) (*Project, error)
	CreateServiceAccount(ctx context.Context, projectID string, req ServiceAccountCreateRequest) (*ServiceAccount, error)
	GetServiceAccount(ctx context.Context, projectID, serviceAccountID string) (*ServiceAccount, error)
	ListServiceAccounts(ctx context.Context, projectID string, req ServiceAccountListRequest) (*ServiceAccountListResponse, error)
	UpdateServiceAccount(ctx context.Context, projectID, serviceAccountID string, req ServiceAccountUpdateRequest) (*ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, projectID, serviceAccountID string) error
	CreateServiceAccountAPIKey(ctx context.Context, projectID, serviceAccountID string, req ServiceAccountAPIKeyCreateRequest) (*ServiceAccountAPIKeyCreateResponse, error)
	GetProjectAPIKey(ctx context.Context, projectID, apiKeyID string) (*APIKey, error)
	DeleteProjectAPIKey(ctx context.Context, projectID, apiKeyID string) error
}

func requireNonEmpty(entity, field, value string) error {
	if value == "" {
		return fmt.Errorf("openai %s response missing required field %s", entity, field)
	}
	return nil
}
