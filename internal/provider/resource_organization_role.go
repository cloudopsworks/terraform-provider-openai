package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationRoleResource{}
	_ resource.ResourceWithConfigure   = &organizationRoleResource{}
	_ resource.ResourceWithImportState = &organizationRoleResource{}
)

type organizationRoleResource struct{ client client.AdminClient }

type organizationRoleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Permissions    types.Set    `tfsdk:"permissions"`
	PredefinedRole types.Bool   `tfsdk:"predefined_role"`
	ResourceType   types.String `tfsdk:"resource_type"`
}

func NewOrganizationRoleResource() resource.Resource { return &organizationRoleResource{} }

func (r *organizationRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_role"
}

func (r *organizationRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization custom role. Destroy deletes the custom role; use lifecycle prevent_destroy for critical organization access controls.",
		Attributes: map[string]resourceschema.Attribute{
			"id":              resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI role ID."},
			"name":            resourceschema.StringAttribute{Required: true, MarkdownDescription: "Unique role name."},
			"description":     resourceschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional role description."},
			"permissions":     resourceschema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Permissions granted by the role."},
			"predefined_role": resourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OpenAI predefines and manages the role."},
			"resource_type":   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Resource type the role applies to, for example api.organization."},
		},
	}
}

func (r *organizationRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := setToSortedStringSlice(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.CreateOrganizationRole(ctx, client.RoleCreateRequest{Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Permissions: permissions})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI organization role", err)
		return
	}
	state, diags := organizationRoleResourceModelFromAPI(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.GetOrganizationRole(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization role", err)
		return
	}
	newState, diags := organizationRoleResourceModelFromAPI(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := setToSortedStringSlice(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.UpdateOrganizationRole(ctx, plan.ID.ValueString(), client.RoleUpdateRequest{Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Permissions: permissions})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization role", err)
		return
	}
	state, diags := organizationRoleResourceModelFromAPI(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganizationRole(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI organization role", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func organizationRoleResourceModelFromAPI(ctx context.Context, role *client.Role) (organizationRoleResourceModel, diag.Diagnostics) {
	permissions, diags := setStringValue(ctx, role.Permissions)
	return organizationRoleResourceModel{ID: types.StringValue(role.ID), Name: types.StringValue(role.Name), Description: stringOrNull(role.Description), Permissions: permissions, PredefinedRole: types.BoolValue(role.PredefinedRole), ResourceType: stringOrNull(role.ResourceType)}, diags
}
