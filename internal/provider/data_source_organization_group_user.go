package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationGroupUserDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupUserDataSource{}
)

type organizationGroupUserDataSource struct{ client client.AdminClient }

type organizationGroupUserDataSourceModel struct {
	GroupID          types.String `tfsdk:"group_id"`
	ID               types.String `tfsdk:"id"`
	UserID           types.String `tfsdk:"user_id"`
	Email            types.String `tfsdk:"email"`
	Name             types.String `tfsdk:"name"`
	IsServiceAccount types.Bool   `tfsdk:"is_service_account"`
	Picture          types.String `tfsdk:"picture"`
	UserType         types.String `tfsdk:"user_type"`
}

func NewOrganizationGroupUserDataSource() datasource.DataSource {
	return &organizationGroupUserDataSource{}
}

func (d *organizationGroupUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_user"
}

func (d *organizationGroupUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := organizationGroupUserDataSourceAttributes(true)
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization group membership by group ID and user ID.", Attributes: attrs}
}

func organizationGroupUserDataSourceAttributes(single bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Composite membership ID in group_id/user_id format."}
	userIDAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI organization user ID."}
	if single {
		userIDAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization user ID."}
	}
	return map[string]datasourceschema.Attribute{
		"group_id":           datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization group ID."},
		"id":                 idAttr,
		"user_id":            userIDAttr,
		"email":              datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User email address."},
		"name":               datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User display name."},
		"is_service_account": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is a service account."},
		"picture":            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User profile picture URL when available."},
		"user_type":          datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User type returned by OpenAI."},
	}
}

func (d *organizationGroupUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationGroupUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := d.client.GetOrganizationGroupUser(ctx, config.GroupID.ValueString(), config.UserID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization group user", err)
		return
	}
	state := organizationGroupUserDataSourceModelFromAPI(user, config.GroupID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func organizationGroupUserDataSourceModelFromAPI(user *client.OrganizationGroupUser, groupID string) organizationGroupUserDataSourceModel {
	return organizationGroupUserDataSourceModel{GroupID: types.StringValue(groupID), ID: types.StringValue(groupID + "/" + user.ID), UserID: types.StringValue(user.ID), Email: stringOrNull(user.Email), Name: stringOrNull(user.Name), IsServiceAccount: types.BoolValue(user.IsServiceAccount), Picture: stringOrNull(user.Picture), UserType: stringOrNull(user.UserType)}
}
