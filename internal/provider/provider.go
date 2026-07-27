package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/cloudopsworks/terraform-provider-openai/internal/secrets"
)

var _ provider.Provider = &openAIProvider{}

type openAIProvider struct {
	version string
}

type providerModel struct {
	AdminAPIKey       types.String `tfsdk:"admin_api_key"`
	BaseURL           types.String `tfsdk:"base_url"`
	OrganizationID    types.String `tfsdk:"organization_id"`
	ProjectID         types.String `tfsdk:"project_id"`
	AWSSecretsManager types.Object `tfsdk:"aws_secrets_manager"`
	GCPSecretManager  types.Object `tfsdk:"gcp_secret_manager"`
	AzureKeyVault     types.Object `tfsdk:"azure_key_vault"`
}

type awsSecretSourceModel struct {
	Region          types.String `tfsdk:"region"`
	SecretID        types.String `tfsdk:"secret_id"`
	VersionID       types.String `tfsdk:"version_id"`
	VersionStage    types.String `tfsdk:"version_stage"`
	JSONKey         types.String `tfsdk:"json_key"`
	RoleARN         types.String `tfsdk:"role_arn"`
	RoleSessionName types.String `tfsdk:"role_session_name"`
	ExternalID      types.String `tfsdk:"external_id"`
	DurationSeconds types.Int64  `tfsdk:"duration_seconds"`
}

type gcpSecretSourceModel struct {
	ProjectID                 types.String `tfsdk:"project_id"`
	SecretID                  types.String `tfsdk:"secret_id"`
	Version                   types.String `tfsdk:"version"`
	JSONKey                   types.String `tfsdk:"json_key"`
	ImpersonateServiceAccount types.String `tfsdk:"impersonate_service_account"`
	Delegates                 types.List   `tfsdk:"delegates"`
	Scopes                    types.List   `tfsdk:"scopes"`
}

type azureSecretSourceModel struct {
	VaultURL   types.String `tfsdk:"vault_url"`
	SecretName types.String `tfsdk:"secret_name"`
	Version    types.String `tfsdk:"version"`
	JSONKey    types.String `tfsdk:"json_key"`
}

type providerData struct {
	client client.AdminClient
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &openAIProvider{version: version}
	}
}

func (p *openAIProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openai"
	resp.Version = p.version
}

