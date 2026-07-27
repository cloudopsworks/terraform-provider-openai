package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationInviteDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationInviteDataSource{}
)

type organizationInviteDataSource struct{ client client.AdminClient }

type organizationInviteDataSourceModel struct {
	ID         types.String         `tfsdk:"id"`
	Email      types.String         `tfsdk:"email"`
	Role       types.String         `tfsdk:"role"`
	Projects   []inviteProjectModel `tfsdk:"projects"`
	Status     types.String         `tfsdk:"status"`
	CreatedAt  types.Int64          `tfsdk:"created_at"`
	AcceptedAt types.Int64          `tfsdk:"accepted_at"`
	ExpiresAt  types.Int64          `tfsdk:"expires_at"`
}

func NewOrganizationInviteDataSource() datasource.DataSource { return &organizationInviteDataSource{} }

func (d *organizationInviteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invite"
}

func (d *organizationInviteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization invite by ID.", Attributes: organizationInviteDataSourceAttributes(true)}
}

func organizationInviteDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI invite ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI invite ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":          idAttr,
		"email":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Email address that received the invite."},
		"role":        datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Organization role granted by the invite."},
		"status":      datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Invite status returned by OpenAI."},
		"created_at":  datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the invite was sent."},
		"accepted_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the invite was accepted, when available."},
		"expires_at":  datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the invite expires, when available."},
		"projects": datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Project memberships granted when the invite is accepted.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
			"id":   datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project public ID."},
			"role": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project role granted by the invite."},
		}}},
	}
}

func (d *organizationInviteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, err := configureClient(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("OpenAI provider not configured", err.Error())
		return
	}
	d.client = data.client
}

func (d *organizationInviteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationInviteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	invite, err := d.client.GetInvite(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization invite", err)
		return
	}
	state := organizationInviteDataSourceModelFromAPI(invite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func organizationInviteDataSourceModelFromAPI(invite *client.Invite) organizationInviteDataSourceModel {
	return organizationInviteDataSourceModel{
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
