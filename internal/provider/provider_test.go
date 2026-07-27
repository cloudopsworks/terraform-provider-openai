package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/cloudopsworks/terraform-provider-openai/internal/secrets"
)

func TestStringValue(t *testing.T) {
	if got := stringValue(types.StringNull()); got != "" {
		t.Fatalf("null stringValue = %q", got)
	}
	if got := stringValue(types.StringValue("x")); got != "x" {
		t.Fatalf("stringValue = %q", got)
	}
}

func TestSourceConfigFromProviderModelDirect(t *testing.T) {
	cfg, diags := sourceConfigFromProviderModel(context.Background(), providerModel{
		AdminAPIKey:    types.StringValue("sk"),
		BaseURL:        types.StringValue("https://api.example/v1"),
		OrganizationID: types.StringValue("org_1"),
		ProjectID:      types.StringValue("proj_1"),
	})
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if cfg.Direct.AdminAPIKey != "sk" || cfg.Direct.BaseURL != "https://api.example/v1" || cfg.Direct.OrganizationID != "org_1" || cfg.Direct.ProjectID != "proj_1" || cfg.AWS != nil || cfg.GCP != nil || cfg.Azure != nil {
		t.Fatalf("unexpected source config: %#v", cfg)
	}
}

func TestProviderSettingsFromModelWithEnvPrecedence(t *testing.T) {
	settings := providerSettingsFromModelWithEnv(providerModel{
		AdminAPIKey:    types.StringValue(" sk-config "),
		BaseURL:        types.StringNull(),
		OrganizationID: types.StringValue(" org-config "),
		ProjectID:      types.StringNull(),
	}, envLookup(map[string]string{
		"OPENAI_ADMIN_KEY":  "sk-env",
		"OPENAI_BASE_URL":   "https://api.env/v1",
		"OPENAI_ORG_ID":     "org-env",
		"OPENAI_PROJECT_ID": "proj-env",
	}))
	if settings.AdminAPIKey != "sk-config" || settings.BaseURL != "https://api.env/v1" || settings.OrganizationID != "org-config" || settings.ProjectID != "proj-env" {
		t.Fatalf("settings = %#v", settings)
	}
}

func TestSourceConfigFromProviderModelEnvDiscoveryAWS(t *testing.T) {
	ctx := context.Background()
	cfg, diags := sourceConfigFromProviderModelWithEnv(ctx, emptyProviderModel(), envLookup(map[string]string{
		"OPENAI_AWS_SECRETS_MANAGER_REGION":            " us-east-1 ",
		"OPENAI_AWS_SECRETS_MANAGER_SECRET_ID":         " prod/openai/admin ",
		"OPENAI_AWS_SECRETS_MANAGER_JSON_KEY":          " admin_api_key ",
		"OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN":          " arn:aws:iam::123456789012:role/openai-secret-reader ",
		"OPENAI_AWS_SECRETS_MANAGER_ROLE_SESSION_NAME": " terraform-provider-openai ",
		"OPENAI_AWS_SECRETS_MANAGER_EXTERNAL_ID":       " external-1 ",
		"OPENAI_AWS_SECRETS_MANAGER_DURATION_SECONDS":  "900",
	}))
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if cfg.AWS == nil || cfg.AWS.Region != "us-east-1" || cfg.AWS.SecretID != "prod/openai/admin" || cfg.AWS.JSONKey != "admin_api_key" || cfg.AWS.RoleARN == "" || cfg.AWS.DurationSeconds != 900 {
		t.Fatalf("AWS env config = %#v", cfg.AWS)
	}
}

func TestSourceConfigFromProviderModelEnvDiscoveryGCP(t *testing.T) {
	ctx := context.Background()
	cfg, diags := sourceConfigFromProviderModelWithEnv(ctx, emptyProviderModel(), envLookup(map[string]string{
		"OPENAI_GCP_SECRET_MANAGER_PROJECT_ID":                  " platform-prod ",
		"OPENAI_GCP_SECRET_MANAGER_SECRET_ID":                   " openai-admin-key ",
		"OPENAI_GCP_SECRET_MANAGER_VERSION":                     " latest ",
		"OPENAI_GCP_SECRET_MANAGER_JSON_KEY":                    " admin_api_key ",
		"OPENAI_GCP_SECRET_MANAGER_IMPERSONATE_SERVICE_ACCOUNT": " secret-reader@platform-prod.iam.gserviceaccount.com ",
		"OPENAI_GCP_SECRET_MANAGER_DELEGATES":                   " delegate-1@platform-prod.iam.gserviceaccount.com, delegate-2@platform-prod.iam.gserviceaccount.com ",
		"OPENAI_GCP_SECRET_MANAGER_SCOPES":                      " https://www.googleapis.com/auth/secretmanager, https://www.googleapis.com/auth/cloud-platform ",
	}))
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if cfg.GCP == nil || cfg.GCP.ProjectID != "platform-prod" || cfg.GCP.SecretID != "openai-admin-key" || cfg.GCP.Version != "latest" || cfg.GCP.ImpersonateServiceAccount == "" {
		t.Fatalf("GCP env config = %#v", cfg.GCP)
	}
	if len(cfg.GCP.Delegates) != 2 || len(cfg.GCP.Scopes) != 2 {
		t.Fatalf("GCP env list config = %#v", cfg.GCP)
	}
}

