package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupsDataSource{}
)

type organizationGroupsDataSource struct{ client client.AdminClient }

type organizationGroupsDataSourceModel struct {
	After   types.String                       `tfsdk:"after"`
	Limit   types.Int64                        `tfsdk:"limit"`
	Order   types.String                       `tfsdk:"order"`
	Items   []organizationGroupDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                         `tfsdk:"has_more"`
	Next    types.String                       `tfsdk:"next"`
}

func NewOrganizationGroupsDataSource() datasource.DataSource { return &organizationGroupsDataSource{} }

func (d *organizationGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_groups"
}

func (d *organizationGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization groups through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of groups to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Organization groups returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: organizationGroupDataSourceAttributes(false)}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more groups are available after this page."},
			"next":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *organizationGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationGroups(ctx, client.OrganizationGroupListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization groups", err)
		return
	}
	state := organizationGroupsDataSourceModel{After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]organizationGroupDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, group := range page.Items {
		state.Items = append(state.Items, organizationGroupDataSourceModelFromAPI(&group))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