func (p *openAIProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		MarkdownDescription: "OpenAI Administration API Terraform provider. The provider uses an existing OpenAI admin API key supplied directly, from environment variables, or from exactly one supported cloud secret manager when no direct key is configured.",
		Attributes: map[string]providerschema.Attribute{
			"admin_api_key": providerschema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Existing OpenAI admin API key. Takes precedence over OPENAI_ADMIN_KEY and cloud secret-manager sources.",
			},
			"base_url": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional OpenAI API base URL override for testing or private gateways. Takes precedence over OPENAI_BASE_URL and secret payload base_url.",
			},
			"organization_id": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional OpenAI organization ID. Takes precedence over OPENAI_ORG_ID and secret payload organization_id.",
			},
			"project_id": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional OpenAI project ID header. Takes precedence over OPENAI_PROJECT_ID and secret payload project_id.",
			},
			"aws_secrets_manager": providerschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Read OpenAI provider settings from AWS Secrets Manager using the default AWS credential chain, optionally assuming an IAM role first. Ignored when admin_api_key or OPENAI_ADMIN_KEY is set; mutually exclusive with other cloud secret sources otherwise.",
				Attributes: map[string]providerschema.Attribute{
					"region":        providerschema.StringAttribute{Optional: true, MarkdownDescription: "AWS region. Optional when the AWS default config already provides a region."},
					"secret_id":     providerschema.StringAttribute{Required: true, MarkdownDescription: "Secrets Manager secret ID or ARN."},
					"version_id":    providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional secret version ID."},
					"version_stage": providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional secret version stage, such as AWSCURRENT."},
					"json_key":      providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional dot-separated JSON key path when the secret value is a JSON object."},
					"role_arn": providerschema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Optional AWS IAM role ARN to assume before reading the secret. When omitted, the AWS default credential chain is used directly.",
					},
					"role_session_name": providerschema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Optional AWS STS AssumeRole session name. Requires role_arn.",
					},
					"external_id": providerschema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Optional AWS STS AssumeRole external ID. Requires role_arn.",
					},
					"duration_seconds": providerschema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Optional AWS STS AssumeRole session duration in seconds. Requires role_arn.",
					},
				},
			},
			"gcp_secret_manager": providerschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Read OpenAI provider settings from Google Secret Manager using Application Default Credentials, optionally impersonating a service account first. Ignored when admin_api_key or OPENAI_ADMIN_KEY is set; mutually exclusive with other cloud secret sources otherwise.",
				Attributes: map[string]providerschema.Attribute{
					"project_id": providerschema.StringAttribute{Required: true, MarkdownDescription: "GCP project ID that owns the secret."},
					"secret_id":  providerschema.StringAttribute{Required: true, MarkdownDescription: "Secret ID."},
					"version":    providerschema.StringAttribute{Optional: true, MarkdownDescription: "Secret version. Defaults to latest."},
					"json_key":   providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional dot-separated JSON key path when the secret value is a JSON object."},
					"impersonate_service_account": providerschema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Optional target service account email to impersonate when reading the secret. Uses Application Default Credentials as the base credential.",
					},
					"delegates": providerschema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Optional service account delegation chain used when impersonate_service_account is configured.",
					},
					"scopes": providerschema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Optional OAuth scopes used for Google Secret Manager authentication. Defaults to the Secret Manager client scopes when omitted.",
					},
				},
			},
			"azure_key_vault": providerschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Read the OpenAI admin API key from Azure Key Vault using DefaultAzureCredential. Mutually exclusive with all other key sources.",
				Attributes: map[string]providerschema.Attribute{
					"vault_url":   providerschema.StringAttribute{Required: true, MarkdownDescription: "Azure Key Vault URL, for example https://example.vault.azure.net/."},
					"secret_name": providerschema.StringAttribute{Required: true, MarkdownDescription: "Key Vault secret name."},
					"version":     providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional secret version."},
					"json_key":    providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional dot-separated JSON key path when the secret value is a JSON object."},
				},
			},
		},
	}
}

func (p *openAIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretConfig, diags := sourceConfigFromProviderModelWithEnv(ctx, config, os.LookupEnv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolver, source, err := secrets.NewResolver(secretConfig)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("admin_api_key"), "Invalid OpenAI admin API key source", err.Error())
		return
	}

	secretSettings, err := resolver.Resolve(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to resolve OpenAI admin API key", err.Error())
		return
	}
	settings := secretConfig.Direct.MergeFallback(secretSettings)
	if settings.AdminAPIKey == "" {
		resp.Diagnostics.AddError("OpenAI admin API key is empty", "The configured OpenAI admin API key source returned an empty value.")
		return
	}

	traceCtx := tflog.SetField(ctx, "provider", "openai")
	traceCtx = tflog.SetField(traceCtx, "admin_api_key_source", source)
	traceCtx = tflog.SetField(traceCtx, "base_url_configured", settings.BaseURL != "")
	traceCtx = tflog.SetField(traceCtx, "organization_id_configured", settings.OrganizationID != "")
	traceCtx = tflog.SetField(traceCtx, "project_id_configured", settings.ProjectID != "")
	tflog.Trace(traceCtx, "configuring OpenAI provider")

	userAgent := fmt.Sprintf("terraform-provider-openai/%s", p.version)
	data := &providerData{client: client.NewWithSettings(client.Settings{
		AdminAPIKey:    settings.AdminAPIKey,
		BaseURL:        settings.BaseURL,
		OrganizationID: settings.OrganizationID,
		ProjectID:      settings.ProjectID,
	}, userAgent, 30*time.Second)}
	resp.DataSourceData = data
	resp.ResourceData = data
}

func (p *openAIProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewServiceAccountResource,
		NewProjectAPIKeyResource,
		NewAdminAPIKeyResource,
		NewOrganizationGroupResource,
		NewOrganizationGroupUserResource,
		NewOrganizationRoleResource,
		NewOrganizationUserRoleResource,
		NewOrganizationGroupRoleResource,
		NewProjectRoleResource,
	}
}