func TestSourceConfigFromProviderModelEnvDirectSuppressesSecretDiscovery(t *testing.T) {
	ctx := context.Background()
	cfg, diags := sourceConfigFromProviderModelWithEnv(ctx, emptyProviderModel(), envLookup(map[string]string{
		"OPENAI_ADMIN_KEY":                     " sk-env ",
		"OPENAI_AWS_SECRETS_MANAGER_SECRET_ID": " prod/openai/admin ",
	}))
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if cfg.Direct.AdminAPIKey != "sk-env" || cfg.AWS != nil || cfg.GCP != nil {
		t.Fatalf("unexpected source config: %#v", cfg)
	}
}

func TestSourceConfigFromProviderModelDirectSuppressesExplicitSecretValidation(t *testing.T) {
	ctx := context.Background()
	awsObject, diags := types.ObjectValue(awsObjectTypes(), map[string]attr.Value{
		"region":            types.StringValue("us-east-1"),
		"secret_id":         types.StringValue("openai-admin-key"),
		"version_id":        types.StringNull(),
		"version_stage":     types.StringNull(),
		"json_key":          types.StringNull(),
		"role_arn":          types.StringNull(),
		"role_session_name": types.StringValue("would-be-invalid-without-role"),
		"external_id":       types.StringNull(),
		"duration_seconds":  types.Int64Null(),
	})
	if diags.HasError() {
		t.Fatalf("object value diagnostics: %v", diags)
	}
	cfg, diags := sourceConfigFromProviderModel(ctx, providerModel{
		AdminAPIKey:       types.StringValue("sk-direct"),
		BaseURL:           types.StringNull(),
		OrganizationID:    types.StringNull(),
		ProjectID:         types.StringNull(),
		AWSSecretsManager: awsObject,
		GCPSecretManager:  objectNull(gcpObjectTypes()),
		AzureKeyVault:     objectNull(azureObjectTypes()),
	})
	if diags.HasError() {
		t.Fatalf("direct configuration should suppress secret validation diagnostics: %v", diags)
	}
	if cfg.Direct.AdminAPIKey != "sk-direct" || cfg.AWS != nil {
		t.Fatalf("unexpected source config: %#v", cfg)
	}
}

func TestSourceConfigFromProviderModelEnvCloudConflict(t *testing.T) {
	ctx := context.Background()
	cfg, diags := sourceConfigFromProviderModelWithEnv(ctx, emptyProviderModel(), envLookup(map[string]string{
		"OPENAI_AWS_SECRETS_MANAGER_SECRET_ID": " prod/openai/admin ",
		"OPENAI_GCP_SECRET_MANAGER_PROJECT_ID": " platform-prod ",
		"OPENAI_GCP_SECRET_MANAGER_SECRET_ID":  " openai-admin-key ",
	}))
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if _, _, err := secrets.NewResolver(cfg); err == nil {
		t.Fatal("expected cloud source conflict")
	}
}

func TestSourceConfigFromProviderModelAWSSecretManagerAssumeRole(t *testing.T) {
	ctx := context.Background()
	awsObject, diags := types.ObjectValue(awsObjectTypes(), map[string]attr.Value{
		"region":            types.StringValue("us-east-1"),
		"secret_id":         types.StringValue("openai-admin-key"),
		"version_id":        types.StringNull(),
		"version_stage":     types.StringNull(),
		"json_key":          types.StringValue("admin_api_key"),
		"role_arn":          types.StringValue(" \n arn:aws:iam::123456789012:role/openai-secret-reader \t "),
		"role_session_name": types.StringValue(" terraform-provider-openai "),
		"external_id":       types.StringValue("\texternal-1\n"),
		"duration_seconds":  types.Int64Value(900),
	})
	if diags.HasError() {
		t.Fatalf("object value diagnostics: %v", diags)
	}

	cfg, diags := sourceConfigFromProviderModel(ctx, providerModel{
		AdminAPIKey:       types.StringNull(),
		AWSSecretsManager: awsObject,
		GCPSecretManager:  objectNull(gcpObjectTypes()),
		AzureKeyVault:     objectNull(azureObjectTypes()),
	})
	if diags.HasError() {
		t.Fatalf("source config diagnostics: %v", diags)
	}
	if cfg.AWS == nil {
		t.Fatal("expected AWS config")
	}
	if cfg.AWS.RoleARN != "arn:aws:iam::123456789012:role/openai-secret-reader" {
		t.Fatalf("RoleARN = %q", cfg.AWS.RoleARN)
	}
	if cfg.AWS.RoleSessionName != "terraform-provider-openai" {
		t.Fatalf("RoleSessionName = %q", cfg.AWS.RoleSessionName)
	}
	if cfg.AWS.ExternalID != "external-1" {
		t.Fatalf("ExternalID = %q", cfg.AWS.ExternalID)
	}
	if cfg.AWS.DurationSeconds != 900 {
		t.Fatalf("DurationSeconds = %d", cfg.AWS.DurationSeconds)
	}
}

