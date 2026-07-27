package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const DefaultBaseURL = "https://api.openai.com/v1"

type OpenAIAdminClient struct {
	client openai.Client
}

type Settings struct {
	AdminAPIKey    string
	BaseURL        string
	OrganizationID string
	ProjectID      string
}

func New(adminAPIKey, baseURL, userAgent string, timeout time.Duration) *OpenAIAdminClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return NewWithSettings(Settings{AdminAPIKey: adminAPIKey, BaseURL: baseURL}, userAgent, timeout)
}

func NewWithSettings(settings Settings, userAgent string, timeout time.Duration) *OpenAIAdminClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return newWithHTTPClient(settings, userAgent, &http.Client{Timeout: timeout})
}

func newWithHTTPClient(settings Settings, userAgent string, httpClient *http.Client) *OpenAIAdminClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	opts := []option.RequestOption{
		option.WithAdminAPIKey(settings.AdminAPIKey),
		option.WithHTTPClient(httpClient),
	}
	if strings.TrimSpace(settings.BaseURL) != "" {
		opts = append(opts, option.WithBaseURL(strings.TrimRight(settings.BaseURL, "/")))
	}
	if strings.TrimSpace(settings.OrganizationID) != "" {
		opts = append(opts, option.WithOrganization(strings.TrimSpace(settings.OrganizationID)))
	}
	if strings.TrimSpace(settings.ProjectID) != "" {
		opts = append(opts, option.WithProject(strings.TrimSpace(settings.ProjectID)))
	}
	if strings.TrimSpace(userAgent) != "" {
		opts = append(opts, option.WithHeader("User-Agent", userAgent))
	}
	return &OpenAIAdminClient{client: openai.NewClient(opts...)}
}

func (c *OpenAIAdminClient) CreateProject(ctx context.Context, req ProjectCreateRequest) (*Project, error) {
	params := openai.AdminOrganizationProjectNewParams{Name: req.Name}
	if req.ExternalKeyID != "" {
		params.ExternalKeyID = openai.String(req.ExternalKeyID)
	}
	if req.Geography != "" {
		params.Geography = openai.String(req.Geography)
	}
	project, err := c.client.Admin.Organization.Projects.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapProject(project)
}

func (c *OpenAIAdminClient) GetProject(ctx context.Context, id string) (*Project, error) {
	project, err := c.client.Admin.Organization.Projects.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapProject(project)
}