func (p *openAIProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewProjectsDataSource,
		NewServiceAccountDataSource,
		NewServiceAccountsDataSource,
		NewAdminAPIKeyDataSource,
		NewAdminAPIKeysDataSource,
		NewOrganizationUserDataSource,
		NewOrganizationUsersDataSource,
		NewOrganizationGroupDataSource,
		NewOrganizationGroupsDataSource,
		NewOrganizationGroupUserDataSource,
		NewOrganizationGroupUsersDataSource,
		NewOrganizationRoleDataSource,
		NewOrganizationRolesDataSource,
		NewOrganizationUserRoleDataSource,
		NewOrganizationUserRolesDataSource,
		NewOrganizationGroupRoleDataSource,
		NewOrganizationGroupRolesDataSource,
		NewProjectRoleDataSource,
		NewProjectRolesDataSource,
	}
}

func sourceConfigFromProviderModel(ctx context.Context, model providerModel) (secrets.SourceConfig, diag.Diagnostics) {
	return sourceConfigFromProviderModelWithEnv(ctx, model, nil)
}

func sourceConfigFromProviderModelWithEnv(ctx context.Context, model providerModel, lookupEnv func(string) (string, bool)) (secrets.SourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	direct := providerSettingsFromModel(model)
	if lookupEnv != nil {
		direct = providerSettingsFromModelWithEnv(model, lookupEnv)
	}
	cfg := secrets.SourceConfig{Direct: direct}
	if cfg.Direct.AdminAPIKey != "" {
		return cfg, diags
	}
	if !model.AWSSecretsManager.IsNull() && !model.AWSSecretsManager.IsUnknown() {
		var awsModel awsSecretSourceModel
		diags.Append(model.AWSSecretsManager.As(ctx, &awsModel, basetypes.ObjectAsOptions{})...)
		awsPath := path.Root("aws_secrets_manager")
		roleARN := strings.TrimSpace(stringValue(awsModel.RoleARN))
		roleSessionName := strings.TrimSpace(stringValue(awsModel.RoleSessionName))
		externalID := strings.TrimSpace(stringValue(awsModel.ExternalID))
		durationSeconds := int64Value(awsModel.DurationSeconds)
		if !awsModel.DurationSeconds.IsNull() && !awsModel.DurationSeconds.IsUnknown() && durationSeconds <= 0 {
			diags.AddAttributeError(awsPath.AtName("duration_seconds"), "Invalid AWS assume role duration", "duration_seconds must be positive when provided.")
		}
		if roleARN == "" {
			if roleSessionName != "" {
				diags.AddAttributeError(awsPath.AtName("role_session_name"), "Invalid AWS assume role configuration", "role_session_name requires role_arn to be configured.")
			}
			if externalID != "" {
				diags.AddAttributeError(awsPath.AtName("external_id"), "Invalid AWS assume role configuration", "external_id requires role_arn to be configured.")
			}
			if !awsModel.DurationSeconds.IsNull() && !awsModel.DurationSeconds.IsUnknown() {
				diags.AddAttributeError(awsPath.AtName("duration_seconds"), "Invalid AWS assume role configuration", "duration_seconds requires role_arn to be configured.")
			}
		}
		cfg.AWS = &secrets.AWSConfig{
			Region:          stringValue(awsModel.Region),
			SecretID:        stringValue(awsModel.SecretID),
			VersionID:       stringValue(awsModel.VersionID),
			VersionStage:    stringValue(awsModel.VersionStage),
			JSONKey:         stringValue(awsModel.JSONKey),
			RoleARN:         roleARN,
			RoleSessionName: roleSessionName,
			ExternalID:      externalID,
			DurationSeconds: durationSeconds,
		}
	}
	if !model.GCPSecretManager.IsNull() && !model.GCPSecretManager.IsUnknown() {
		var gcpModel gcpSecretSourceModel
		diags.Append(model.GCPSecretManager.As(ctx, &gcpModel, basetypes.ObjectAsOptions{})...)
		gcpPath := path.Root("gcp_secret_manager")
		impersonateServiceAccount := strings.TrimSpace(stringValue(gcpModel.ImpersonateServiceAccount))
		delegates, delegateDiags := trimmedStringList(ctx, gcpModel.Delegates, gcpPath.AtName("delegates"), "Invalid GCP impersonation delegates", "delegates cannot contain empty or whitespace-only values.")
		diags.Append(delegateDiags...)
		scopes, scopeDiags := trimmedStringList(ctx, gcpModel.Scopes, gcpPath.AtName("scopes"), "Invalid GCP OAuth scopes", "scopes cannot contain empty or whitespace-only values.")
		diags.Append(scopeDiags...)
		if impersonateServiceAccount == "" && len(delegates) > 0 {
			diags.AddAttributeError(gcpPath.AtName("delegates"), "Invalid GCP impersonation delegates", "delegates requires impersonate_service_account to be configured.")
		}
		cfg.GCP = &secrets.GCPConfig{
			ProjectID:                 stringValue(gcpModel.ProjectID),
			SecretID:                  stringValue(gcpModel.SecretID),
			Version:                   stringValue(gcpModel.Version),
			JSONKey:                   stringValue(gcpModel.JSONKey),
			ImpersonateServiceAccount: impersonateServiceAccount,
			Delegates:                 delegates,
			Scopes:                    scopes,
		}
	}
	if !model.AzureKeyVault.IsNull() && !model.AzureKeyVault.IsUnknown() {
		var azureModel azureSecretSourceModel
		diags.Append(model.AzureKeyVault.As(ctx, &azureModel, basetypes.ObjectAsOptions{})...)
		cfg.Azure = &secrets.AzureConfig{
			VaultURL:   stringValue(azureModel.VaultURL),
			SecretName: stringValue(azureModel.SecretName),
			Version:    stringValue(azureModel.Version),
			JSONKey:    stringValue(azureModel.JSONKey),
		}
	}
	if lookupEnv != nil && cfg.AWS == nil && cfg.GCP == nil && cfg.Azure == nil {
		envDiags := applySecretSourceEnv(&cfg, lookupEnv)
		diags.Append(envDiags...)
	}
	return cfg, diags
}

