package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                 = &serviceAccountResource{}
	_ resource.ResourceWithConfigure    = &serviceAccountResource{}
	_ resource.ResourceWithImportState  = &serviceAccountResource{}
	_ resource.ResourceWithUpgradeState = &serviceAccountResource{}
)

type serviceAccountResource struct {
	client client.AdminClient
}

type serviceAccountResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	Name                types.String `tfsdk:"name"`
	Role                types.String `tfsdk:"role"`
	CreatedAt           types.Int64  `tfsdk:"created_at"`
	Scopes              types.Set    `tfsdk:"scopes"`
	APIKeyID            types.String `tfsdk:"api_key_id"`
	APIKeyName          types.String `tfsdk:"api_key_name"`
	APIKeyValue         types.String `tfsdk:"api_key_value"`
	APIKeyRedactedValue types.String `tfsdk:"api_key_redacted_value"`
	APIKeyCreatedAt     types.Int64  `tfsdk:"api_key_created_at"`
	APIKeyLastUsedAt    types.Int64  `tfsdk:"api_key_last_used_at"`
	APIKeyOwnerAccess   types.String `tfsdk:"api_key_owner_project_access"`
}

func NewServiceAccountResource() resource.Resource { return &serviceAccountResource{} }

func (r *serviceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		Version:             1,
		MarkdownDescription: "OpenAI project service account. OpenAI creates and returns a default API key during service-account creation unless scopes are configured; scoped bootstrap keys are created through the service-account API-key endpoint.",
		Attributes: map[string]resourceschema.Attribute{
			"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI service account ID."},
			"project_id": resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the service account.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "Service account name."},
			"role":       resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("member"), MarkdownDescription: "Project role for the service account. Supported values are member and owner.", Validators: []validator.String{serviceAccountRoleValidator{}}},
			"created_at": resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for service-account creation."},

			"scopes": resourceschema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Optional scopes for the API key created during service-account creation. OpenAI does not accept scopes on the service-account create endpoint, so the provider creates the service account without a default key and then creates a scoped service-account API key. Scope changes replace the service account and bootstrap key.",
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"api_key_id":                   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the API key created during service-account creation."},
			"api_key_name":                 resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the API key created during service-account creation."},
			"api_key_value":                resourceschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Unredacted API key value returned only during service-account creation. Protect Terraform state accordingly."},
			"api_key_redacted_value":       resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Redacted API key value returned by read operations when available."},
			"api_key_created_at":           resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for API key creation."},
			"api_key_last_used_at":         resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the API key was last used, if known."},
			"api_key_owner_project_access": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Whether the API key owner currently has effective access to the project."},
		},
	}
}

func (r *serviceAccountResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: serviceAccountResourceV0Schema(),
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior serviceAccountResourceModel
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				resp.State.Schema = serviceAccountResourceCurrentSchema()
				resp.Diagnostics.Append(resp.State.Set(ctx, &prior)...)
			},
		},
	}
}

func (r *serviceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, err := configureClient(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("OpenAI provider not configured", err.Error())
		return
	}
	r.client = data.client
}