func (c *OpenAIAdminClient) ListProjects(ctx context.Context, req ProjectListRequest) (*ProjectListResponse, error) {
	params := openai.AdminOrganizationProjectListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	if req.IncludeArchived {
		params.IncludeArchived = openai.Bool(true)
	}
	page, err := c.client.Admin.Organization.Projects.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &ProjectListResponse{
		Items:   make([]Project, 0, len(page.Data)),
		HasMore: page.HasMore,
		LastID:  page.LastID,
	}
	for _, project := range page.Data {
		mapped, err := mapProject(&project)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateProject(ctx context.Context, id string, req ProjectUpdateRequest) (*Project, error) {
	params := openai.AdminOrganizationProjectUpdateParams{}
	if req.Name != "" {
		params.Name = openai.String(req.Name)
	}
	if req.ExternalKeyID != "" {
		params.ExternalKeyID = openai.String(req.ExternalKeyID)
	}
	if req.Geography != "" {
		params.Geography = openai.String(req.Geography)
	}
	project, err := c.client.Admin.Organization.Projects.Update(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return mapProject(project)
}

func (c *OpenAIAdminClient) ArchiveProject(ctx context.Context, id string) (*Project, error) {
	project, err := c.client.Admin.Organization.Projects.Archive(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapProject(project)
}

func (c *OpenAIAdminClient) CreateServiceAccount(ctx context.Context, projectID string, req ServiceAccountCreateRequest) (*ServiceAccount, error) {
	params := openai.AdminOrganizationProjectServiceAccountNewParams{
		Name:                     req.Name,
		CreateServiceAccountOnly: openai.Bool(true),
	}
	created, err := c.client.Admin.Organization.Projects.ServiceAccounts.New(ctx, projectID, params)
	if err != nil {
		return nil, err
	}
	account := mapServiceAccount(created.ID, created.Name, string(created.Role), created.CreatedAt)
	if err := validateServiceAccount(account); err != nil {
		return nil, err
	}
	if req.Role != "" && req.Role != account.Role {
		return c.UpdateServiceAccount(ctx, projectID, account.ID, ServiceAccountUpdateRequest{Name: req.Name, Role: req.Role})
	}
	return account, nil
}

func (c *OpenAIAdminClient) GetServiceAccount(ctx context.Context, projectID, serviceAccountID string) (*ServiceAccount, error) {
	account, err := c.client.Admin.Organization.Projects.ServiceAccounts.Get(ctx, projectID, serviceAccountID)
	if err != nil {
		return nil, err
	}
	mapped := mapServiceAccount(account.ID, account.Name, string(account.Role), account.CreatedAt)
	if err := validateServiceAccount(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) ListServiceAccounts(ctx context.Context, projectID string, req ServiceAccountListRequest) (*ServiceAccountListResponse, error) {
	params := openai.AdminOrganizationProjectServiceAccountListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	page, err := c.client.Admin.Organization.Projects.ServiceAccounts.List(ctx, projectID, params)
	if err != nil {
		return nil, err
	}
	resp := &ServiceAccountListResponse{
		Items:   make([]ServiceAccount, 0, len(page.Data)),
		HasMore: page.HasMore,
		LastID:  page.LastID,
	}
	for _, account := range page.Data {
		mapped := mapServiceAccount(account.ID, account.Name, string(account.Role), account.CreatedAt)
		if err := validateServiceAccount(mapped); err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateServiceAccount(ctx context.Context, projectID, serviceAccountID string, req ServiceAccountUpdateRequest) (*ServiceAccount, error) {
	params := openai.AdminOrganizationProjectServiceAccountUpdateParams{}
	if req.Name != "" {
		params.Name = openai.String(req.Name)
	}
	switch req.Role {
	case "owner":
		params.Role = openai.AdminOrganizationProjectServiceAccountUpdateParamsRoleOwner
	case "member":
		params.Role = openai.AdminOrganizationProjectServiceAccountUpdateParamsRoleMember
	}
	account, err := c.client.Admin.Organization.Projects.ServiceAccounts.Update(ctx, projectID, serviceAccountID, params)
	if err != nil {
		return nil, err
	}
	mapped := mapServiceAccount(account.ID, account.Name, string(account.Role), account.CreatedAt)
	if err := validateServiceAccount(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) DeleteServiceAccount(ctx context.Context, projectID, serviceAccountID string) error {
	_, err := c.client.Admin.Organization.Projects.ServiceAccounts.Delete(ctx, projectID, serviceAccountID)
	return err
}

func (c *OpenAIAdminClient) CreateServiceAccountAPIKey(ctx context.Context, projectID, serviceAccountID string, req ServiceAccountAPIKeyCreateRequest) (*ServiceAccountAPIKeyCreateResponse, error) {
	params := openai.AdminOrganizationProjectServiceAccountAPIKeyNewParams{
		Name:   openai.String(req.Name),
		Scopes: append([]string(nil), req.Scopes...),
	}
	created, err := c.client.Admin.Organization.Projects.ServiceAccounts.APIKeys.New(ctx, projectID, serviceAccountID, params)
	if err != nil {
		return nil, err
	}
	mapped := &ServiceAccountAPIKeyCreateResponse{ID: created.ID, Name: created.Name, Value: created.Value, CreatedAt: created.CreatedAt}
	if err := validateServiceAccountAPIKeyCreate(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) GetProjectAPIKey(ctx context.Context, projectID, apiKeyID string) (*APIKey, error) {
	apiKey, err := c.client.Admin.Organization.Projects.APIKeys.Get(ctx, projectID, apiKeyID)
	if err != nil {
		return nil, err
	}
	ownerID := ""
	ownerName := ""
	if apiKey.Owner.Type == "service_account" {
		ownerID = apiKey.Owner.ServiceAccount.ID
		ownerName = apiKey.Owner.ServiceAccount.Name
	}
	mapped := &APIKey{
		ID:                 apiKey.ID,
		Name:               apiKey.Name,
		RedactedValue:      apiKey.RedactedValue,
		OwnerType:          apiKey.Owner.Type,
		OwnerID:            ownerID,
		OwnerName:          ownerName,
		OwnerProjectAccess: string(apiKey.OwnerProjectAccess),
		CreatedAt:          apiKey.CreatedAt,
		LastUsedAt:         apiKey.LastUsedAt,
	}
	if err := validateAPIKey(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) DeleteProjectAPIKey(ctx context.Context, projectID, apiKeyID string) error {
	_, err := c.client.Admin.Organization.Projects.APIKeys.Delete(ctx, projectID, apiKeyID)
	return err
}

func IsNotFound(err error) bool {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func ErrorSummary(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		if strings.TrimSpace(apiErr.Message) != "" {
			return fmt.Sprintf("OpenAI API returned HTTP %d: %s", apiErr.StatusCode, apiErr.Message)
		}
		return fmt.Sprintf("OpenAI API returned HTTP %d", apiErr.StatusCode)
	}
	return err.Error()
}

func mapProject(project *openai.Project) (*Project, error) {
	if project == nil {
		return nil, fmt.Errorf("openai project response was empty")
	}
	mapped := &Project{
		ID:            project.ID,
		Name:          project.Name,
		Status:        project.Status,
		ExternalKeyID: project.ExternalKeyID,
		CreatedAt:     project.CreatedAt,
		ArchivedAt:    project.ArchivedAt,
	}
	if err := requireNonEmpty("project", "id", mapped.ID); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapServiceAccount(id, name, role string, createdAt int64) *ServiceAccount {
	return &ServiceAccount{ID: id, Name: name, Role: role, CreatedAt: createdAt}
}

func validateServiceAccount(account *ServiceAccount) error {
	if account == nil {
		return fmt.Errorf("openai service account response was empty")
	}
	if err := requireNonEmpty("service account", "id", account.ID); err != nil {
		return err
	}
	return requireNonEmpty("service account", "name", account.Name)
}

func validateServiceAccountAPIKeyCreate(apiKey *ServiceAccountAPIKeyCreateResponse) error {
	if apiKey == nil {
		return fmt.Errorf("openai service account api key response was empty")
	}
	if err := requireNonEmpty("service account api key", "id", apiKey.ID); err != nil {
		return err
	}
	if err := requireNonEmpty("service account api key", "name", apiKey.Name); err != nil {
		return err
	}
	return requireNonEmpty("service account api key", "value", apiKey.Value)
}

func validateAPIKey(apiKey *APIKey) error {
	if apiKey == nil {
		return fmt.Errorf("openai project api key response was empty")
	}
	if err := requireNonEmpty("project api key", "id", apiKey.ID); err != nil {
		return err
	}
	return requireNonEmpty("project api key", "name", apiKey.Name)
}
