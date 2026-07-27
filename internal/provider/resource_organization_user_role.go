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
	_ resource.Resource                = &organizationUserRoleResource{}
	_ resource.ResourceWithConfigure   = &organizationUserRoleResource{}
	_ resource.ResourceWithImportState = &organizationUserRoleResource{}
)

type organizationUserRoleResource struct{ client client.AdminClient }

type organizationUserRoleResourceModel struct {
	ID                types.String            `tfsdk:"id"`
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

func NewOrganizationUserRoleResource() resource.Resource { return &organizationUserRoleResource{} }

func (r *organizationUserRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_role"
}

func (r *organizationUserRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI organization role assignment for a user. Destroy unassigns the role from the user.", Attributes: roleAssignmentResourceAttributes("user_id", "OpenAI organization user ID.")}
}

func (r *organizationUserRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationUserRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationUserRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.CreateOrganizationUserRole(ctx, plan.UserID.ValueString(), client.RoleAssignmentCreateRequest{RoleID: plan.RoleID.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to assign OpenAI organization user role", err)
		return
	}
	state, diags := organizationUserRoleResourceModelFromAPI(ctx, assignment, plan.UserID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationUserRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationUserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.GetOrganizationUserRole(ctx, state.UserID.ValueString(), state.RoleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization user role", err)
		return
	}
	newState, diags := organizationUserRoleResourceModelFromAPI(ctx, assignment, state.UserID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationUserRoleResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI organization user role assignments are immutable", "Change user_id or role_id by replacing the resource.")
}

func (r *organizationUserRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationUserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganizationUserRole(ctx, state.UserID.ValueString(), state.RoleID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to unassign OpenAI organization user role", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationUserRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userID, roleID, err := parseTwoPartImportID(req.ID, "user_id", "role_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), types.StringValue(userID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), types.StringValue(roleID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func organizationUserRoleResourceModelFromAPI(ctx context.Context, assignment *client.RoleAssignment, userID string) (organizationUserRoleResourceModel, diag.Diagnostics) {
	item, diags := roleAssignmentItemModelFromAPI(ctx, *assignment)
	return organizationUserRoleResourceModel{ID: types.StringValue(userID + "/" + assignment.ID), UserID: types.StringValue(userID), RoleID: types.StringValue(assignment.ID), Name: item.Name, Description: item.Description, Permissions: item.Permissions, PredefinedRole: item.PredefinedRole, ResourceType: item.ResourceType, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, CreatedBy: item.CreatedBy, AssignmentSources: item.AssignmentSources}, diags
}
