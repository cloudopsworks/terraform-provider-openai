package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationGroupUserResource{}
	_ resource.ResourceWithConfigure   = &organizationGroupUserResource{}
	_ resource.ResourceWithImportState = &organizationGroupUserResource{}
)

type organizationGroupUserResource struct{ client client.AdminClient }

type organizationGroupUserResourceModel struct {
	ID               types.String `tfsdk:"id"`
	GroupID          types.String `tfsdk:"group_id"`
	UserID           types.String `tfsdk:"user_id"`
	Email            types.String `tfsdk:"email"`
	Name             types.String `tfsdk:"name"`
	IsServiceAccount types.Bool   `tfsdk:"is_service_account"`
	Picture          types.String `tfsdk:"picture"`
	UserType         types.String `tfsdk:"user_type"`
}

func NewOrganizationGroupUserResource() resource.Resource { return &organizationGroupUserResource{} }

func (r *organizationGroupUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_user"
}

func (r *organizationGroupUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization group membership. Destroy removes the user from the group.",
		Attributes: map[string]resourceschema.Attribute{
			"id":                 resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Composite membership ID in group_id/user_id format."},
			"group_id":           resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization group ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"user_id":            resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization user ID to add to the group.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"email":              resourceschema.StringAttribute{Computed: true, MarkdownDescription: "User email address."},
			"name":               resourceschema.StringAttribute{Computed: true, MarkdownDescription: "User display name."},
			"is_service_account": resourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is a service account."},
			"picture":            resourceschema.StringAttribute{Computed: true, MarkdownDescription: "User profile picture URL when available."},
			"user_type":          resourceschema.StringAttribute{Computed: true, MarkdownDescription: "User type returned by OpenAI, such as user or tenant_user."},
		},
	}
}

func (r *organizationGroupUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationGroupUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationGroupUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := r.client.CreateOrganizationGroupUser(ctx, plan.GroupID.ValueString(), client.OrganizationGroupUserCreateRequest{UserID: plan.UserID.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to add OpenAI organization group user", err)
		return
	}
	state := organizationGroupUserResourceModelFromAPI(user, plan.GroupID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationGroupUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationGroupUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := r.client.GetOrganizationGroupUser(ctx, state.GroupID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization group user", err)
		return
	}
	newState := organizationGroupUserResourceModelFromAPI(user, state.GroupID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationGroupUserResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI organization group memberships are immutable", "Change group_id or user_id by replacing the resource.")
}

func (r *organizationGroupUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationGroupUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganizationGroupUser(ctx, state.GroupID.ValueString(), state.UserID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to remove OpenAI organization group user", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationGroupUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	groupID, userID, err := parseTwoPartImportID(req.ID, "group_id", "user_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), types.StringValue(groupID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), types.StringValue(userID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func organizationGroupUserResourceModelFromAPI(user *client.OrganizationGroupUser, groupID string) organizationGroupUserResourceModel {
	return organizationGroupUserResourceModel{ID: types.StringValue(groupID + "/" + user.ID), GroupID: types.StringValue(groupID), UserID: types.StringValue(user.ID), Email: stringOrNull(user.Email), Name: stringOrNull(user.Name), IsServiceAccount: types.BoolValue(user.IsServiceAccount), Picture: stringOrNull(user.Picture), UserType: stringOrNull(user.UserType)}
}
