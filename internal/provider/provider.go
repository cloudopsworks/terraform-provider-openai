package provider

import (
	"context"
	"fmt"
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
	AWSSecretsManager types.Object `tfsdk:"aws_secrets_manager"`
	GCPSecretManager  types.Object `tfsdk:"gcp_secret_manager"`
	AzureKeyVault     types.Object `tfsdk:"azure_key_vault"`
}

type awsSecretSourceModel struct {
	Region       types.String `tfsdk:"region"`
	SecretID     types.String `tfsdk:"secret_id"`
	VersionID    types.String `tfsdk:"version_id"`
	VersionStage types.String `tfsdk:"version_stage"`
	JSONKey      types.String `tfsdk:"json_key"`
}

type gcpSecretSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	SecretID  types.String `tfsdk:"secret_id"`
	Version   types.String `tfsdk:"version"`
	JSONKey   types.String `tfsdk:"json_key"`
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
		MarkdownDescription: "OpenAI Administration API Terraform provider. The provider uses an existing OpenAI admin API key, supplied directly or read from exactly one supported cloud secret manager.",
		Attributes: map[string]providerschema.Attribute{
			"admin_api_key": providerschema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Existing OpenAI admin API key. Mutually exclusive with aws_secrets_manager, gcp_secret_manager, and azure_key_vault.",
			},
			"base_url": providerschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional OpenAI API base URL override for testing or private gateways. Defaults to the OpenAI SDK default.",
			},
			"aws_secrets_manager": providerschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Read the OpenAI admin API key from AWS Secrets Manager using the default AWS credential chain. Mutually exclusive with all other key sources.",
				Attributes: map[string]providerschema.Attribute{
					"region":        providerschema.StringAttribute{Optional: true, MarkdownDescription: "AWS region. Optional when the AWS default config already provides a region."},
					"secret_id":     providerschema.StringAttribute{Required: true, MarkdownDescription: "Secrets Manager secret ID or ARN."},
					"version_id":    providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional secret version ID."},
					"version_stage": providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional secret version stage, such as AWSCURRENT."},
					"json_key":      providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional dot-separated JSON key path when the secret value is a JSON object."},
				},
			},
			"gcp_secret_manager": providerschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Read the OpenAI admin API key from Google Secret Manager using Application Default Credentials. Mutually exclusive with all other key sources.",
				Attributes: map[string]providerschema.Attribute{
					"project_id": providerschema.StringAttribute{Required: true, MarkdownDescription: "GCP project ID that owns the secret."},
					"secret_id":  providerschema.StringAttribute{Required: true, MarkdownDescription: "Secret ID."},
					"version":    providerschema.StringAttribute{Optional: true, MarkdownDescription: "Secret version. Defaults to latest."},
					"json_key":   providerschema.StringAttribute{Optional: true, MarkdownDescription: "Optional dot-separated JSON key path when the secret value is a JSON object."},
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

	secretConfig, diags := sourceConfigFromProviderModel(ctx, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolver, source, err := secrets.NewResolver(secretConfig)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("admin_api_key"), "Invalid OpenAI admin API key source", err.Error())
		return
	}

	adminAPIKey, err := resolver.Resolve(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to resolve OpenAI admin API key", err.Error())
		return
	}
	if adminAPIKey == "" {
		resp.Diagnostics.AddError("OpenAI admin API key is empty", "The configured OpenAI admin API key source returned an empty value.")
		return
	}

	baseURL := ""
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}
	traceCtx := tflog.SetField(ctx, "provider", "openai")
	traceCtx = tflog.SetField(traceCtx, "admin_api_key_source", source)
	traceCtx = tflog.SetField(traceCtx, "base_url_configured", baseURL != "")
	tflog.Trace(traceCtx, "configuring OpenAI provider")

	userAgent := fmt.Sprintf("terraform-provider-openai/%s", p.version)
	data := &providerData{client: client.New(adminAPIKey, baseURL, userAgent, 30*time.Second)}
	resp.DataSourceData = data
	resp.ResourceData = data
}

func (p *openAIProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewServiceAccountResource,
		NewProjectAPIKeyResource,
	}
}

func (p *openAIProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func sourceConfigFromProviderModel(ctx context.Context, model providerModel) (secrets.SourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := secrets.SourceConfig{Direct: stringValue(model.AdminAPIKey)}
	if !model.AWSSecretsManager.IsNull() && !model.AWSSecretsManager.IsUnknown() {
		var awsModel awsSecretSourceModel
		diags.Append(model.AWSSecretsManager.As(ctx, &awsModel, basetypes.ObjectAsOptions{})...)
		cfg.AWS = &secrets.AWSConfig{
			Region:       stringValue(awsModel.Region),
			SecretID:     stringValue(awsModel.SecretID),
			VersionID:    stringValue(awsModel.VersionID),
			VersionStage: stringValue(awsModel.VersionStage),
			JSONKey:      stringValue(awsModel.JSONKey),
		}
	}
	if !model.GCPSecretManager.IsNull() && !model.GCPSecretManager.IsUnknown() {
		var gcpModel gcpSecretSourceModel
		diags.Append(model.GCPSecretManager.As(ctx, &gcpModel, basetypes.ObjectAsOptions{})...)
		cfg.GCP = &secrets.GCPConfig{
			ProjectID: stringValue(gcpModel.ProjectID),
			SecretID:  stringValue(gcpModel.SecretID),
			Version:   stringValue(gcpModel.Version),
			JSONKey:   stringValue(gcpModel.JSONKey),
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
	return cfg, diags
}

func stringValue(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}
