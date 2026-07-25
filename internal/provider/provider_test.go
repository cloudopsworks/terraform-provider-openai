package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
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
	cfg, diags := sourceConfigFromProviderModel(context.Background(), providerModel{AdminAPIKey: types.StringValue("sk")})
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if cfg.Direct != "sk" || cfg.AWS != nil || cfg.GCP != nil || cfg.Azure != nil {
		t.Fatalf("unexpected source config: %#v", cfg)
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
	if len(p.Resources(ctx)) != 3 {
		t.Fatalf("expected 3 resources")
	}
	if p.DataSources(ctx) != nil {
		t.Fatalf("expected nil data sources")
	}

	configPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := configPlan.Set(ctx, &providerModel{AdminAPIKey: types.StringValue("sk-admin"), AWSSecretsManager: objectNull(awsObjectTypes()), GCPSecretManager: objectNull(gcpObjectTypes()), AzureKeyVault: objectNull(azureObjectTypes())}); diags.HasError() {
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
	ctx := context.Background()
	p := &openAIProvider{version: "test"}
	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	configPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := configPlan.Set(ctx, &providerModel{AdminAPIKey: types.StringNull(), BaseURL: types.StringNull(), AWSSecretsManager: objectNull(awsObjectTypes()), GCPSecretManager: objectNull(gcpObjectTypes()), AzureKeyVault: objectNull(azureObjectTypes())}); diags.HasError() {
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

func awsObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{"region": types.StringType, "secret_id": types.StringType, "version_id": types.StringType, "version_stage": types.StringType, "json_key": types.StringType}
}

func gcpObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{"project_id": types.StringType, "secret_id": types.StringType, "version": types.StringType, "json_key": types.StringType}
}

func azureObjectTypes() map[string]attr.Type {
	return map[string]attr.Type{"vault_url": types.StringType, "secret_name": types.StringType, "version": types.StringType, "json_key": types.StringType}
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
	if diags := plan.Set(ctx, &serviceAccountResourceModel{ID: types.StringNull(), ProjectID: types.StringNull(), Name: types.StringNull(), Role: types.StringNull(), CreatedAt: types.Int64Null()}); diags.HasError() {
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