func providerSettingsFromModel(model providerModel) secrets.Settings {
	return secrets.Settings{
		AdminAPIKey:    stringValue(model.AdminAPIKey),
		BaseURL:        stringValue(model.BaseURL),
		OrganizationID: stringValue(model.OrganizationID),
		ProjectID:      stringValue(model.ProjectID),
	}.Trimmed()
}

func providerSettingsFromModelWithEnv(model providerModel, lookupEnv func(string) (string, bool)) secrets.Settings {
	settings := providerSettingsFromModel(model)
	settings = settings.MergeFallback(secrets.Settings{
		AdminAPIKey:    envString(lookupEnv, "OPENAI_ADMIN_KEY", "OPENAI_ADMIN_API_KEY"),
		BaseURL:        envString(lookupEnv, "OPENAI_BASE_URL"),
		OrganizationID: envString(lookupEnv, "OPENAI_ORG_ID", "OPENAI_ORGANIZATION_ID"),
		ProjectID:      envString(lookupEnv, "OPENAI_PROJECT_ID"),
	})
	return settings.Trimmed()
}

func applySecretSourceEnv(cfg *secrets.SourceConfig, lookupEnv func(string) (string, bool)) diag.Diagnostics {
	var diags diag.Diagnostics
	awsCfg := awsConfigFromEnv(lookupEnv, &diags)
	gcpCfg := gcpConfigFromEnv(lookupEnv)
	if awsCfg != nil {
		cfg.AWS = awsCfg
	}
	if gcpCfg != nil {
		if strings.TrimSpace(gcpCfg.ImpersonateServiceAccount) == "" && len(gcpCfg.Delegates) > 0 {
			diags.AddAttributeError(path.Root("gcp_secret_manager").AtName("delegates"), "Invalid GCP impersonation delegates", "OPENAI_GCP_SECRET_MANAGER_DELEGATES requires OPENAI_GCP_SECRET_MANAGER_IMPERSONATE_SERVICE_ACCOUNT to be configured.")
		}
		cfg.GCP = gcpCfg
	}
	return diags
}

