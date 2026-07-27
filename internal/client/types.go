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
	APIKey    *ServiceAccountAPIKeyCreateResponse
}

type ServiceAccountCreateRequest struct {
	Name                     string
	Role                     string
	CreateServiceAccountOnly bool
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

type InviteProject struct {
	ID   string
	Role string
}

type Invite struct {
	ID         string
	Email      string
	Role       string
	Status     string
	Projects   []InviteProject
	CreatedAt  int64
	AcceptedAt int64
	ExpiresAt  int64
}

type InviteCreateRequest struct {
	Email    string
	Role     string
	Projects []InviteProject
}

type InviteListRequest struct {
	After string
	Limit int64
}

type InviteListResponse struct {
	Items   []Invite
	HasMore bool
	LastID  string
}

type DataRetention struct {
	ID   string
	Type string
}

type DataRetentionUpdateRequest struct {
	Type string
}

type SpendLimit struct {
	ID                string
	Currency          string
	Interval          string
	ThresholdAmount   int64
	EnforcementStatus string
}

type SpendLimitUpdateRequest struct {
	Currency        string
	Interval        string
	ThresholdAmount int64
}

type SpendAlertNotificationChannel struct {
	Recipients    []string
	Type          string
	SubjectPrefix string
}

type SpendAlert struct {
	ID                  string
	Currency            string
	Interval            string
	ThresholdAmount     int64
	NotificationChannel SpendAlertNotificationChannel
}

type SpendAlertCreateRequest struct {
	Currency            string
	Interval            string
	ThresholdAmount     int64
	NotificationChannel SpendAlertNotificationChannel
}

type SpendAlertUpdateRequest struct {
	Currency            string
	Interval            string
	ThresholdAmount     int64
	NotificationChannel SpendAlertNotificationChannel
}

type SpendAlertListRequest struct {
	After  string
	Before string
	Limit  int64
	Order  string
}

type SpendAlertListResponse struct {
	Items   []SpendAlert
	HasMore bool
	LastID  string
}

type CertificateDetails struct {
	Content   string
	ExpiresAt int64
	ValidAt   int64
}

type Certificate struct {
	ID                 string
	Name               string
	Object             string
	Active             bool
	CertificateDetails CertificateDetails
	CreatedAt          int64
}

type CertificateCreateRequest struct {
	Name        string
	Certificate string
}

type CertificateUpdateRequest struct {
	Name string
}

type CertificateListRequest struct {
	After string
	Limit int64
	Order string
}

type CertificateListResponse struct {
	Items   []Certificate
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

	CreateInvite(ctx context.Context, req InviteCreateRequest) (*Invite, error)
	GetInvite(ctx context.Context, id string) (*Invite, error)
	ListInvites(ctx context.Context, req InviteListRequest) (*InviteListResponse, error)
	DeleteInvite(ctx context.Context, id string) error

	GetOrganizationDataRetention(ctx context.Context) (*DataRetention, error)
	UpdateOrganizationDataRetention(ctx context.Context, req DataRetentionUpdateRequest) (*DataRetention, error)

	GetOrganizationSpendLimit(ctx context.Context) (*SpendLimit, error)
	UpdateOrganizationSpendLimit(ctx context.Context, req SpendLimitUpdateRequest) (*SpendLimit, error)
	DeleteOrganizationSpendLimit(ctx context.Context) error

	CreateOrganizationSpendAlert(ctx context.Context, req SpendAlertCreateRequest) (*SpendAlert, error)
	GetOrganizationSpendAlert(ctx context.Context, id string) (*SpendAlert, error)
	ListOrganizationSpendAlerts(ctx context.Context, req SpendAlertListRequest) (*SpendAlertListResponse, error)
	UpdateOrganizationSpendAlert(ctx context.Context, id string, req SpendAlertUpdateRequest) (*SpendAlert, error)
	DeleteOrganizationSpendAlert(ctx context.Context, id string) error

	CreateOrganizationCertificate(ctx context.Context, req CertificateCreateRequest) (*Certificate, error)
	GetOrganizationCertificate(ctx context.Context, id string, includeContent bool) (*Certificate, error)
	ListOrganizationCertificates(ctx context.Context, req CertificateListRequest) (*CertificateListResponse, error)
	UpdateOrganizationCertificate(ctx context.Context, id string, req CertificateUpdateRequest) (*Certificate, error)
	SetOrganizationCertificatesActive(ctx context.Context, ids []string, active bool) ([]Certificate, error)
	DeleteOrganizationCertificate(ctx context.Context, id string) error

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