func TestSourceConfigFromProviderModelAWSSecretManagerRejectsAssumeRoleFieldsWithoutRoleARN(t *testing.T) {
	ctx := context.Background()
	tests := map[string]struct {
		values       map[string]attr.Value
		expectedPath path.Path
	}{
		"role_session_name": {
			values:       map[string]attr.Value{"role_session_name": types.StringValue("terraform-provider-openai")},
			expectedPath: path.Root("aws_secrets_manager").AtName("role_session_name"),
		},
		"external_id": {
			values:       map[string]attr.Value{"external_id": types.StringValue("external-1")},
			expectedPath: path.Root("aws_secrets_manager").AtName("external_id"),
		},
		"duration_seconds": {
			values:       map[string]attr.Value{"duration_seconds": types.Int64Value(900)},
			expectedPath: path.Root("aws_secrets_manager").AtName("duration_seconds"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			values := map[string]attr.Value{
				"region":            types.StringValue("us-east-1"),
				"secret_id":         types.StringValue("openai-admin-key"),
				"version_id":        types.StringNull(),
				"version_stage":     types.StringNull(),
				"json_key":          types.StringNull(),
				"role_arn":          types.StringValue(" \t "),
				"role_session_name": types.StringNull(),
				"external_id":       types.StringNull(),
				"duration_seconds":  types.Int64Null(),
			}
			for key, value := range test.values {
				values[key] = value
			}
			awsObject, diags := types.ObjectValue(awsObjectTypes(), values)
			if diags.HasError() {
				t.Fatalf("object value diagnostics: %v", diags)
			}

			_, diags = sourceConfigFromProviderModel(ctx, providerModel{
				AdminAPIKey:       types.StringNull(),
				AWSSecretsManager: awsObject,
				GCPSecretManager:  objectNull(gcpObjectTypes()),
				AzureKeyVault:     objectNull(azureObjectTypes()),
			})
			if !diags.HasError() {
				t.Fatal("expected AWS assume role dependency diagnostic")
			}
			if !diagnosticsContainPath(diags, test.expectedPath) {
				t.Fatalf("expected diagnostic path %s, got %#v", test.expectedPath, diags)
			}
		})
	}
}

func TestSourceConfigFromProviderModelAWSSecretManagerRejectsNonPositiveDurationSeconds(t *testing.T) {
	ctx := context.Background()
	tests := map[string]int64{
		"zero":     0,
		"negative": -1,
	}

	for name, durationSeconds := range tests {
		t.Run(name, func(t *testing.T) {
			awsObject, diags := types.ObjectValue(awsObjectTypes(), map[string]attr.Value{
				"region":            types.StringValue("us-east-1"),
				"secret_id":         types.StringValue("openai-admin-key"),
				"version_id":        types.StringNull(),
				"version_stage":     types.StringNull(),
				"json_key":          types.StringNull(),
				"role_arn":          types.StringValue("arn:aws:iam::123456789012:role/openai-secret-reader"),
				"role_session_name": types.StringNull(),
				"external_id":       types.StringNull(),
				"duration_seconds":  types.Int64Value(durationSeconds),
			})
			if diags.HasError() {
				t.Fatalf("object value diagnostics: %v", diags)
			}

			_, diags = sourceConfigFromProviderModel(ctx, providerModel{
				AdminAPIKey:       types.StringNull(),
				AWSSecretsManager: awsObject,
				GCPSecretManager:  objectNull(gcpObjectTypes()),
				AzureKeyVault:     objectNull(azureObjectTypes()),
			})
			if !diags.HasError() {
				t.Fatal("expected duration_seconds diagnostic")
			}
			expectedPath := path.Root("aws_secrets_manager").AtName("duration_seconds")
			if !diagnosticsContainPath(diags, expectedPath) {
				t.Fatalf("expected diagnostic path %s, got %#v", expectedPath, diags)
			}
		})
	}
}