func awsConfigFromEnv(lookupEnv func(string) (string, bool), diags *diag.Diagnostics) *secrets.AWSConfig {
	secretID := envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_SECRET_ID", "OPENAI_AWS_SECRET_ID")
	if secretID == "" {
		return nil
	}
	cfg := &secrets.AWSConfig{
		Region:          envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_REGION", "AWS_REGION", "AWS_DEFAULT_REGION"),
		SecretID:        secretID,
		VersionID:       envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_VERSION_ID"),
		VersionStage:    envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_VERSION_STAGE"),
		JSONKey:         envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_JSON_KEY"),
		RoleARN:         envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN", "OPENAI_AWS_ROLE_ARN"),
		RoleSessionName: envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_ROLE_SESSION_NAME", "OPENAI_AWS_ROLE_SESSION_NAME"),
		ExternalID:      envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_EXTERNAL_ID", "OPENAI_AWS_EXTERNAL_ID"),
	}
	if raw := envString(lookupEnv, "OPENAI_AWS_SECRETS_MANAGER_DURATION_SECONDS", "OPENAI_AWS_ROLE_DURATION_SECONDS"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			diags.AddAttributeError(path.Root("aws_secrets_manager").AtName("duration_seconds"), "Invalid AWS assume role duration", "OPENAI_AWS_SECRETS_MANAGER_DURATION_SECONDS must be a positive integer when provided.")
		} else {
			cfg.DurationSeconds = value
		}
	}
	if cfg.RoleARN == "" {
		if cfg.RoleSessionName != "" {
			diags.AddAttributeError(path.Root("aws_secrets_manager").AtName("role_session_name"), "Invalid AWS assume role configuration", "OPENAI_AWS_SECRETS_MANAGER_ROLE_SESSION_NAME requires OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN to be configured.")
		}
		if cfg.ExternalID != "" {
			diags.AddAttributeError(path.Root("aws_secrets_manager").AtName("external_id"), "Invalid AWS assume role configuration", "OPENAI_AWS_SECRETS_MANAGER_EXTERNAL_ID requires OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN to be configured.")
		}
		if cfg.DurationSeconds > 0 {
			diags.AddAttributeError(path.Root("aws_secrets_manager").AtName("duration_seconds"), "Invalid AWS assume role configuration", "OPENAI_AWS_SECRETS_MANAGER_DURATION_SECONDS requires OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN to be configured.")
		}
	}
	return cfg
}

func gcpConfigFromEnv(lookupEnv func(string) (string, bool)) *secrets.GCPConfig {
	projectID := envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_PROJECT_ID", "GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT")
	secretID := envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_SECRET_ID", "OPENAI_GCP_SECRET_ID")
	if projectID == "" || secretID == "" {
		return nil
	}
	return &secrets.GCPConfig{
		ProjectID:                 projectID,
		SecretID:                  secretID,
		Version:                   envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_VERSION"),
		JSONKey:                   envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_JSON_KEY"),
		ImpersonateServiceAccount: envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_IMPERSONATE_SERVICE_ACCOUNT", "OPENAI_GCP_IMPERSONATE_SERVICE_ACCOUNT"),
		Delegates:                 splitEnvList(envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_DELEGATES")),
		Scopes:                    splitEnvList(envString(lookupEnv, "OPENAI_GCP_SECRET_MANAGER_SCOPES")),
	}
}

func envString(lookupEnv func(string) (string, bool), names ...string) string {
	if lookupEnv == nil {
		return ""
	}
	for _, name := range names {
		value, ok := lookupEnv(name)
		if ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func splitEnvList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func stringValue(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}

func trimmedStringList(ctx context.Context, value types.List, attrPath path.Path, summary, detail string) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	var raw []string
	diags.Append(value.ElementsAs(ctx, &raw, false)...)
	if diags.HasError() {
		return nil, diags
	}
	values := make([]string, 0, len(raw))
	for i, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			diags.AddAttributeError(attrPath.AtListIndex(i), summary, detail)
			continue
		}
		values = append(values, trimmed)
	}
	return values, diags
}
