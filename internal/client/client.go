package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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

func terraformLogEnablesOpenAIDebugLog(tfLog string) bool {
	switch strings.ToUpper(strings.TrimSpace(tfLog)) {
	case "TRACE", "JSON":
		return true
	default:
		return false
	}
}

func newWithHTTPClient(settings Settings, userAgent string, httpClient *http.Client) *OpenAIAdminClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	opts := []option.RequestOption{
		option.WithAdminAPIKey(settings.AdminAPIKey),
		option.WithHTTPClient(httpClient),
	}
	if terraformLogEnablesOpenAIDebugLog(os.Getenv("TF_LOG")) {
		opts = append(opts, option.WithDebugLog(nil))
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
		Name: req.Name,
	}
	if req.CreateServiceAccountOnly {
		params.CreateServiceAccountOnly = openai.Bool(true)
	}
	created, err := c.client.Admin.Organization.Projects.ServiceAccounts.New(ctx, projectID, params)
	if err != nil {
		return nil, err
	}
	account := mapServiceAccount(created.ID, created.Name, string(created.Role), created.CreatedAt)
	if err := validateServiceAccount(account); err != nil {
		return nil, err
	}
	if !req.CreateServiceAccountOnly {
		apiKey := mapServiceAccountAPIKeyCreate(created.APIKey.ID, created.APIKey.Name, created.APIKey.Value, created.APIKey.CreatedAt)
		if err := validateServiceAccountAPIKeyCreate(apiKey); err != nil {
			return nil, err
		}
		account.APIKey = apiKey
	}
	if req.Role != "" && req.Role != account.Role {
		updated, err := c.UpdateServiceAccount(ctx, projectID, account.ID, ServiceAccountUpdateRequest{Name: req.Name, Role: req.Role})
		if err != nil {
			return nil, err
		}
		updated.APIKey = account.APIKey
		return updated, nil
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
	mapped := mapServiceAccountAPIKeyCreate(created.ID, created.Name, created.Value, created.CreatedAt)
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

func (c *OpenAIAdminClient) CreateAdminAPIKey(ctx context.Context, req AdminAPIKeyCreateRequest) (*AdminAPIKey, error) {
	params := openai.AdminOrganizationAdminAPIKeyNewParams{Name: req.Name}
	if req.ExpiresInSeconds > 0 {
		params.ExpiresInSeconds = openai.Int(req.ExpiresInSeconds)
	}
	created, err := c.client.Admin.Organization.AdminAPIKeys.New(ctx, params)
	if err != nil {
		return nil, err
	}
	mapped, err := mapAdminAPIKey(&created.AdminAPIKey)
	if err != nil {
		return nil, err
	}
	mapped.Value = created.Value
	if err := validateAdminAPIKeyCreate(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) GetAdminAPIKey(ctx context.Context, id string) (*AdminAPIKey, error) {
	apiKey, err := c.client.Admin.Organization.AdminAPIKeys.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapAdminAPIKey(apiKey)
}

func (c *OpenAIAdminClient) ListAdminAPIKeys(ctx context.Context, req AdminAPIKeyListRequest) (*AdminAPIKeyListResponse, error) {
	params := openai.AdminOrganizationAdminAPIKeyListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationAdminAPIKeyListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationAdminAPIKeyListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.AdminAPIKeys.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &AdminAPIKeyListResponse{Items: make([]AdminAPIKey, 0, len(page.Data)), HasMore: page.HasMore}
	for _, apiKey := range page.Data {
		mapped, err := mapAdminAPIKey(&apiKey)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	if len(resp.Items) > 0 {
		resp.LastID = resp.Items[len(resp.Items)-1].ID
	}
	return resp, nil
}

func (c *OpenAIAdminClient) DeleteAdminAPIKey(ctx context.Context, id string) error {
	deleted, err := c.client.Admin.Organization.AdminAPIKeys.Delete(ctx, id)
	if err != nil {
		return err
	}
	if deleted == nil {
		return fmt.Errorf("openai admin api key delete response for %q was empty", id)
	}
	if !deleted.Deleted {
		return fmt.Errorf("openai admin api key %q was not revoked: delete response returned deleted=false", id)
	}
	if deleted.ID != "" && deleted.ID != id {
		return fmt.Errorf("openai admin api key delete response id mismatch: requested %q, got %q", id, deleted.ID)
	}
	return nil
}

func (c *OpenAIAdminClient) CreateInvite(ctx context.Context, req InviteCreateRequest) (*Invite, error) {
	params := openai.AdminOrganizationInviteNewParams{
		Email: req.Email,
		Role:  openai.AdminOrganizationInviteNewParamsRole(req.Role),
	}
	if req.Projects != nil {
		params.Projects = make([]openai.AdminOrganizationInviteNewParamsProject, 0, len(req.Projects))
		for _, project := range req.Projects {
			params.Projects = append(params.Projects, openai.AdminOrganizationInviteNewParamsProject{
				ID:   project.ID,
				Role: project.Role,
			})
		}
	}
	invite, err := c.client.Admin.Organization.Invites.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapInvite(invite)
}

func (c *OpenAIAdminClient) GetInvite(ctx context.Context, id string) (*Invite, error) {
	invite, err := c.client.Admin.Organization.Invites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapInvite(invite)
}

func (c *OpenAIAdminClient) ListInvites(ctx context.Context, req InviteListRequest) (*InviteListResponse, error) {
	params := openai.AdminOrganizationInviteListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	page, err := c.client.Admin.Organization.Invites.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &InviteListResponse{Items: make([]Invite, 0, len(page.Data)), HasMore: page.HasMore, LastID: page.LastID}
	for _, invite := range page.Data {
		mapped, err := mapInvite(&invite)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	if resp.LastID == "" && len(resp.Items) > 0 {
		resp.LastID = resp.Items[len(resp.Items)-1].ID
	}
	return resp, nil
}

func (c *OpenAIAdminClient) DeleteInvite(ctx context.Context, id string) error {
	deleted, err := c.client.Admin.Organization.Invites.Delete(ctx, id)
	if err != nil {
		return err
	}
	if deleted == nil {
		return fmt.Errorf("openai invite delete response for %q was empty", id)
	}
	if !deleted.Deleted {
		return fmt.Errorf("openai invite %q was not deleted: delete response returned deleted=false", id)
	}
	if deleted.ID != "" && deleted.ID != id {
		return fmt.Errorf("openai invite delete response id mismatch: requested %q, got %q", id, deleted.ID)
	}
	return nil
}

func (c *OpenAIAdminClient) GetOrganizationDataRetention(ctx context.Context) (*DataRetention, error) {
	retention, err := c.client.Admin.Organization.DataRetention.Get(ctx)
	if err != nil {
		return nil, err
	}
	return mapOrganizationDataRetention(retention)
}

func (c *OpenAIAdminClient) UpdateOrganizationDataRetention(ctx context.Context, req DataRetentionUpdateRequest) (*DataRetention, error) {
	retention, err := c.client.Admin.Organization.DataRetention.Update(ctx, openai.AdminOrganizationDataRetentionUpdateParams{
		RetentionType: openai.AdminOrganizationDataRetentionUpdateParamsRetentionType(req.Type),
	})
	if err != nil {
		return nil, err
	}
	return mapOrganizationDataRetention(retention)
}

func (c *OpenAIAdminClient) GetOrganizationSpendLimit(ctx context.Context) (*SpendLimit, error) {
	limit, err := c.client.Admin.Organization.SpendLimit.Get(ctx)
	if err != nil {
		return nil, err
	}
	return mapOrganizationSpendLimit(limit)
}

func (c *OpenAIAdminClient) UpdateOrganizationSpendLimit(ctx context.Context, req SpendLimitUpdateRequest) (*SpendLimit, error) {
	params := openai.AdminOrganizationSpendLimitUpdateParams{
		Currency:        openai.AdminOrganizationSpendLimitUpdateParamsCurrency(req.Currency),
		Interval:        openai.AdminOrganizationSpendLimitUpdateParamsInterval(req.Interval),
		ThresholdAmount: req.ThresholdAmount,
	}
	if params.Currency == "" {
		params.Currency = openai.AdminOrganizationSpendLimitUpdateParamsCurrencyUsd
	}
	if params.Interval == "" {
		params.Interval = openai.AdminOrganizationSpendLimitUpdateParamsIntervalMonth
	}
	limit, err := c.client.Admin.Organization.SpendLimit.Update(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapOrganizationSpendLimit(limit)
}

func (c *OpenAIAdminClient) DeleteOrganizationSpendLimit(ctx context.Context) error {
	deleted, err := c.client.Admin.Organization.SpendLimit.Delete(ctx)
	if err != nil {
		return err
	}
	if deleted == nil {
		return fmt.Errorf("openai organization spend limit delete response was empty")
	}
	if !deleted.Deleted {
		return fmt.Errorf("openai organization spend limit was not deleted: delete response returned deleted=false")
	}
	return nil
}

func (c *OpenAIAdminClient) CreateOrganizationSpendAlert(ctx context.Context, req SpendAlertCreateRequest) (*SpendAlert, error) {
	params := openai.AdminOrganizationSpendAlertNewParams{
		Currency:        openai.AdminOrganizationSpendAlertNewParamsCurrency(req.Currency),
		Interval:        openai.AdminOrganizationSpendAlertNewParamsInterval(req.Interval),
		ThresholdAmount: req.ThresholdAmount,
		NotificationChannel: openai.AdminOrganizationSpendAlertNewParamsNotificationChannel{
			Recipients: append([]string(nil), req.NotificationChannel.Recipients...),
		},
	}
	if params.Currency == "" {
		params.Currency = openai.AdminOrganizationSpendAlertNewParamsCurrencyUsd
	}
	if params.Interval == "" {
		params.Interval = openai.AdminOrganizationSpendAlertNewParamsIntervalMonth
	}
	if req.NotificationChannel.SubjectPrefix != "" {
		params.NotificationChannel.SubjectPrefix = openai.String(req.NotificationChannel.SubjectPrefix)
	}
	alert, err := c.client.Admin.Organization.SpendAlerts.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapOrganizationSpendAlert(alert)
}

func (c *OpenAIAdminClient) GetOrganizationSpendAlert(ctx context.Context, id string) (*SpendAlert, error) {
	alert, err := c.client.Admin.Organization.SpendAlerts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapOrganizationSpendAlert(alert)
}

func (c *OpenAIAdminClient) ListOrganizationSpendAlerts(ctx context.Context, req SpendAlertListRequest) (*SpendAlertListResponse, error) {
	params := openai.AdminOrganizationSpendAlertListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Before != "" {
		params.Before = openai.String(req.Before)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationSpendAlertListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationSpendAlertListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.SpendAlerts.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &SpendAlertListResponse{Items: make([]SpendAlert, 0, len(page.Data)), HasMore: page.HasMore, LastID: page.LastID}
	for _, alert := range page.Data {
		mapped, err := mapOrganizationSpendAlert(&alert)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	if resp.LastID == "" && len(resp.Items) > 0 {
		resp.LastID = resp.Items[len(resp.Items)-1].ID
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateOrganizationSpendAlert(ctx context.Context, id string, req SpendAlertUpdateRequest) (*SpendAlert, error) {
	params := openai.AdminOrganizationSpendAlertUpdateParams{
		Currency:        openai.AdminOrganizationSpendAlertUpdateParamsCurrency(req.Currency),
		Interval:        openai.AdminOrganizationSpendAlertUpdateParamsInterval(req.Interval),
		ThresholdAmount: req.ThresholdAmount,
		NotificationChannel: openai.AdminOrganizationSpendAlertUpdateParamsNotificationChannel{
			Recipients: append([]string(nil), req.NotificationChannel.Recipients...),
		},
	}
	if params.Currency == "" {
		params.Currency = openai.AdminOrganizationSpendAlertUpdateParamsCurrencyUsd
	}
	if params.Interval == "" {
		params.Interval = openai.AdminOrganizationSpendAlertUpdateParamsIntervalMonth
	}
	if req.NotificationChannel.SubjectPrefix != "" {
		params.NotificationChannel.SubjectPrefix = openai.String(req.NotificationChannel.SubjectPrefix)
	}
	alert, err := c.client.Admin.Organization.SpendAlerts.Update(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return mapOrganizationSpendAlert(alert)
}

func (c *OpenAIAdminClient) DeleteOrganizationSpendAlert(ctx context.Context, id string) error {
	deleted, err := c.client.Admin.Organization.SpendAlerts.Delete(ctx, id)
	if err != nil {
		return err
	}
	if deleted == nil {
		return fmt.Errorf("openai organization spend alert delete response for %q was empty", id)
	}
	if !deleted.Deleted {
		return fmt.Errorf("openai organization spend alert %q was not deleted: delete response returned deleted=false", id)
	}
	if deleted.ID != "" && deleted.ID != id {
		return fmt.Errorf("openai organization spend alert delete response id mismatch: requested %q, got %q", id, deleted.ID)
	}
	return nil
}

func (c *OpenAIAdminClient) CreateOrganizationCertificate(ctx context.Context, req CertificateCreateRequest) (*Certificate, error) {
	params := openai.AdminOrganizationCertificateNewParams{Certificate: req.Certificate}
	if req.Name != "" {
		params.Name = openai.String(req.Name)
	}
	certificate, err := c.client.Admin.Organization.Certificates.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapCertificate(certificate)
}

func (c *OpenAIAdminClient) GetOrganizationCertificate(ctx context.Context, id string, includeContent bool) (*Certificate, error) {
	params := openai.AdminOrganizationCertificateGetParams{}
	if includeContent {
		params.Include = []string{"content"}
	}
	certificate, err := c.client.Admin.Organization.Certificates.Get(ctx, id, params)
	if err != nil {
		return nil, err
	}
	mapped, err := mapCertificate(certificate)
	if err != nil {
		return nil, err
	}
	active, found, err := c.organizationCertificateActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if found {
		mapped.Active = active
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) ListOrganizationCertificates(ctx context.Context, req CertificateListRequest) (*CertificateListResponse, error) {
	params := openai.AdminOrganizationCertificateListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationCertificateListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationCertificateListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Certificates.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &CertificateListResponse{Items: make([]Certificate, 0, len(page.Data)), HasMore: page.HasMore, LastID: page.LastID}
	for _, certificate := range page.Data {
		mapped, err := mapOrganizationCertificateListItem(&certificate)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	if resp.LastID == "" && len(resp.Items) > 0 {
		resp.LastID = resp.Items[len(resp.Items)-1].ID
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateOrganizationCertificate(ctx context.Context, id string, req CertificateUpdateRequest) (*Certificate, error) {
	params := openai.AdminOrganizationCertificateUpdateParams{}
	if req.Name != "" {
		params.Name = openai.String(req.Name)
	}
	certificate, err := c.client.Admin.Organization.Certificates.Update(ctx, id, params)
	if err != nil {
		return nil, err
	}
	mapped, err := mapCertificate(certificate)
	if err != nil {
		return nil, err
	}
	active, found, err := c.organizationCertificateActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if found {
		mapped.Active = active
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) SetOrganizationCertificatesActive(ctx context.Context, ids []string, active bool) ([]Certificate, error) {
	cleanIDs := append([]string(nil), ids...)
	if active {
		page, err := c.client.Admin.Organization.Certificates.Activate(ctx, openai.AdminOrganizationCertificateActivateParams{CertificateIDs: cleanIDs})
		if err != nil {
			return nil, err
		}
		certificates := make([]Certificate, 0, len(page.Data))
		for _, certificate := range page.Data {
			mapped, err := mapOrganizationCertificateActivateItem(&certificate)
			if err != nil {
				return nil, err
			}
			certificates = append(certificates, *mapped)
		}
		return certificates, nil
	}
	page, err := c.client.Admin.Organization.Certificates.Deactivate(ctx, openai.AdminOrganizationCertificateDeactivateParams{CertificateIDs: cleanIDs})
	if err != nil {
		return nil, err
	}
	certificates := make([]Certificate, 0, len(page.Data))
	for _, certificate := range page.Data {
		mapped, err := mapOrganizationCertificateDeactivateItem(&certificate)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, *mapped)
	}
	return certificates, nil
}

func (c *OpenAIAdminClient) DeleteOrganizationCertificate(ctx context.Context, id string) error {
	deleted, err := c.client.Admin.Organization.Certificates.Delete(ctx, id)
	if err != nil {
		return err
	}
	if deleted == nil {
		return fmt.Errorf("openai organization certificate delete response for %q was empty", id)
	}
	if deleted.ID != "" && deleted.ID != id {
		return fmt.Errorf("openai organization certificate delete response id mismatch: requested %q, got %q", id, deleted.ID)
	}
	return nil
}

func (c *OpenAIAdminClient) organizationCertificateActive(ctx context.Context, id string) (bool, bool, error) {
	page, err := c.client.Admin.Organization.Certificates.List(ctx, openai.AdminOrganizationCertificateListParams{Limit: openai.Int(100)})
	if err != nil {
		return false, false, err
	}
	for _, certificate := range page.Data {
		if certificate.ID == id {
			return certificate.Active, true, nil
		}
	}
	return false, false, nil
}

func (c *OpenAIAdminClient) GetOrganizationUser(ctx context.Context, userID string) (*OrganizationUser, error) {
	user, err := c.client.Admin.Organization.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapOrganizationUser(user)
}

func (c *OpenAIAdminClient) ListOrganizationUsers(ctx context.Context, req OrganizationUserListRequest) (*OrganizationUserListResponse, error) {
	params := openai.AdminOrganizationUserListParams{Emails: append([]string(nil), req.Emails...)}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	page, err := c.client.Admin.Organization.Users.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &OrganizationUserListResponse{Items: make([]OrganizationUser, 0, len(page.Data)), HasMore: page.HasMore, LastID: page.LastID}
	for _, user := range page.Data {
		mapped, err := mapOrganizationUser(&user)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	if resp.LastID == "" && len(resp.Items) > 0 {
		resp.LastID = resp.Items[len(resp.Items)-1].ID
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateOrganizationUser(ctx context.Context, userID string, req OrganizationUserUpdateRequest) (*OrganizationUser, error) {
	params := openai.AdminOrganizationUserUpdateParams{}
	if req.Role != "" {
		params.Role = openai.String(req.Role)
	}
	if req.RoleID != "" {
		params.RoleID = openai.String(req.RoleID)
	}
	if req.DeveloperPersona != "" {
		params.DeveloperPersona = openai.String(req.DeveloperPersona)
	}
	if req.TechnicalLevel != "" {
		params.TechnicalLevel = openai.String(req.TechnicalLevel)
	}
	user, err := c.client.Admin.Organization.Users.Update(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	return mapOrganizationUser(user)
}

func (c *OpenAIAdminClient) DeleteOrganizationUser(ctx context.Context, userID string) error {
	_, err := c.client.Admin.Organization.Users.Delete(ctx, userID)
	return err
}

func (c *OpenAIAdminClient) CreateOrganizationGroup(ctx context.Context, req OrganizationGroupCreateRequest) (*OrganizationGroup, error) {
	group, err := c.client.Admin.Organization.Groups.New(ctx, openai.AdminOrganizationGroupNewParams{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return mapOrganizationGroup(group)
}

func (c *OpenAIAdminClient) GetOrganizationGroup(ctx context.Context, groupID string) (*OrganizationGroup, error) {
	group, err := c.client.Admin.Organization.Groups.Get(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return mapOrganizationGroup(group)
}

func (c *OpenAIAdminClient) ListOrganizationGroups(ctx context.Context, req OrganizationGroupListRequest) (*OrganizationGroupListResponse, error) {
	params := openai.AdminOrganizationGroupListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationGroupListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationGroupListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Groups.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &OrganizationGroupListResponse{Items: make([]OrganizationGroup, 0, len(page.Data)), HasMore: page.HasMore, Next: page.Next}
	for _, group := range page.Data {
		mapped, err := mapOrganizationGroup(&group)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateOrganizationGroup(ctx context.Context, groupID string, req OrganizationGroupUpdateRequest) (*OrganizationGroup, error) {
	updated, err := c.client.Admin.Organization.Groups.Update(ctx, groupID, openai.AdminOrganizationGroupUpdateParams{Name: req.Name})
	if err != nil {
		return nil, err
	}
	mapped := &OrganizationGroup{ID: updated.ID, Name: updated.Name, IsScimManaged: updated.IsScimManaged, CreatedAt: updated.CreatedAt}
	if err := validateOrganizationGroup(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) DeleteOrganizationGroup(ctx context.Context, groupID string) error {
	_, err := c.client.Admin.Organization.Groups.Delete(ctx, groupID)
	return err
}

func (c *OpenAIAdminClient) CreateOrganizationGroupUser(ctx context.Context, groupID string, req OrganizationGroupUserCreateRequest) (*OrganizationGroupUser, error) {
	created, err := c.client.Admin.Organization.Groups.Users.New(ctx, groupID, openai.AdminOrganizationGroupUserNewParams{UserID: req.UserID})
	if err != nil {
		return nil, err
	}
	return c.GetOrganizationGroupUser(ctx, created.GroupID, created.UserID)
}

func (c *OpenAIAdminClient) GetOrganizationGroupUser(ctx context.Context, groupID, userID string) (*OrganizationGroupUser, error) {
	user, err := c.client.Admin.Organization.Groups.Users.Get(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	mapped := &OrganizationGroupUser{ID: user.ID, Name: user.Name, Email: user.Email, IsServiceAccount: user.IsServiceAccount, Picture: user.Picture, UserType: string(user.UserType)}
	if err := validateOrganizationGroupUser(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) ListOrganizationGroupUsers(ctx context.Context, groupID string, req OrganizationGroupUserListRequest) (*OrganizationGroupUserListResponse, error) {
	params := openai.AdminOrganizationGroupUserListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationGroupUserListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationGroupUserListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Groups.Users.List(ctx, groupID, params)
	if err != nil {
		return nil, err
	}
	resp := &OrganizationGroupUserListResponse{Items: make([]OrganizationGroupUser, 0, len(page.Data)), HasMore: page.HasMore, Next: page.Next}
	for _, user := range page.Data {
		mapped := OrganizationGroupUser{ID: user.ID, Name: user.Name, Email: user.Email}
		if err := validateOrganizationGroupUser(&mapped); err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) DeleteOrganizationGroupUser(ctx context.Context, groupID, userID string) error {
	_, err := c.client.Admin.Organization.Groups.Users.Delete(ctx, groupID, userID)
	return err
}

func (c *OpenAIAdminClient) CreateOrganizationRole(ctx context.Context, req RoleCreateRequest) (*Role, error) {
	params := openai.AdminOrganizationRoleNewParams{RoleName: req.Name, Permissions: append([]string(nil), req.Permissions...)}
	if req.Description != "" {
		params.Description = openai.String(req.Description)
	}
	role, err := c.client.Admin.Organization.Roles.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapRole(role)
}

func (c *OpenAIAdminClient) GetOrganizationRole(ctx context.Context, roleID string) (*Role, error) {
	role, err := c.client.Admin.Organization.Roles.Get(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return mapRole(role)
}

func (c *OpenAIAdminClient) ListOrganizationRoles(ctx context.Context, req RoleListRequest) (*RoleListResponse, error) {
	params := openai.AdminOrganizationRoleListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationRoleListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationRoleListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Roles.List(ctx, params)
	if err != nil {
		return nil, err
	}
	resp := &RoleListResponse{Items: make([]Role, 0, len(page.Data)), HasMore: page.HasMore, Next: page.Next}
	for _, role := range page.Data {
		mapped, err := mapRole(&role)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateOrganizationRole(ctx context.Context, roleID string, req RoleUpdateRequest) (*Role, error) {
	params := openai.AdminOrganizationRoleUpdateParams{Permissions: append([]string(nil), req.Permissions...)}
	if req.Name != "" {
		params.RoleName = openai.String(req.Name)
	}
	if req.Description != "" {
		params.Description = openai.String(req.Description)
	}
	role, err := c.client.Admin.Organization.Roles.Update(ctx, roleID, params)
	if err != nil {
		return nil, err
	}
	return mapRole(role)
}

func (c *OpenAIAdminClient) DeleteOrganizationRole(ctx context.Context, roleID string) error {
	_, err := c.client.Admin.Organization.Roles.Delete(ctx, roleID)
	return err
}

func (c *OpenAIAdminClient) CreateProjectRole(ctx context.Context, projectID string, req RoleCreateRequest) (*Role, error) {
	params := openai.AdminOrganizationProjectRoleNewParams{RoleName: req.Name, Permissions: append([]string(nil), req.Permissions...)}
	if req.Description != "" {
		params.Description = openai.String(req.Description)
	}
	role, err := c.client.Admin.Organization.Projects.Roles.New(ctx, projectID, params)
	if err != nil {
		return nil, err
	}
	return mapRole(role)
}

func (c *OpenAIAdminClient) GetProjectRole(ctx context.Context, projectID, roleID string) (*Role, error) {
	role, err := c.client.Admin.Organization.Projects.Roles.Get(ctx, projectID, roleID)
	if err != nil {
		return nil, err
	}
	return mapRole(role)
}

func (c *OpenAIAdminClient) ListProjectRoles(ctx context.Context, projectID string, req RoleListRequest) (*RoleListResponse, error) {
	params := openai.AdminOrganizationProjectRoleListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationProjectRoleListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationProjectRoleListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Projects.Roles.List(ctx, projectID, params)
	if err != nil {
		return nil, err
	}
	resp := &RoleListResponse{Items: make([]Role, 0, len(page.Data)), HasMore: page.HasMore, Next: page.Next}
	for _, role := range page.Data {
		mapped, err := mapRole(&role)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) UpdateProjectRole(ctx context.Context, projectID, roleID string, req RoleUpdateRequest) (*Role, error) {
	params := openai.AdminOrganizationProjectRoleUpdateParams{Permissions: append([]string(nil), req.Permissions...)}
	if req.Name != "" {
		params.RoleName = openai.String(req.Name)
	}
	if req.Description != "" {
		params.Description = openai.String(req.Description)
	}
	role, err := c.client.Admin.Organization.Projects.Roles.Update(ctx, projectID, roleID, params)
	if err != nil {
		return nil, err
	}
	return mapRole(role)
}

func (c *OpenAIAdminClient) DeleteProjectRole(ctx context.Context, projectID, roleID string) error {
	_, err := c.client.Admin.Organization.Projects.Roles.Delete(ctx, projectID, roleID)
	return err
}

func (c *OpenAIAdminClient) CreateOrganizationUserRole(ctx context.Context, userID string, req RoleAssignmentCreateRequest) (*RoleAssignment, error) {
	created, err := c.client.Admin.Organization.Users.Roles.New(ctx, userID, openai.AdminOrganizationUserRoleNewParams{RoleID: req.RoleID})
	if err != nil {
		return nil, err
	}
	mapped, err := mapRoleAssignmentFromRole(&created.Role, userID, "user")
	if err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) GetOrganizationUserRole(ctx context.Context, userID, roleID string) (*RoleAssignment, error) {
	role, err := c.client.Admin.Organization.Users.Roles.Get(ctx, userID, roleID)
	if err != nil {
		return nil, err
	}
	return mapUserRoleAssignmentGet(role, userID)
}

func (c *OpenAIAdminClient) ListOrganizationUserRoles(ctx context.Context, userID string, req RoleAssignmentListRequest) (*RoleAssignmentListResponse, error) {
	params := openai.AdminOrganizationUserRoleListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationUserRoleListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationUserRoleListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Users.Roles.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	resp := &RoleAssignmentListResponse{Items: make([]RoleAssignment, 0, len(page.Data)), HasMore: page.HasMore, Next: page.Next}
	for _, role := range page.Data {
		mapped, err := mapUserRoleAssignmentList(&role, userID)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) DeleteOrganizationUserRole(ctx context.Context, userID, roleID string) error {
	_, err := c.client.Admin.Organization.Users.Roles.Delete(ctx, userID, roleID)
	return err
}

func (c *OpenAIAdminClient) CreateOrganizationGroupRole(ctx context.Context, groupID string, req RoleAssignmentCreateRequest) (*RoleAssignment, error) {
	created, err := c.client.Admin.Organization.Groups.Roles.New(ctx, groupID, openai.AdminOrganizationGroupRoleNewParams{RoleID: req.RoleID})
	if err != nil {
		return nil, err
	}
	mapped, err := mapRoleAssignmentFromRole(&created.Role, groupID, "group")
	if err != nil {
		return nil, err
	}
	return mapped, nil
}

func (c *OpenAIAdminClient) GetOrganizationGroupRole(ctx context.Context, groupID, roleID string) (*RoleAssignment, error) {
	role, err := c.client.Admin.Organization.Groups.Roles.Get(ctx, groupID, roleID)
	if err != nil {
		return nil, err
	}
	return mapGroupRoleAssignmentGet(role, groupID)
}

func (c *OpenAIAdminClient) ListOrganizationGroupRoles(ctx context.Context, groupID string, req RoleAssignmentListRequest) (*RoleAssignmentListResponse, error) {
	params := openai.AdminOrganizationGroupRoleListParams{}
	if req.After != "" {
		params.After = openai.String(req.After)
	}
	if req.Limit > 0 {
		params.Limit = openai.Int(req.Limit)
	}
	switch req.Order {
	case "asc":
		params.Order = openai.AdminOrganizationGroupRoleListParamsOrderAsc
	case "desc":
		params.Order = openai.AdminOrganizationGroupRoleListParamsOrderDesc
	}
	page, err := c.client.Admin.Organization.Groups.Roles.List(ctx, groupID, params)
	if err != nil {
		return nil, err
	}
	resp := &RoleAssignmentListResponse{Items: make([]RoleAssignment, 0, len(page.Data)), HasMore: page.HasMore, Next: page.Next}
	for _, role := range page.Data {
		mapped, err := mapGroupRoleAssignmentList(&role, groupID)
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, *mapped)
	}
	return resp, nil
}

func (c *OpenAIAdminClient) DeleteOrganizationGroupRole(ctx context.Context, groupID, roleID string) error {
	_, err := c.client.Admin.Organization.Groups.Roles.Delete(ctx, groupID, roleID)
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

func mapServiceAccountAPIKeyCreate(id, name, value string, createdAt int64) *ServiceAccountAPIKeyCreateResponse {
	return &ServiceAccountAPIKeyCreateResponse{ID: id, Name: name, Value: value, CreatedAt: createdAt}
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

func mapAdminAPIKey(apiKey *openai.AdminAPIKey) (*AdminAPIKey, error) {
	if apiKey == nil {
		return nil, fmt.Errorf("openai admin api key response was empty")
	}
	mapped := &AdminAPIKey{
		ID:            apiKey.ID,
		Name:          apiKey.Name,
		RedactedValue: apiKey.RedactedValue,
		OwnerType:     apiKey.Owner.Type,
		OwnerID:       apiKey.Owner.ID,
		OwnerName:     apiKey.Owner.Name,
		OwnerRole:     apiKey.Owner.Role,
		CreatedAt:     apiKey.CreatedAt,
		ExpiresAt:     apiKey.ExpiresAt,
		LastUsedAt:    apiKey.LastUsedAt,
	}
	if err := validateAdminAPIKey(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateAdminAPIKey(apiKey *AdminAPIKey) error {
	if apiKey == nil {
		return fmt.Errorf("openai admin api key response was empty")
	}
	if err := requireNonEmpty("admin api key", "id", apiKey.ID); err != nil {
		return err
	}
	return requireNonEmpty("admin api key", "redacted_value", apiKey.RedactedValue)
}

func validateAdminAPIKeyCreate(apiKey *AdminAPIKey) error {
	if err := validateAdminAPIKey(apiKey); err != nil {
		return err
	}
	return requireNonEmpty("admin api key", "value", apiKey.Value)
}

func mapInvite(invite *openai.Invite) (*Invite, error) {
	if invite == nil {
		return nil, fmt.Errorf("openai invite response was empty")
	}
	mapped := &Invite{
		ID:         invite.ID,
		Email:      invite.Email,
		Role:       string(invite.Role),
		Status:     string(invite.Status),
		CreatedAt:  invite.CreatedAt,
		AcceptedAt: invite.AcceptedAt,
		ExpiresAt:  invite.ExpiresAt,
		Projects:   make([]InviteProject, 0, len(invite.Projects)),
	}
	for _, project := range invite.Projects {
		mapped.Projects = append(mapped.Projects, InviteProject{ID: project.ID, Role: project.Role})
	}
	if err := validateInvite(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateInvite(invite *Invite) error {
	if invite == nil {
		return fmt.Errorf("openai invite response was empty")
	}
	if err := requireNonEmpty("invite", "id", invite.ID); err != nil {
		return err
	}
	return requireNonEmpty("invite", "email", invite.Email)
}

func mapOrganizationDataRetention(retention *openai.OrganizationDataRetention) (*DataRetention, error) {
	if retention == nil {
		return nil, fmt.Errorf("openai organization data retention response was empty")
	}
	mapped := &DataRetention{ID: "organization", Type: string(retention.Type)}
	if err := validateDataRetention(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateDataRetention(retention *DataRetention) error {
	if retention == nil {
		return fmt.Errorf("openai organization data retention response was empty")
	}
	return requireNonEmpty("organization data retention", "type", retention.Type)
}

func mapOrganizationSpendLimit(limit *openai.OrganizationSpendLimit) (*SpendLimit, error) {
	if limit == nil {
		return nil, fmt.Errorf("openai organization spend limit response was empty")
	}
	mapped := &SpendLimit{
		ID:                "organization",
		Currency:          string(limit.Currency),
		Interval:          string(limit.Interval),
		ThresholdAmount:   limit.ThresholdAmount,
		EnforcementStatus: limit.Enforcement.Status,
	}
	if err := validateSpendLimit(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateSpendLimit(limit *SpendLimit) error {
	if limit == nil {
		return fmt.Errorf("openai organization spend limit response was empty")
	}
	if err := requireNonEmpty("organization spend limit", "currency", limit.Currency); err != nil {
		return err
	}
	return requireNonEmpty("organization spend limit", "interval", limit.Interval)
}

func mapOrganizationSpendAlert(alert *openai.OrganizationSpendAlert) (*SpendAlert, error) {
	if alert == nil {
		return nil, fmt.Errorf("openai organization spend alert response was empty")
	}
	mapped := &SpendAlert{
		ID:              alert.ID,
		Currency:        string(alert.Currency),
		Interval:        string(alert.Interval),
		ThresholdAmount: alert.ThresholdAmount,
		NotificationChannel: SpendAlertNotificationChannel{
			Recipients:    append([]string(nil), alert.NotificationChannel.Recipients...),
			Type:          string(alert.NotificationChannel.Type),
			SubjectPrefix: alert.NotificationChannel.SubjectPrefix,
		},
	}
	if mapped.NotificationChannel.Type == "" {
		mapped.NotificationChannel.Type = "email"
	}
	if err := validateSpendAlert(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateSpendAlert(alert *SpendAlert) error {
	if alert == nil {
		return fmt.Errorf("openai organization spend alert response was empty")
	}
	if err := requireNonEmpty("organization spend alert", "id", alert.ID); err != nil {
		return err
	}
	if err := requireNonEmpty("organization spend alert", "currency", alert.Currency); err != nil {
		return err
	}
	return requireNonEmpty("organization spend alert", "interval", alert.Interval)
}

func mapCertificate(certificate *openai.Certificate) (*Certificate, error) {
	if certificate == nil {
		return nil, fmt.Errorf("openai organization certificate response was empty")
	}
	mapped := &Certificate{
		ID:        certificate.ID,
		Name:      certificate.Name,
		Object:    string(certificate.Object),
		Active:    certificate.Active,
		CreatedAt: certificate.CreatedAt,
		CertificateDetails: CertificateDetails{
			Content:   certificate.CertificateDetails.Content,
			ExpiresAt: certificate.CertificateDetails.ExpiresAt,
			ValidAt:   certificate.CertificateDetails.ValidAt,
		},
	}
	if err := validateCertificate(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapOrganizationCertificateListItem(certificate *openai.AdminOrganizationCertificateListResponse) (*Certificate, error) {
	if certificate == nil {
		return nil, fmt.Errorf("openai organization certificate response was empty")
	}
	mapped := &Certificate{
		ID:        certificate.ID,
		Name:      certificate.Name,
		Object:    string(certificate.Object),
		Active:    certificate.Active,
		CreatedAt: certificate.CreatedAt,
		CertificateDetails: CertificateDetails{
			ExpiresAt: certificate.CertificateDetails.ExpiresAt,
			ValidAt:   certificate.CertificateDetails.ValidAt,
		},
	}
	if err := validateCertificate(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapOrganizationCertificateActivateItem(certificate *openai.AdminOrganizationCertificateActivateResponse) (*Certificate, error) {
	if certificate == nil {
		return nil, fmt.Errorf("openai organization certificate activate response was empty")
	}
	mapped := &Certificate{
		ID:        certificate.ID,
		Name:      certificate.Name,
		Object:    string(certificate.Object),
		Active:    certificate.Active,
		CreatedAt: certificate.CreatedAt,
		CertificateDetails: CertificateDetails{
			ExpiresAt: certificate.CertificateDetails.ExpiresAt,
			ValidAt:   certificate.CertificateDetails.ValidAt,
		},
	}
	if err := validateCertificate(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapOrganizationCertificateDeactivateItem(certificate *openai.AdminOrganizationCertificateDeactivateResponse) (*Certificate, error) {
	if certificate == nil {
		return nil, fmt.Errorf("openai organization certificate deactivate response was empty")
	}
	mapped := &Certificate{
		ID:        certificate.ID,
		Name:      certificate.Name,
		Object:    string(certificate.Object),
		Active:    certificate.Active,
		CreatedAt: certificate.CreatedAt,
		CertificateDetails: CertificateDetails{
			ExpiresAt: certificate.CertificateDetails.ExpiresAt,
			ValidAt:   certificate.CertificateDetails.ValidAt,
		},
	}
	if err := validateCertificate(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateCertificate(certificate *Certificate) error {
	if certificate == nil {
		return fmt.Errorf("openai organization certificate response was empty")
	}
	return requireNonEmpty("organization certificate", "id", certificate.ID)
}

func mapOrganizationUser(user *openai.OrganizationUser) (*OrganizationUser, error) {
	if user == nil {
		return nil, fmt.Errorf("openai organization user response was empty")
	}
	mapped := &OrganizationUser{
		ID:                             user.ID,
		Name:                           user.Name,
		Email:                          user.Email,
		Role:                           user.Role,
		UserID:                         user.User.ID,
		UserName:                       user.User.Name,
		UserEmail:                      user.User.Email,
		Picture:                        user.User.Picture,
		DeveloperPersona:               user.DeveloperPersona,
		TechnicalLevel:                 user.TechnicalLevel,
		AddedAt:                        user.AddedAt,
		Created:                        user.Created,
		APIKeyLastUsedAt:               user.APIKeyLastUsedAt,
		BannedAt:                       user.User.BannedAt,
		IsDefault:                      user.IsDefault,
		IsScaleTierAuthorizedPurchaser: user.IsScaleTierAuthorizedPurchaser,
		IsScimManaged:                  user.IsScimManaged,
		IsServiceAccount:               user.IsServiceAccount,
		Banned:                         user.User.Banned,
		Enabled:                        user.User.Enabled,
		Projects:                       make([]OrganizationUserProject, 0, len(user.Projects.Data)),
	}
	for _, project := range user.Projects.Data {
		mapped.Projects = append(mapped.Projects, OrganizationUserProject{ID: project.ID, Name: project.Name, Role: project.Role})
	}
	if err := validateOrganizationUser(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateOrganizationUser(user *OrganizationUser) error {
	if user == nil {
		return fmt.Errorf("openai organization user response was empty")
	}
	return requireNonEmpty("organization user", "id", user.ID)
}

func mapOrganizationGroup(group *openai.Group) (*OrganizationGroup, error) {
	if group == nil {
		return nil, fmt.Errorf("openai organization group response was empty")
	}
	mapped := &OrganizationGroup{ID: group.ID, Name: group.Name, GroupType: string(group.GroupType), IsScimManaged: group.IsScimManaged, CreatedAt: group.CreatedAt}
	if err := validateOrganizationGroup(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateOrganizationGroup(group *OrganizationGroup) error {
	if group == nil {
		return fmt.Errorf("openai organization group response was empty")
	}
	if err := requireNonEmpty("organization group", "id", group.ID); err != nil {
		return err
	}
	return requireNonEmpty("organization group", "name", group.Name)
}

func validateOrganizationGroupUser(user *OrganizationGroupUser) error {
	if user == nil {
		return fmt.Errorf("openai organization group user response was empty")
	}
	return requireNonEmpty("organization group user", "id", user.ID)
}

func mapRole(role *openai.Role) (*Role, error) {
	if role == nil {
		return nil, fmt.Errorf("openai role response was empty")
	}
	mapped := &Role{
		ID:             role.ID,
		Name:           role.Name,
		Description:    role.Description,
		Permissions:    append([]string(nil), role.Permissions...),
		PredefinedRole: role.PredefinedRole,
		ResourceType:   role.ResourceType,
	}
	if err := validateRole(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateRole(role *Role) error {
	if role == nil {
		return fmt.Errorf("openai role response was empty")
	}
	if err := requireNonEmpty("role", "id", role.ID); err != nil {
		return err
	}
	return requireNonEmpty("role", "name", role.Name)
}

func mapRoleAssignmentFromRole(role *openai.Role, principalID, principalType string) (*RoleAssignment, error) {
	mappedRole, err := mapRole(role)
	if err != nil {
		return nil, err
	}
	return &RoleAssignment{Role: *mappedRole, PrincipalID: principalID, PrincipalType: principalType}, nil
}

func mapUserRoleAssignmentGet(role *openai.AdminOrganizationUserRoleGetResponse, userID string) (*RoleAssignment, error) {
	if role == nil {
		return nil, fmt.Errorf("openai organization user role response was empty")
	}
	mapped := &RoleAssignment{
		Role: Role{
			ID:             role.ID,
			Name:           role.Name,
			Description:    role.Description,
			Permissions:    append([]string(nil), role.Permissions...),
			PredefinedRole: role.PredefinedRole,
			ResourceType:   role.ResourceType,
		},
		PrincipalID:   userID,
		PrincipalType: "user",
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
		CreatedBy:     role.CreatedBy,
	}
	for _, source := range role.AssignmentSources {
		mapped.AssignmentSources = append(mapped.AssignmentSources, RoleAssignmentSource{PrincipalID: source.PrincipalID, PrincipalType: source.PrincipalType})
	}
	if err := validateRoleAssignment(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapUserRoleAssignmentList(role *openai.AdminOrganizationUserRoleListResponse, userID string) (*RoleAssignment, error) {
	if role == nil {
		return nil, fmt.Errorf("openai organization user role response was empty")
	}
	mapped := &RoleAssignment{
		Role: Role{
			ID:             role.ID,
			Name:           role.Name,
			Description:    role.Description,
			Permissions:    append([]string(nil), role.Permissions...),
			PredefinedRole: role.PredefinedRole,
			ResourceType:   role.ResourceType,
		},
		PrincipalID:   userID,
		PrincipalType: "user",
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
		CreatedBy:     role.CreatedBy,
	}
	for _, source := range role.AssignmentSources {
		mapped.AssignmentSources = append(mapped.AssignmentSources, RoleAssignmentSource{PrincipalID: source.PrincipalID, PrincipalType: source.PrincipalType})
	}
	if err := validateRoleAssignment(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapGroupRoleAssignmentGet(role *openai.AdminOrganizationGroupRoleGetResponse, groupID string) (*RoleAssignment, error) {
	if role == nil {
		return nil, fmt.Errorf("openai organization group role response was empty")
	}
	mapped := &RoleAssignment{
		Role: Role{
			ID:             role.ID,
			Name:           role.Name,
			Description:    role.Description,
			Permissions:    append([]string(nil), role.Permissions...),
			PredefinedRole: role.PredefinedRole,
			ResourceType:   role.ResourceType,
		},
		PrincipalID:   groupID,
		PrincipalType: "group",
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
		CreatedBy:     role.CreatedBy,
	}
	for _, source := range role.AssignmentSources {
		mapped.AssignmentSources = append(mapped.AssignmentSources, RoleAssignmentSource{PrincipalID: source.PrincipalID, PrincipalType: source.PrincipalType})
	}
	if err := validateRoleAssignment(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func mapGroupRoleAssignmentList(role *openai.AdminOrganizationGroupRoleListResponse, groupID string) (*RoleAssignment, error) {
	if role == nil {
		return nil, fmt.Errorf("openai organization group role response was empty")
	}
	mapped := &RoleAssignment{
		Role: Role{
			ID:             role.ID,
			Name:           role.Name,
			Description:    role.Description,
			Permissions:    append([]string(nil), role.Permissions...),
			PredefinedRole: role.PredefinedRole,
			ResourceType:   role.ResourceType,
		},
		PrincipalID:   groupID,
		PrincipalType: "group",
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
		CreatedBy:     role.CreatedBy,
	}
	for _, source := range role.AssignmentSources {
		mapped.AssignmentSources = append(mapped.AssignmentSources, RoleAssignmentSource{PrincipalID: source.PrincipalID, PrincipalType: source.PrincipalType})
	}
	if err := validateRoleAssignment(mapped); err != nil {
		return nil, err
	}
	return mapped, nil
}

func validateRoleAssignment(assignment *RoleAssignment) error {
	if assignment == nil {
		return fmt.Errorf("openai role assignment response was empty")
	}
	if err := validateRole(&assignment.Role); err != nil {
		return err
	}
	return requireNonEmpty("role assignment", "principal_id", assignment.PrincipalID)
}