func TestSourceConfigFromProviderModelGCPSecretManagerImpersonation(t *testing.T) {
	ctx := context.Background()
	gcpObject, diags := types.ObjectValue(gcpObjectTypes(), map[string]attr.Value{
		"project_id":                  types.StringValue("project-1"),
		"secret_id":                   types.StringValue("openai-admin-key"),
		"version":                     types.StringValue("latest"),
		"json_key":                    types.StringValue("admin_api_key"),
		"impersonate_service_account": types.StringValue(" target-sa@example.iam.gserviceaccount.com "),
		"delegates":                   types.ListValueMust(types.StringType, []attr.Value{types.StringValue(" delegate-1@example.iam.gserviceaccount.com "), types.StringValue("\ndelegate-2@example.iam.gserviceaccount.com\t")}),
		"scopes":                      types.ListValueMust(types.StringType, []attr.Value{types.StringValue(" https://www.googleapis.com/auth/secretmanager "), types.StringValue("\thttps://www.googleapis.com/auth/cloud-platform\n")}),
	})
	if diags.HasError() {
		t.Fatalf("object value diagnostics: %v", diags)
	}

	cfg, diags := sourceConfigFromProviderModel(ctx, providerModel{
		AdminAPIKey:       types.StringNull(),
		AWSSecretsManager: objectNull(awsObjectTypes()),
		GCPSecretManager:  gcpObject,
		AzureKeyVault:     objectNull(azureObjectTypes()),
	})
	if diags.HasError() {
		t.Fatalf("source config diagnostics: %v", diags)
	}
	if cfg.GCP == nil {
		t.Fatal("expected GCP config")
	}
	if cfg.GCP.ImpersonateServiceAccount != "target-sa@example.iam.gserviceaccount.com" {
		t.Fatalf("ImpersonateServiceAccount = %q", cfg.GCP.ImpersonateServiceAccount)
	}
	if len(cfg.GCP.Delegates) != 2 || cfg.GCP.Delegates[0] != "delegate-1@example.iam.gserviceaccount.com" || cfg.GCP.Delegates[1] != "delegate-2@example.iam.gserviceaccount.com" {
		t.Fatalf("Delegates = %#v", cfg.GCP.Delegates)
	}
	if len(cfg.GCP.Scopes) != 2 || cfg.GCP.Scopes[0] != "https://www.googleapis.com/auth/secretmanager" || cfg.GCP.Scopes[1] != "https://www.googleapis.com/auth/cloud-platform" {
		t.Fatalf("Scopes = %#v", cfg.GCP.Scopes)
	}
	if cfg.GCP.ProjectID != "project-1" || cfg.GCP.SecretID != "openai-admin-key" || cfg.GCP.Version != "latest" || cfg.GCP.JSONKey != "admin_api_key" {
		t.Fatalf("unexpected GCP config: %#v", cfg.GCP)
	}
}

func TestProviderSchemaGCPSecretManagerImpersonationAttributes(t *testing.T) {
	ctx := context.Background()
	p := &openAIProvider{version: "test"}
	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}
	gcpAttr, ok := schemaResp.Schema.Attributes["gcp_secret_manager"].(providerschema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("unexpected gcp_secret_manager schema attribute: %T", schemaResp.Schema.Attributes["gcp_secret_manager"])
	}
	if _, ok := gcpAttr.Attributes["credentials_json"]; ok {
		t.Fatal("credentials_json should not be accepted in provider configuration")
	}
	if _, ok := gcpAttr.Attributes["impersonate_service_account"].(providerschema.StringAttribute); !ok {
		t.Fatalf("unexpected impersonate_service_account schema attribute: %T", gcpAttr.Attributes["impersonate_service_account"])
	}
	delegatesAttr, ok := gcpAttr.Attributes["delegates"].(providerschema.ListAttribute)
	if !ok {
		t.Fatalf("unexpected delegates schema attribute: %T", gcpAttr.Attributes["delegates"])
	}
	if delegatesAttr.ElementType != types.StringType {
		t.Fatalf("delegates element type = %#v", delegatesAttr.ElementType)
	}
	scopesAttr, ok := gcpAttr.Attributes["scopes"].(providerschema.ListAttribute)
	if !ok {
		t.Fatalf("unexpected scopes schema attribute: %T", gcpAttr.Attributes["scopes"])
	}
	if scopesAttr.ElementType != types.StringType {
		t.Fatalf("scopes element type = %#v", scopesAttr.ElementType)
	}
}

