package provider

import (
	"context"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &projectGroupResource{}
	_ resource.ResourceWithConfigure   = &projectGroupResource{}
	_ resource.ResourceWithImportState = &projectGroupResource{}
)

type projectGroupResource struct{ client client.AdminClient }

type projectGroupResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	GroupID   types.String `tfsdk:"group_id"`
	GroupName types.String `tfsdk:"group_name"`
	GroupType types.String `tfsdk:"group_type"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewProjectGroupResource() resource.Resource { return &projectGroupResource{} }
func (r *projectGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group"
}
func (r *projectGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI project group membership. Destroy revokes the group's project access.",
		Attributes: map[string]resourceschema.Attribute{
			"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Composite project/group membership ID."},
			"project_id": resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "OpenAI project ID."},
			"group_id":   resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "OpenAI organization group ID."},
			"group_name": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group display name."},
			"group_type": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group type returned by OpenAI."},
			"role":       resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Project role granted to the group."},
			"created_at": resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when membership was created."},
		},
	}
}

func (r *projectGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.CreateProjectGroup(ctx, plan.ProjectID.ValueString(), client.ProjectGroupCreateRequest{GroupID: plan.GroupID.ValueString(), Role: plan.Role.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to add OpenAI project group", err)
		return
	}
	state := projectGroupResourceModelFromAPI(group, plan.Role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.GetProjectGroup(ctx, state.ProjectID.ValueString(), state.GroupID.ValueString(), client.ProjectGroupGetRequest{GroupType: state.GroupType.ValueString()})
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project group", err)
		return
	}
	next := projectGroupResourceModelFromAPI(group, state.Role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}
func (r *projectGroupResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI project group memberships are immutable", "Change project_id, group_id, group_type, or role by replacing the resource.")
}
func (r *projectGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectGroup(ctx, state.ProjectID.ValueString(), state.GroupID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to remove OpenAI project group", err)
		return
	}
	resp.State.RemoveResource(ctx)
}
func (r *projectGroupResource) ImportState(_ context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Project group membership import is not supported", "OpenAI does not return the project role granted to a group membership. Importing would produce incomplete state and could replace or revoke access on the next apply. Declare the membership with its role instead.")
}
func projectGroupResourceModelFromAPI(group *client.ProjectGroup, role types.String) projectGroupResourceModel {
	return projectGroupResourceModel{ID: types.StringValue(group.ProjectID + "/" + group.GroupID), ProjectID: types.StringValue(group.ProjectID), GroupID: types.StringValue(group.GroupID), GroupName: stringOrNull(group.GroupName), GroupType: stringOrNull(group.GroupType), Role: role, CreatedAt: int64OrNull(group.CreatedAt)}
}
