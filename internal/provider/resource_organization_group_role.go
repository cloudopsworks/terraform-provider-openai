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
	_ resource.Resource                = &organizationGroupRoleResource{}
	_ resource.ResourceWithConfigure   = &organizationGroupRoleResource{}
	_ resource.ResourceWithImportState = &organizationGroupRoleResource{}
)

type organizationGroupRoleResource struct{ client client.AdminClient }

type organizationGroupRoleResourceModel struct {
	ID                types.String            `tfsdk:"id"`
	GroupID           types.String            `tfsdk:"group_id"`
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

func NewOrganizationGroupRoleResource() resource.Resource { return &organizationGroupRoleResource{} }

func (r *organizationGroupRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_role"
}

func (r *organizationGroupRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI organization role assignment for a group. Destroy unassigns the role from the group.", Attributes: roleAssignmentResourceAttributes("group_id", "OpenAI organization group ID.")}
}

func (r *organizationGroupRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationGroupRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationGroupRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.CreateOrganizationGroupRole(ctx, plan.GroupID.ValueString(), client.RoleAssignmentCreateRequest{RoleID: plan.RoleID.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to assign OpenAI organization group role", err)
		return
	}
	state, diags := organizationGroupRoleResourceModelFromAPI(ctx, assignment, plan.GroupID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationGroupRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationGroupRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.GetOrganizationGroupRole(ctx, state.GroupID.ValueString(), state.RoleID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization group role", err)
		return
	}
	newState, diags := organizationGroupRoleResourceModelFromAPI(ctx, assignment, state.GroupID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationGroupRoleResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI organization group role assignments are immutable", "Change group_id or role_id by replacing the resource.")
}

func (r *organizationGroupRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationGroupRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganizationGroupRole(ctx, state.GroupID.ValueString(), state.RoleID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to unassign OpenAI organization group role", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationGroupRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	groupID, roleID, err := parseTwoPartImportID(req.ID, "group_id", "role_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), types.StringValue(groupID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), types.StringValue(roleID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func organizationGroupRoleResourceModelFromAPI(ctx context.Context, assignment *client.RoleAssignment, groupID string) (organizationGroupRoleResourceModel, diag.Diagnostics) {
	item, diags := roleAssignmentItemModelFromAPI(ctx, *assignment)
	return organizationGroupRoleResourceModel{ID: types.StringValue(groupID + "/" + assignment.ID), GroupID: types.StringValue(groupID), RoleID: types.StringValue(assignment.ID), Name: item.Name, Description: item.Description, Permissions: item.Permissions, PredefinedRole: item.PredefinedRole, ResourceType: item.ResourceType, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, CreatedBy: item.CreatedBy, AssignmentSources: item.AssignmentSources}, diags
}