func TestSourceConfigFromProviderModelGCPSecretManagerRejectsEmptyDelegates(t *testing.T) {
	ctx := context.Background()
	gcpObject, diags := types.ObjectValue(gcpObjectTypes(), map[string]attr.Value{
		"project_id":                  types.StringValue("project-1"),
		"secret_id":                   types.StringValue("openai-admin-key"),
		"version":                     types.StringNull(),
		"json_key":                    types.StringNull(),
		"impersonate_service_account": types.StringValue("target-sa@example.iam.gserviceaccount.com"),
		"delegates":                   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("delegate-1@example.iam.gserviceaccount.com"), types.StringValue(" \t ")}),
		"scopes":                      types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("object value diagnostics: %v", diags)
	}

	_, diags = sourceConfigFromProviderModel(ctx, providerModel{
		AdminAPIKey:       types.StringNull(),
		AWSSecretsManager: objectNull(awsObjectTypes()),
		GCPSecretManager:  gcpObject,
		AzureKeyVault:     objectNull(azureObjectTypes()),
	})
	if !diags.HasError() {
		t.Fatal("expected delegates diagnostic")
	}
}

func TestSourceConfigFromProviderModelGCPSecretManagerRejectsDelegatesWithoutImpersonation(t *testing.T) {
	ctx := context.Background()
	gcpObject, diags := types.ObjectValue(gcpObjectTypes(), map[string]attr.Value{
		"project_id":                  types.StringValue("project-1"),
		"secret_id":                   types.StringValue("openai-admin-key"),
		"version":                     types.StringNull(),
		"json_key":                    types.StringNull(),
		"impersonate_service_account": types.StringValue(" \t "),
		"delegates":                   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("delegate-1@example.iam.gserviceaccount.com")}),
		"scopes":                      types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("object value diagnostics: %v", diags)
	}

	_, diags = sourceConfigFromProviderModel(ctx, providerModel{
		AdminAPIKey:       types.StringNull(),
		AWSSecretsManager: objectNull(awsObjectTypes()),
		GCPSecretManager:  gcpObject,
		AzureKeyVault:     objectNull(azureObjectTypes()),
	})
	if !diags.HasError() {
		t.Fatal("expected delegates dependency diagnostic")
	}
}

func TestSourceConfigFromProviderModelGCPSecretManagerRejectsEmptyScopes(t *testing.T) {
	ctx := context.Background()
	gcpObject, diags := types.ObjectValue(gcpObjectTypes(), map[string]attr.Value{
		"project_id":                  types.StringValue("project-1"),
		"secret_id":                   types.StringValue("openai-admin-key"),
		"version":                     types.StringNull(),
		"json_key":                    types.StringNull(),
		"impersonate_service_account": types.StringNull(),
		"delegates":                   types.ListNull(types.StringType),
		"scopes":                      types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://www.googleapis.com/auth/secretmanager"), types.StringValue(" \t ")}),
	})
	if diags.HasError() {
		t.Fatalf("object value diagnostics: %v", diags)
	}

	_, diags = sourceConfigFromProviderModel(ctx, providerModel{
		AdminAPIKey:       types.StringNull(),
		AWSSecretsManager: objectNull(awsObjectTypes()),
		GCPSecretManager:  gcpObject,
		AzureKeyVault:     objectNull(azureObjectTypes()),
	})
	if !diags.HasError() {
		t.Fatal("expected scopes diagnostic")
	}
}

func TestValidServiceAccountRole(t *testing.T) {
	if !validServiceAccountRole("member") || !validServiceAccountRole("owner") {
		t.Fatal("expected member and owner to be valid")
	}
	if validServiceAccountRole("none") || validServiceAccountRole("") {
		t.Fatal("expected invalid roles to be rejected")
	}
}

func TestParseImportIDs(t *testing.T) {
	p, s, err := parseTwoPartImportID("proj/sa", "project_id", "service_account_id")
	if err != nil || p != "proj" || s != "sa" {
		t.Fatalf("parseTwoPartImportID = %q %q %v", p, s, err)
	}
	if _, _, err := parseTwoPartImportID("bad", "project_id", "service_account_id"); err == nil {
		t.Fatal("expected two-part parse error")
	}
	p, s, k, err := parseThreePartImportID("proj/sa/key", "project_id", "service_account_id", "api_key_id")
	if err != nil || p != "proj" || s != "sa" || k != "key" {
		t.Fatalf("parseThreePartImportID = %q %q %q %v", p, s, k, err)
	}
	if _, _, _, err := parseThreePartImportID("proj/key", "project_id", "service_account_id", "api_key_id"); err == nil {
		t.Fatal("expected three-part parse error")
	}
}

