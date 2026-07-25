package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &projectAPIKeyResource{}
	_ resource.ResourceWithConfigure   = &projectAPIKeyResource{}
	_ resource.ResourceWithImportState = &projectAPIKeyResource{}
)

type projectAPIKeyResource struct {
	client client.AdminClient
}

type projectAPIKeyResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	ServiceAccountID   types.String `tfsdk:"service_account_id"`
	Name               types.String `tfsdk:"name"`
	Scopes             types.Set    `tfsdk:"scopes"`
	Value              types.String `tfsdk:"value"`
	RedactedValue      types.String `tfsdk:"redacted_value"`
	OwnerType          types.String `tfsdk:"owner_type"`
	OwnerProjectAccess types.String `tfsdk:"owner_project_access"`
	CreatedAt          types.Int64  `tfsdk:"created_at"`
	LastUsedAt         types.Int64  `tfsdk:"last_used_at"`
}

func NewProjectAPIKeyResource() resource.Resource { return &projectAPIKeyResource{} }

func (r *projectAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_key"
}

func (r *projectAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI project API key for a project service account. The unredacted key value is returned only on create and preserved as a Sensitive computed state attribute during refresh.",
		Attributes: map[string]resourceschema.Attribute{
			"id":                   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project API key ID."},
			"project_id":           resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"service_account_id":   resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI service account ID that owns the key.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":                 resourceschema.StringAttribute{Required: true, MarkdownDescription: "API key name. OpenAI does not expose update for service-account API keys, so changes replace the resource.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"scopes":               resourceschema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional API key scopes. Scopes are create-only and changing them replaces the resource.", PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}},
			"value":                resourceschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Unredacted API key value returned only during create. Terraform state must be protected as sensitive."},
			"redacted_value":       resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Redacted API key value returned by read operations."},
			"owner_type":           resourceschema.StringAttribute{Computed: true, MarkdownDescription: "API key owner type returned by OpenAI."},
			"owner_project_access": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Whether the key owner currently has effective access to the project."},
			"created_at":           resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for API key creation."},
			"last_used_at":         resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the key was last used, if known."},
		},
	}
}

func (r *projectAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopes, diags := setToStringSlice(ctx, plan.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateServiceAccountAPIKey(ctx, plan.ProjectID.ValueString(), plan.ServiceAccountID.ValueString(), client.ServiceAccountAPIKeyCreateRequest{Name: plan.Name.ValueString(), Scopes: scopes})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI project API key", err)
		return
	}
	state := projectAPIKeyModelFromCreate(ctx, created, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := r.client.GetProjectAPIKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project API key", err)
		return
	}
	newState := projectAPIKeyModelFromAPI(apiKey, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *projectAPIKeyResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI project API keys are immutable", "OpenAI does not expose an update operation for service-account project API keys. Change name, scopes, project_id, or service_account_id by replacing the resource.")
}

func (r *projectAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteProjectAPIKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI project API key", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *projectAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, serviceAccountID, apiKeyID, err := parseThreePartImportID(req.ID, "project_id", "service_account_id", "api_key_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_id"), serviceAccountID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), apiKeyID)...)
}

func projectAPIKeyModelFromCreate(ctx context.Context, created *client.ServiceAccountAPIKeyCreateResponse, plan projectAPIKeyResourceModel) projectAPIKeyResourceModel {
	scopes := plan.Scopes
	if scopes.IsUnknown() {
		scopes, _ = setStringValueOrNull(ctx, nil)
	}
	return projectAPIKeyResourceModel{
		ID:               types.StringValue(created.ID),
		ProjectID:        plan.ProjectID,
		ServiceAccountID: plan.ServiceAccountID,
		Name:             types.StringValue(created.Name),
		Scopes:           scopes,
		Value:            types.StringValue(created.Value),
		RedactedValue:    types.StringNull(),
		OwnerType:        types.StringValue("service_account"),
		CreatedAt:        int64OrNull(created.CreatedAt),
		LastUsedAt:       types.Int64Null(),
	}
}

func projectAPIKeyModelFromAPI(apiKey *client.APIKey, prior projectAPIKeyResourceModel) projectAPIKeyResourceModel {
	value := prior.Value
	if value.IsUnknown() {
		value = types.StringNull()
	}
	serviceAccountID := prior.ServiceAccountID
	if apiKey.OwnerID != "" {
		serviceAccountID = types.StringValue(apiKey.OwnerID)
	}
	return projectAPIKeyResourceModel{
		ID:                 types.StringValue(apiKey.ID),
		ProjectID:          prior.ProjectID,
		ServiceAccountID:   serviceAccountID,
		Name:               types.StringValue(apiKey.Name),
		Scopes:             prior.Scopes,
		Value:              value,
		RedactedValue:      stringOrNull(apiKey.RedactedValue),
		OwnerType:          stringOrNull(apiKey.OwnerType),
		OwnerProjectAccess: stringOrNull(apiKey.OwnerProjectAccess),
		CreatedAt:          int64OrNull(apiKey.CreatedAt),
		LastUsedAt:         int64OrNull(apiKey.LastUsedAt),
	}
}
