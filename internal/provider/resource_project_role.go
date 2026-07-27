package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &projectRoleResource{}
	_ resource.ResourceWithConfigure   = &projectRoleResource{}
	_ resource.ResourceWithImportState = &projectRoleResource{}
)

type projectRoleResource struct{ client client.AdminClient }

type projectRoleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Permissions    types.Set    `tfsdk:"permissions"`
	PredefinedRole types.Bool   `tfsdk:"predefined_role"`
	ResourceType   types.String `tfsdk:"resource_type"`
}

func NewProjectRoleResource() resource.Resource { return &projectRoleResource{} }

func (r *projectRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_role"
}

func (r *projectRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI project custom role. Destroy deletes the custom role from the project; use lifecycle prevent_destroy for critical project access controls.",
		Attributes: map[string]resourceschema.Attribute{
			"id":              resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI role ID."},
			"project_id":      resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the role.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":            resourceschema.StringAttribute{Required: true, MarkdownDescription: "Unique role name."},
			"description":     resourceschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional role description."},
			"permissions":     resourceschema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Permissions granted by the role."},
			"predefined_role": resourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OpenAI predefines and manages the role."},
			"resource_type":   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Resource type the role applies to, for example api.project."},
		},
	}
}

func (r *projectRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := setToSortedStringSlice(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.CreateProjectRole(ctx, plan.ProjectID.ValueString(), client.RoleCreateRequest{Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Permissions: permissions})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI project role", err)
		return
	}
	state, diags := projectRoleResourceModelFromAPI(ctx, role, plan.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.GetProjectRole(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project role", err)
		return
	}
	newState, diags := projectRoleResourceModelFromAPI(ctx, role, state.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *projectRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := setToSortedStringSlice(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.UpdateProjectRole(ctx, plan.ProjectID.ValueString(), plan.ID.ValueString(), client.RoleUpdateRequest{Name: plan.Name.ValueString(), Description: stringValue(plan.Description), Permissions: permissions})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project role", err)
		return
	}
	state, diags := projectRoleResourceModelFromAPI(ctx, role, plan.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectRole(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI project role", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *projectRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, roleID, err := parseTwoPartImportID(req.ID, "project_id", "role_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(roleID))...)
}

func projectRoleResourceModelFromAPI(ctx context.Context, role *client.Role, projectID types.String) (projectRoleResourceModel, diag.Diagnostics) {
	permissions, diags := setStringValue(ctx, role.Permissions)
	return projectRoleResourceModel{ID: types.StringValue(role.ID), ProjectID: projectID, Name: types.StringValue(role.Name), Description: stringOrNull(role.Description), Permissions: permissions, PredefinedRole: types.BoolValue(role.PredefinedRole), ResourceType: stringOrNull(role.ResourceType)}, diags
}