func TestProjectAPIKeyModelPreservesSensitiveValue(t *testing.T) {
	prior := projectAPIKeyResourceModel{
		ProjectID:        types.StringValue("proj"),
		ServiceAccountID: types.StringValue("sa"),
		Scopes:           types.SetNull(types.StringType),
		Value:            types.StringValue("sk-created"),
	}
	state := projectAPIKeyModelFromAPI(&client.APIKey{ID: "key", Name: "n", RedactedValue: "sk-...", OwnerType: "service_account", OwnerID: "sa", OwnerProjectAccess: "active", CreatedAt: 1}, prior)
	if state.Value.ValueString() != "sk-created" {
		t.Fatalf("value not preserved: %#v", state)
	}
	if state.RedactedValue.ValueString() != "sk-..." {
		t.Fatalf("redacted value missing: %#v", state)
	}
}

func TestProviderMetadataSchemaResourcesAndConfigure(t *testing.T) {
	clearProviderEnv(t)
	ctx := context.Background()
	factory := New("test")
	p := factory().(*openAIProvider)

	var metaResp provider.MetadataResponse
	p.Metadata(ctx, provider.MetadataRequest{}, &metaResp)
	if metaResp.TypeName != "openai" || metaResp.Version != "test" {
		t.Fatalf("metadata = %#v", metaResp)
	}

	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() || len(schemaResp.Schema.Attributes) == 0 {
		t.Fatalf("schema diagnostics=%v attrs=%d", schemaResp.Diagnostics, len(schemaResp.Schema.Attributes))
	}
	if len(p.Resources(ctx)) != 15 {
		t.Fatalf("expected 15 resources")
	}
	if len(p.DataSources(ctx)) != 28 {
		t.Fatalf("expected 28 data sources")
	}

	configPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := configPlan.Set(ctx, &providerModel{AdminAPIKey: types.StringValue("sk-admin"), BaseURL: types.StringNull(), OrganizationID: types.StringNull(), ProjectID: types.StringNull(), AWSSecretsManager: objectNull(awsObjectTypes()), GCPSecretManager: objectNull(gcpObjectTypes()), AzureKeyVault: objectNull(azureObjectTypes())}); diags.HasError() {
		t.Fatalf("config set: %v", diags)
	}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: configPlan.Raw}
	configureResp := provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{Config: config}, &configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("configure diagnostics: %v", configureResp.Diagnostics)
	}
	if configureResp.ResourceData == nil || configureResp.DataSourceData == nil {
		t.Fatalf("provider data not set")
	}
}

func TestProviderConfigureMissingSource(t *testing.T) {
	clearProviderEnv(t)
	ctx := context.Background()
	p := &openAIProvider{version: "test"}
	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	configPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := configPlan.Set(ctx, &providerModel{AdminAPIKey: types.StringNull(), BaseURL: types.StringNull(), OrganizationID: types.StringNull(), ProjectID: types.StringNull(), AWSSecretsManager: objectNull(awsObjectTypes()), GCPSecretManager: objectNull(gcpObjectTypes()), AzureKeyVault: objectNull(azureObjectTypes())}); diags.HasError() {
		t.Fatalf("config set: %v", diags)
	}
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: configPlan.Raw}
	configureResp := provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{Config: config}, &configureResp)
	if !configureResp.Diagnostics.HasError() {
		t.Fatal("expected missing source diagnostics")
	}
}

