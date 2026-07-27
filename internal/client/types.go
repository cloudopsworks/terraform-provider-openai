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
	OwnerRole     string
	CreatedAt     int64
	ExpiresAt     int64
	LastUsedAt    int64
}

type AdminAPIKeyCreateRequest struct {
	Name             string
	ExpiresInSeconds int64
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

type OrganizationUserProject struct {
	ID   string
	Name string
	Role string
}

type OrganizationUser struct {
	ID                             string
	Name                           string
	Email                          string
	Role                           string
	UserID                         string
	UserName                       string
	UserEmail                      string
	Picture                        string
	DeveloperPersona               string
	TechnicalLevel                 string
	AddedAt                        int64
	Created                        int64
	APIKeyLastUsedAt               int64
	BannedAt                       int64
	IsDefault                      bool
	IsScaleTierAuthorizedPurchaser bool
	IsScimManaged                  bool
	IsServiceAccount               bool
	Banned                         bool
	Enabled                        bool
	Projects                       []OrganizationUserProject
}

type OrganizationUserListRequest struct {
	After  string
	Limit  int64
	Emails []string
}

type OrganizationUserUpdateRequest struct {
	Role             string
	RoleID           string
	DeveloperPersona string
	TechnicalLevel   string
}

type OrganizationUserListResponse struct {
	Items   []OrganizationUser
	HasMore bool
	LastID  string
}

type OrganizationGroup struct {
	ID            string
	Name          string
	GroupType     string
	IsScimManaged bool
	CreatedAt     int64
}

type OrganizationGroupCreateRequest struct {
	Name string
}

type OrganizationGroupUpdateRequest struct {
	Name string
}

type OrganizationGroupListRequest struct {
	After string
	Limit int64
	Order string
}

type OrganizationGroupListResponse struct {
	Items   []OrganizationGroup
	HasMore bool
	Next    string
}

type OrganizationGroupUser struct {
	ID               string
	Email            string
	Name             string
	IsServiceAccount bool
	Picture          string
	UserType         string
}

type OrganizationGroupUserCreateRequest struct {
	UserID string
}

type OrganizationGroupUserListRequest struct {
	After string
	Limit int64
	Order string
}

type OrganizationGroupUserListResponse struct {
	Items   []OrganizationGroupUser
	HasMore bool
	Next    string
}

type Role struct {
	ID             string
	Name           string
	Description    string
	Permissions    []string
	PredefinedRole bool
	ResourceType   string
}

type RoleCreateRequest struct {
	Name        string
	Description string
	Permissions []string
}

type RoleUpdateRequest struct {
	Name        string
	Description string
	Permissions []string
}

type RoleListRequest struct {
	After string
	Limit int64
	Order string
}

type RoleListResponse struct {
	Items   []Role
	HasMore bool
	Next    string
}

type RoleAssignmentSource struct {
	PrincipalID   string
	PrincipalType string
}

type RoleAssignment struct {
	Role
	PrincipalID       string
	PrincipalType     string
	CreatedAt         int64
	UpdatedAt         int64
	CreatedBy         string
	AssignmentSources []RoleAssignmentSource
}

type RoleAssignmentCreateRequest struct {
	RoleID string
}

type RoleAssignmentListRequest struct {
	After string
	Limit int64
	Order string
}

type RoleAssignmentListResponse struct {
	Items   []RoleAssignment
	HasMore bool
	Next    string
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

	CreateAdminAPIKey(ctx context.Context, req AdminAPIKeyCreateRequest) (*AdminAPIKey, error)
	GetAdminAPIKey(ctx context.Context, id string) (*AdminAPIKey, error)
	ListAdminAPIKeys(ctx context.Context, req AdminAPIKeyListRequest) (*AdminAPIKeyListResponse, error)
	DeleteAdminAPIKey(ctx context.Context, id string) error

	GetOrganizationUser(ctx context.Context, userID string) (*OrganizationUser, error)
	ListOrganizationUsers(ctx context.Context, req OrganizationUserListRequest) (*OrganizationUserListResponse, error)
	UpdateOrganizationUser(ctx context.Context, userID string, req OrganizationUserUpdateRequest) (*OrganizationUser, error)
	DeleteOrganizationUser(ctx context.Context, userID string) error

	CreateOrganizationGroup(ctx context.Context, req OrganizationGroupCreateRequest) (*OrganizationGroup, error)
	GetOrganizationGroup(ctx context.Context, groupID string) (*OrganizationGroup, error)
	ListOrganizationGroups(ctx context.Context, req OrganizationGroupListRequest) (*OrganizationGroupListResponse, error)
	UpdateOrganizationGroup(ctx context.Context, groupID string, req OrganizationGroupUpdateRequest) (*OrganizationGroup, error)
	DeleteOrganizationGroup(ctx context.Context, groupID string) error

	CreateOrganizationGroupUser(ctx context.Context, groupID string, req OrganizationGroupUserCreateRequest) (*OrganizationGroupUser, error)
	GetOrganizationGroupUser(ctx context.Context, groupID, userID string) (*OrganizationGroupUser, error)
	ListOrganizationGroupUsers(ctx context.Context, groupID string, req OrganizationGroupUserListRequest) (*OrganizationGroupUserListResponse, error)
	DeleteOrganizationGroupUser(ctx context.Context, groupID, userID string) error

	CreateOrganizationRole(ctx context.Context, req RoleCreateRequest) (*Role, error)
	GetOrganizationRole(ctx context.Context, roleID string) (*Role, error)
	ListOrganizationRoles(ctx context.Context, req RoleListRequest) (*RoleListResponse, error)
	UpdateOrganizationRole(ctx context.Context, roleID string, req RoleUpdateRequest) (*Role, error)
	DeleteOrganizationRole(ctx context.Context, roleID string) error

	CreateProjectRole(ctx context.Context, projectID string, req RoleCreateRequest) (*Role, error)
	GetProjectRole(ctx context.Context, projectID, roleID string) (*Role, error)
	ListProjectRoles(ctx context.Context, projectID string, req RoleListRequest) (*RoleListResponse, error)
	UpdateProjectRole(ctx context.Context, projectID, roleID string, req RoleUpdateRequest) (*Role, error)
	DeleteProjectRole(ctx context.Context, projectID, roleID string) error

	CreateOrganizationUserRole(ctx context.Context, userID string, req RoleAssignmentCreateRequest) (*RoleAssignment, error)
	GetOrganizationUserRole(ctx context.Context, userID, roleID string) (*RoleAssignment, error)
	ListOrganizationUserRoles(ctx context.Context, userID string, req RoleAssignmentListRequest) (*RoleAssignmentListResponse, error)
	DeleteOrganizationUserRole(ctx context.Context, userID, roleID string) error

	CreateOrganizationGroupRole(ctx context.Context, groupID string, req RoleAssignmentCreateRequest) (*RoleAssignment, error)
	GetOrganizationGroupRole(ctx context.Context, groupID, roleID string) (*RoleAssignment, error)
	ListOrganizationGroupRoles(ctx context.Context, groupID string, req RoleAssignmentListRequest) (*RoleAssignmentListResponse, error)
	DeleteOrganizationGroupRole(ctx context.Context, groupID, roleID string) error
}

func requireNonEmpty(entity, field, value string) error {
	if value == "" {
		return fmt.Errorf("openai %s response missing required field %s", entity, field)
	}
	return nil
}
