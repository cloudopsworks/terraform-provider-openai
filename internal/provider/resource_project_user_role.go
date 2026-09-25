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
	_ resource.Resource                = &projectUserRoleResource{}
	_ resource.ResourceWithConfigure   = &projectUserRoleResource{}
	_ resource.ResourceWithImportState = &projectUserRoleResource{}
)

type projectUserRoleResource struct{ client client.AdminClient }

type projectUserRoleResourceModel struct {
	ID                types.String            `tfsdk:"id"`
	ProjectID         types.String            `tfsdk:"project_id"`
	UserID            types.String            `tfsdk:"user_id"`
	RoleID            types.String            `tfsdk:"role_id"`
	Name              types.String            `tfsdk:"name"`
	Description       types.String            `tfsdk:"description"`
	Permissions       types.Set               `tfsdk:"permissions"`
	PredefinedRole    types.Bool              `tfsdk:"predefined_role"`
	ResourceType      types.String            `tfsdk:"resource_type"`
	CreatedAt         types.Int64             `tfsdk:"created_at"`
	UpdatedAt         types.Int64             `tfsdk:"updated_at"`
	CreatedBy         types.String            `tfsdk:"created_by"`
	AssignmentSources []assignmentSourceModel `tfsdk:"assignment_sources"`
}

func NewProjectUserRoleResource() resource.Resource { return &projectUserRoleResource{} }

func (r *projectUserRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user_role"
}

func (r *projectUserRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := roleAssignmentResourceAttributes("user_id", "OpenAI project user ID.")
	attrs["project_id"] = resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI project role assignment for a user. Destroy unassigns the role from the user.", Attributes: attrs}
}

func (r *projectUserRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectUserRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectUserRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.CreateProjectUserRole(ctx, plan.ProjectID.ValueString(), plan.UserID.ValueString(), client.RoleAssignmentCreateRequest{RoleID: plan.RoleID.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to assign OpenAI project user role", err)
		return
	}
	state, diags := projectUserRoleResourceModelFromAPI(ctx, assignment, plan.ProjectID.ValueString(), plan.UserID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectUserRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectUserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.GetProjectUserRole(ctx, state.ProjectID.ValueString(), state.UserID.ValueString(), state.RoleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project user role", err)
		return
	}
	newState, diags := projectUserRoleResourceModelFromAPI(ctx, assignment, state.ProjectID.ValueString(), state.UserID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *projectUserRoleResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI project user role assignments are immutable", "Change project_id, user_id, or role_id by replacing the resource.")
}

func (r *projectUserRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectUserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectUserRole(ctx, state.ProjectID.ValueString(), state.UserID.ValueString(), state.RoleID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to unassign OpenAI project user role", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *projectUserRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, userID, roleID, err := parseThreePartImportID(req.ID, "project_id", "user_id", "role_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), types.StringValue(userID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), types.StringValue(roleID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func projectUserRoleResourceModelFromAPI(ctx context.Context, assignment *client.RoleAssignment, projectID, userID string) (projectUserRoleResourceModel, diag.Diagnostics) {
	item, diags := roleAssignmentItemModelFromAPI(ctx, *assignment)
	return projectUserRoleResourceModel{ID: types.StringValue(projectID + "/" + userID + "/" + assignment.ID), ProjectID: types.StringValue(projectID), UserID: types.StringValue(userID), RoleID: types.StringValue(assignment.ID), Name: item.Name, Description: item.Description, Permissions: item.Permissions, PredefinedRole: item.PredefinedRole, ResourceType: item.ResourceType, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, CreatedBy: item.CreatedBy, AssignmentSources: item.AssignmentSources}, diags
}