func TestResourceMetadataConfigureAndImport(t *testing.T) {
	ctx := context.Background()
	data := &providerData{client: &fakeAdminClient{}}

	project := NewProjectResource().(*projectResource)
	var projectMeta resource.MetadataResponse
	project.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "openai"}, &projectMeta)
	if projectMeta.TypeName != "openai_project" {
		t.Fatalf("project metadata = %q", projectMeta.TypeName)
	}
	project.Configure(ctx, resource.ConfigureRequest{ProviderData: data}, &resource.ConfigureResponse{})
	projectSchema := resourceSchema(ctx, t, project)
	projectImport := resource.ImportStateResponse{State: emptyProjectState(ctx, t, projectSchema)}
	project.ImportState(ctx, resource.ImportStateRequest{ID: "proj_1"}, &projectImport)
	if projectImport.Diagnostics.HasError() {
		t.Fatalf("project import diagnostics: %v", projectImport.Diagnostics)
	}

	projectDS := NewProjectDataSource().(*projectDataSource)
	var projectDSMeta datasource.MetadataResponse
	projectDS.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "openai"}, &projectDSMeta)
	if projectDSMeta.TypeName != "openai_project" {
		t.Fatalf("project data source metadata = %q", projectDSMeta.TypeName)
	}
	projectDS.Configure(ctx, datasource.ConfigureRequest{ProviderData: data}, &datasource.ConfigureResponse{})

	projectsDS := NewProjectsDataSource().(*projectsDataSource)
	var projectsDSMeta datasource.MetadataResponse
	projectsDS.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "openai"}, &projectsDSMeta)
	if projectsDSMeta.TypeName != "openai_projects" {
		t.Fatalf("projects data source metadata = %q", projectsDSMeta.TypeName)
	}
	projectsDS.Configure(ctx, datasource.ConfigureRequest{ProviderData: data}, &datasource.ConfigureResponse{})

	accountDS := NewServiceAccountDataSource().(*serviceAccountDataSource)
	var accountDSMeta datasource.MetadataResponse
	accountDS.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "openai"}, &accountDSMeta)
	if accountDSMeta.TypeName != "openai_service_account" {
		t.Fatalf("account data source metadata = %q", accountDSMeta.TypeName)
	}
	accountDS.Configure(ctx, datasource.ConfigureRequest{ProviderData: data}, &datasource.ConfigureResponse{})

	accountsDS := NewServiceAccountsDataSource().(*serviceAccountsDataSource)
	var accountsDSMeta datasource.MetadataResponse
	accountsDS.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "openai"}, &accountsDSMeta)
	if accountsDSMeta.TypeName != "openai_service_accounts" {
		t.Fatalf("accounts data source metadata = %q", accountsDSMeta.TypeName)
	}
	accountsDS.Configure(ctx, datasource.ConfigureRequest{ProviderData: data}, &datasource.ConfigureResponse{})

	account := NewServiceAccountResource().(*serviceAccountResource)
	var accountMeta resource.MetadataResponse
	account.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "openai"}, &accountMeta)
	if accountMeta.TypeName != "openai_service_account" {
		t.Fatalf("account metadata = %q", accountMeta.TypeName)
	}
	account.Configure(ctx, resource.ConfigureRequest{ProviderData: data}, &resource.ConfigureResponse{})
	accountSchema := resourceSchema(ctx, t, account)
	accountImport := resource.ImportStateResponse{State: emptyServiceAccountState(ctx, t, accountSchema)}
	account.ImportState(ctx, resource.ImportStateRequest{ID: "proj_1/sa_1"}, &accountImport)
	if accountImport.Diagnostics.HasError() {
		t.Fatalf("account import diagnostics: %v", accountImport.Diagnostics)
	}

	key := NewProjectAPIKeyResource().(*projectAPIKeyResource)
	var keyMeta resource.MetadataResponse
	key.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "openai"}, &keyMeta)
	if keyMeta.TypeName != "openai_project_api_key" {
		t.Fatalf("key metadata = %q", keyMeta.TypeName)
	}
	key.Configure(ctx, resource.ConfigureRequest{ProviderData: data}, &resource.ConfigureResponse{})
	keySchema := resourceSchema(ctx, t, key)
	keyImport := resource.ImportStateResponse{State: emptyProjectAPIKeyState(ctx, t, keySchema)}
	key.ImportState(ctx, resource.ImportStateRequest{ID: "proj_1/sa_1/key_1"}, &keyImport)
	if keyImport.Diagnostics.HasError() {
		t.Fatalf("key import diagnostics: %v", keyImport.Diagnostics)
	}
}

func objectNull(typesMap map[string]attr.Type) types.Object {
	return types.ObjectNull(typesMap)
}

func emptyProviderModel() providerModel {
	return providerModel{
		AdminAPIKey:       types.StringNull(),
		BaseURL:           types.StringNull(),
		OrganizationID:    types.StringNull(),
		ProjectID:         types.StringNull(),
		AWSSecretsManager: objectNull(awsObjectTypes()),
		GCPSecretManager:  objectNull(gcpObjectTypes()),
		AzureKeyVault:     objectNull(azureObjectTypes()),
	}
}

