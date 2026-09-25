package provider

import (
	"context"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &projectUserResource{}
	_ resource.ResourceWithConfigure   = &projectUserResource{}
	_ resource.ResourceWithImportState = &projectUserResource{}
)

type projectUserResource struct{ client client.AdminClient }

type projectUserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	Email     types.String `tfsdk:"email"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	AddedAt   types.Int64  `tfsdk:"added_at"`
}

func NewProjectUserResource() resource.Resource { return &projectUserResource{} }
func (r *projectUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}
func (r *projectUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI project user membership. Destroy revokes the user's project access.",
		Attributes: map[string]resourceschema.Attribute{
			"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI user ID."},
			"project_id": resourceschema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "OpenAI project ID."},
			"user_id":    resourceschema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Existing OpenAI user ID. Provide user_id or email."},
			"email":      resourceschema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "User email address. Provide user_id or email."},
			"name":       resourceschema.StringAttribute{Computed: true, MarkdownDescription: "User display name."},
			"role":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "Project role for the user."},
			"added_at":   resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when access was granted."},
		},
	}
}

func (r *projectUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.UserID.IsNull() && plan.Email.IsNull() {
		resp.Diagnostics.AddError("Missing OpenAI project user identity", "Set user_id or email to grant project access.")
		return
	}
	user, err := r.client.CreateProjectUser(ctx, plan.ProjectID.ValueString(), client.ProjectUserCreateRequest{UserID: stringValue(plan.UserID), Email: stringValue(plan.Email), Role: plan.Role.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to add OpenAI project user", err)
		return
	}
	state := projectUserResourceModelFromAPI(user, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := r.client.GetProjectUser(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project user", err)
		return
	}
	next := projectUserResourceModelFromAPI(user, state.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}
func (r *projectUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := r.client.UpdateProjectUser(ctx, plan.ProjectID.ValueString(), plan.ID.ValueString(), client.ProjectUserUpdateRequest{Role: plan.Role.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project user", err)
		return
	}
	state := projectUserResourceModelFromAPI(user, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectUser(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to remove OpenAI project user", err)
		return
	}
	resp.State.RemoveResource(ctx)
}
func (r *projectUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, userID, err := parseTwoPartImportID(req.ID, "project_id", "user_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(userID))...)
}
func projectUserResourceModelFromAPI(user *client.ProjectUser, projectID types.String) projectUserResourceModel {
	return projectUserResourceModel{ID: types.StringValue(user.ID), ProjectID: projectID, UserID: types.StringValue(user.ID), Email: stringOrNull(user.Email), Name: stringOrNull(user.Name), Role: stringOrNull(user.Role), AddedAt: int64OrNull(user.AddedAt)}
}