func (r *serviceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !validServiceAccountRole(plan.Role.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("role"), "Invalid OpenAI service account role", "Supported values are member and owner.")
		return
	}
	scopes, scopeDiags := setToStringSlice(ctx, plan.Scopes)
	resp.Diagnostics.Append(scopeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createScopedKey := len(scopes) > 0
	account, err := r.client.CreateServiceAccount(ctx, plan.ProjectID.ValueString(), client.ServiceAccountCreateRequest{Name: plan.Name.ValueString(), Role: plan.Role.ValueString(), CreateServiceAccountOnly: createScopedKey})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI service account", err)
		return
	}
	state := serviceAccountModelFromAPI(account, plan)
	if createScopedKey {
		apiKey, err := createProjectAPIKey(ctx, r.client, plan.ProjectID.ValueString(), account.ID, plan.Name.ValueString(), scopes)
		if err != nil {
			if cleanupErr := r.client.DeleteServiceAccount(ctx, plan.ProjectID.ValueString(), account.ID); cleanupErr != nil && !client.IsNotFound(cleanupErr) {
				resp.Diagnostics.AddWarning("OpenAI service account cleanup failed", client.ErrorSummary(cleanupErr))
			}
			addClientError(&resp.Diagnostics, "Unable to create scoped OpenAI service account API key", err)
			return
		}
		state = serviceAccountModelWithAPIKeyCreate(apiKey, state)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := r.client.GetServiceAccount(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI service account", err)
		return
	}
	newState := serviceAccountModelFromAPI(account, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !validServiceAccountRole(plan.Role.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("role"), "Invalid OpenAI service account role", "Supported values are member and owner.")
		return
	}
	account, err := r.client.UpdateServiceAccount(ctx, plan.ProjectID.ValueString(), plan.ID.ValueString(), client.ServiceAccountUpdateRequest{Name: plan.Name.ValueString(), Role: plan.Role.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI service account", err)
		return
	}
	newState := serviceAccountModelFromAPI(account, state)
	newState.Name = plan.Name
	newState.Role = plan.Role
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !state.APIKeyID.IsNull() && !state.APIKeyID.IsUnknown() && state.APIKeyID.ValueString() != "" {
		if err := r.client.DeleteProjectAPIKey(ctx, state.ProjectID.ValueString(), state.APIKeyID.ValueString()); err != nil && !client.IsNotFound(err) {
			addClientError(&resp.Diagnostics, "Unable to delete OpenAI service account API key", err)
			return
		}
	}
	err := r.client.DeleteServiceAccount(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI service account", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *serviceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, serviceAccountID, err := parseTwoPartImportID(req.ID, "project_id", "service_account_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(serviceAccountID))...)
}

func serviceAccountModelFromAPI(account *client.ServiceAccount, prior serviceAccountResourceModel) serviceAccountResourceModel {
	state := serviceAccountResourceModel{
		ID:                  types.StringValue(account.ID),
		ProjectID:           prior.ProjectID,
		Name:                types.StringValue(account.Name),
		Role:                stringOrNull(account.Role),
		CreatedAt:           int64OrNull(account.CreatedAt),
		Scopes:              setOrNull(prior.Scopes),
		APIKeyID:            stringOrNullFromState(prior.APIKeyID),
		APIKeyName:          stringOrNullFromState(prior.APIKeyName),
		APIKeyValue:         stringOrNullFromState(prior.APIKeyValue),
		APIKeyRedactedValue: stringOrNullFromState(prior.APIKeyRedactedValue),
		APIKeyCreatedAt:     int64OrNullFromState(prior.APIKeyCreatedAt),
		APIKeyLastUsedAt:    int64OrNullFromState(prior.APIKeyLastUsedAt),
		APIKeyOwnerAccess:   stringOrNullFromState(prior.APIKeyOwnerAccess),
	}
	if account.APIKey != nil {
		state = serviceAccountModelWithAPIKeyCreate(account.APIKey, state)
	}
	return state
}

func serviceAccountModelWithAPIKeyCreate(apiKey *client.ServiceAccountAPIKeyCreateResponse, state serviceAccountResourceModel) serviceAccountResourceModel {
	state.APIKeyID = types.StringValue(apiKey.ID)
	state.APIKeyName = stringOrNull(apiKey.Name)
	state.APIKeyValue = types.StringValue(apiKey.Value)
	state.APIKeyRedactedValue = types.StringNull()
	state.APIKeyCreatedAt = int64OrNull(apiKey.CreatedAt)
	state.APIKeyLastUsedAt = types.Int64Null()
	state.APIKeyOwnerAccess = types.StringNull()
	return state
}

func validServiceAccountRole(role string) bool {
	switch role {
	case "member", "owner":
		return true
	default:
		return false
	}
}

func serviceAccountResourceCurrentSchema() resourceschema.Schema {
	var resp resource.SchemaResponse
	NewServiceAccountResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp.Schema
}

func serviceAccountResourceV0Schema() *resourceschema.Schema {
	schema := serviceAccountResourceCurrentSchema()
	schema.Version = 0
	roleAttr := schema.Attributes["role"].(resourceschema.StringAttribute)
	roleAttr.Validators = nil
	schema.Attributes["role"] = roleAttr
	return &schema
}

type serviceAccountRoleValidator struct{}

func (serviceAccountRoleValidator) Description(_ context.Context) string {
	return "value must be one of: member, owner"
}

func (v serviceAccountRoleValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (serviceAccountRoleValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || validServiceAccountRole(req.ConfigValue.ValueString()) {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid OpenAI service account role", "Supported values are member and owner.")
}
