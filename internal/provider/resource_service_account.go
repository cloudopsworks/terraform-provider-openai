package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewServiceAccountResource() resource.Resource { return &serviceAccountResource{} }

func (r *serviceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		Version:             1,
		MarkdownDescription: "OpenAI project service account. Creation uses create_service_account_only=true to avoid an unmanaged default key; use openai_project_api_key for explicit key issuance.",
		Attributes: map[string]resourceschema.Attribute{
			"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI service account ID."},
			"project_id": resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the service account.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "Service account name."},
			"role":       resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("member"), MarkdownDescription: "Project role for the service account. Supported values are member and owner.", Validators: []validator.String{serviceAccountRoleValidator{}}},
			"created_at": resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for service-account creation."},
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
	account, err := r.client.CreateServiceAccount(ctx, plan.ProjectID.ValueString(), client.ServiceAccountCreateRequest{Name: plan.Name.ValueString(), Role: plan.Role.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI service account", err)
		return
	}
	state := serviceAccountModelFromAPI(account, plan.ProjectID.ValueString())
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
	newState := serviceAccountModelFromAPI(account, state.ProjectID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
	state := serviceAccountModelFromAPI(account, plan.ProjectID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
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

func serviceAccountModelFromAPI(account *client.ServiceAccount, projectID string) serviceAccountResourceModel {
	return serviceAccountResourceModel{
		ID:        types.StringValue(account.ID),
		ProjectID: types.StringValue(projectID),
		Name:      types.StringValue(account.Name),
		Role:      stringOrNull(account.Role),
		CreatedAt: int64OrNull(account.CreatedAt),
	}
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
