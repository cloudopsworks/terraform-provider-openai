package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationInviteResource{}
	_ resource.ResourceWithConfigure   = &organizationInviteResource{}
	_ resource.ResourceWithImportState = &organizationInviteResource{}
)

type organizationInviteResource struct{ client client.AdminClient }

type organizationInviteResourceModel struct {
	ID         types.String         `tfsdk:"id"`
	Email      types.String         `tfsdk:"email"`
	Role       types.String         `tfsdk:"role"`
	Projects   []inviteProjectModel `tfsdk:"projects"`
	Status     types.String         `tfsdk:"status"`
	CreatedAt  types.Int64          `tfsdk:"created_at"`
	AcceptedAt types.Int64          `tfsdk:"accepted_at"`
	ExpiresAt  types.Int64          `tfsdk:"expires_at"`
}

func NewOrganizationInviteResource() resource.Resource { return &organizationInviteResource{} }

func (r *organizationInviteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invite"
}

func (r *organizationInviteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization invite. OpenAI supports create, read, list, and delete for invites; email, role, and project grants are create-only and changes replace the invite.",
		Attributes: map[string]resourceschema.Attribute{
			"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI invite ID."},
			"email":      resourceschema.StringAttribute{Required: true, MarkdownDescription: "Email address to invite.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"role":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "Organization role for the invite: reader or owner.", Validators: []validator.String{newStringEnumValidator(inviteRoles...)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"status":     resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Invite status returned by OpenAI: pending, accepted, or expired."},
			"created_at": resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the invite was sent."},
			"accepted_at": resourceschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp when the invite was accepted, when available.",
			},
			"expires_at": resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the invite expires, when available."},
			"projects": resourceschema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Project memberships granted when the invite is accepted. Omit to use OpenAI's default-project compatibility behavior, or set an empty list to grant no project access.",
				PlanModifiers:       []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: resourceschema.NestedAttributeObject{Attributes: map[string]resourceschema.Attribute{
					"id":   resourceschema.StringAttribute{Required: true, MarkdownDescription: "Project public ID."},
					"role": resourceschema.StringAttribute{Required: true, MarkdownDescription: "Project role granted on invite acceptance: member or owner.", Validators: []validator.String{newStringEnumValidator(projectMemberRoles...)}},
				}},
			},
		},
	}
}

func (r *organizationInviteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationInviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationInviteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	invite, err := r.client.CreateInvite(ctx, client.InviteCreateRequest{Email: plan.Email.ValueString(), Role: plan.Role.ValueString(), Projects: inviteProjectsFromModel(plan.Projects)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI organization invite", err)
		return
	}
	state := organizationInviteResourceModelFromAPI(invite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationInviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationInviteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	invite, err := r.client.GetInvite(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization invite", err)
		return
	}
	newState := organizationInviteResourceModelFromAPI(invite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationInviteResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI organization invites are immutable", "OpenAI does not expose an update operation for organization invites. Change email, role, or projects by replacing the resource.")
}

func (r *organizationInviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationInviteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteInvite(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI organization invite", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationInviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func organizationInviteResourceModelFromAPI(invite *client.Invite) organizationInviteResourceModel {
	return organizationInviteResourceModel{
		ID:         types.StringValue(invite.ID),
		Email:      stringOrNull(invite.Email),
		Role:       stringOrNull(invite.Role),
		Projects:   inviteProjectModelsFromAPI(invite.Projects),
		Status:     stringOrNull(invite.Status),
		CreatedAt:  int64OrNull(invite.CreatedAt),
		AcceptedAt: int64OrNull(invite.AcceptedAt),
		ExpiresAt:  int64OrNull(invite.ExpiresAt),
	}
}