func envLookup(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func clearProviderEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"OPENAI_ADMIN_KEY",
		"OPENAI_ADMIN_API_KEY",
		"OPENAI_BASE_URL",
		"OPENAI_ORG_ID",
		"OPENAI_ORGANIZATION_ID",
		"OPENAI_PROJECT_ID",
		"OPENAI_AWS_SECRETS_MANAGER_SECRET_ID",
		"OPENAI_AWS_SECRET_ID",
		"OPENAI_AWS_SECRETS_MANAGER_REGION",
		"OPENAI_AWS_SECRETS_MANAGER_VERSION_ID",
		"OPENAI_AWS_SECRETS_MANAGER_VERSION_STAGE",
		"OPENAI_AWS_SECRETS_MANAGER_JSON_KEY",
		"OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN",
		"OPENAI_AWS_ROLE_ARN",
		"OPENAI_AWS_SECRETS_MANAGER_ROLE_SESSION_NAME",
		"OPENAI_AWS_ROLE_SESSION_NAME",
		"OPENAI_AWS_SECRETS_MANAGER_EXTERNAL_ID",
		"OPENAI_AWS_EXTERNAL_ID",
		"OPENAI_AWS_SECRETS_MANAGER_DURATION_SECONDS",
		"OPENAI_AWS_ROLE_DURATION_SECONDS",
		"OPENAI_GCP_SECRET_MANAGER_PROJECT_ID",
		"GOOGLE_CLOUD_PROJECT",
		"GCLOUD_PROJECT",
		"OPENAI_GCP_SECRET_MANAGER_SECRET_ID",
		"OPENAI_GCP_SECRET_ID",
		"OPENAI_GCP_SECRET_MANAGER_VERSION",
		"OPENAI_GCP_SECRET_MANAGER_JSON_KEY",
		"OPENAI_GCP_SECRET_MANAGER_IMPERSONATE_SERVICE_ACCOUNT",
		"OPENAI_GCP_IMPERSONATE_SERVICE_ACCOUNT",
		"OPENAI_GCP_SECRET_MANAGER_DELEGATES",
		"OPENAI_GCP_SECRET_MANAGER_SCOPES",
	} {
		t.Setenv(name, "")
	}
}

func awsObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{"region": types.StringType, "secret_id": types.StringType, "version_id": types.StringType, "version_stage": types.StringType, "json_key": types.StringType, "role_arn": types.StringType, "role_session_name": types.StringType, "external_id": types.StringType, "duration_seconds": types.Int64Type}
}

func gcpObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{"project_id": types.StringType, "secret_id": types.StringType, "version": types.StringType, "json_key": types.StringType, "impersonate_service_account": types.StringType, "delegates": types.ListType{ElemType: types.StringType}, "scopes": types.ListType{ElemType: types.StringType}}
}

func azureObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{"vault_url": types.StringType, "secret_name": types.StringType, "version": types.StringType, "json_key": types.StringType}
}

func diagnosticsContainPath(diags diag.Diagnostics, expected path.Path) bool {
	for _, diagnostic := range diags {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		if ok && withPath.Path().Equal(expected) {
			return true
		}
	}
	return false
}

func emptyProjectState(ctx context.Context, t *testing.T, schema resourceschema.Schema) tfsdk.State {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &projectResourceModel{ID: types.StringNull(), Name: types.StringNull(), ExternalKeyID: types.StringNull(), Geography: types.StringNull(), Status: types.StringNull(), CreatedAt: types.Int64Null(), ArchivedAt: types.Int64Null()}); diags.HasError() {
		t.Fatalf("empty project plan: %v", diags)
	}
	return tfsdk.State{Schema: schema, Raw: plan.Raw}
}

func emptyServiceAccountState(ctx context.Context, t *testing.T, schema resourceschema.Schema) tfsdk.State {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &serviceAccountResourceModel{
		ID:                  types.StringNull(),
		ProjectID:           types.StringNull(),
		Name:                types.StringNull(),
		Role:                types.StringNull(),
		CreatedAt:           types.Int64Null(),
		Scopes:              types.SetNull(types.StringType),
		APIKeyID:            types.StringNull(),
		APIKeyName:          types.StringNull(),
		APIKeyValue:         types.StringNull(),
		APIKeyRedactedValue: types.StringNull(),
		APIKeyCreatedAt:     types.Int64Null(),
		APIKeyLastUsedAt:    types.Int64Null(),
		APIKeyOwnerAccess:   types.StringNull(),
	}); diags.HasError() {
		t.Fatalf("empty account plan: %v", diags)
	}
	return tfsdk.State{Schema: schema, Raw: plan.Raw}
}

func emptyProjectAPIKeyState(ctx context.Context, t *testing.T, schema resourceschema.Schema) tfsdk.State {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &projectAPIKeyResourceModel{ID: types.StringNull(), ProjectID: types.StringNull(), ServiceAccountID: types.StringNull(), Name: types.StringNull(), Scopes: types.SetNull(types.StringType), Value: types.StringNull(), RedactedValue: types.StringNull(), OwnerType: types.StringNull(), OwnerProjectAccess: types.StringNull(), CreatedAt: types.Int64Null(), LastUsedAt: types.Int64Null()}); diags.HasError() {
		t.Fatalf("empty key plan: %v", diags)
	}
	return tfsdk.State{Schema: schema, Raw: plan.Raw}
}
