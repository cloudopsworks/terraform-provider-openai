package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationInvitesDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationInvitesDataSource{}
)

type organizationInvitesDataSource struct{ client client.AdminClient }

type organizationInvitesDataSourceModel struct {
	After   types.String                        `tfsdk:"after"`
	Limit   types.Int64                         `tfsdk:"limit"`
	Items   []organizationInviteDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                          `tfsdk:"has_more"`
	LastID  types.String                        `tfsdk:"last_id"`
}

func NewOrganizationInvitesDataSource() datasource.DataSource {
	return &organizationInvitesDataSource{}
}

func (d *organizationInvitesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_invites"
}

func (d *organizationInvitesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization invites through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of invites to return."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Organization invites returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: organizationInviteDataSourceAttributes(false)}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more invites are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last invite in the page."},
		},
	}
}

func (d *organizationInvitesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationInvitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationInvitesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListInvites(ctx, client.InviteListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization invites", err)
		return
	}
	state := organizationInvitesDataSourceModel{After: config.After, Limit: config.Limit, Items: make([]organizationInviteDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for _, invite := range page.Items {
		state.Items = append(state.Items, organizationInviteDataSourceModelFromAPI(&invite))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
