package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationGroupUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupUsersDataSource{}
)

type organizationGroupUsersDataSource struct{ client client.AdminClient }

type organizationGroupUsersDataSourceModel struct {
	GroupID types.String                           `tfsdk:"group_id"`
	After   types.String                           `tfsdk:"after"`
	Limit   types.Int64                            `tfsdk:"limit"`
	Order   types.String                           `tfsdk:"order"`
	Items   []organizationGroupUserDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                             `tfsdk:"has_more"`
	Next    types.String                           `tfsdk:"next"`
}

func NewOrganizationGroupUsersDataSource() datasource.DataSource {
	return &organizationGroupUsersDataSource{}
}

func (d *organizationGroupUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_users"
}

func (d *organizationGroupUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := organizationGroupUserDataSourceAttributes(false)
	attrs["group_id"] = datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI organization group ID."}
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists users in an OpenAI organization group.",
		Attributes: map[string]datasourceschema.Attribute{
			"group_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization group ID."},
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of users to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Group users returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: attrs}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more users are available after this page."},
			"next":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *organizationGroupUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationGroupUsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationGroupUsers(ctx, config.GroupID.ValueString(), client.OrganizationGroupUserListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization group users", err)
		return
	}
	state := organizationGroupUsersDataSourceModel{GroupID: config.GroupID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]organizationGroupUserDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, user := range page.Items {
		state.Items = append(state.Items, organizationGroupUserDataSourceModelFromAPI(&user, config.GroupID.ValueString()))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
