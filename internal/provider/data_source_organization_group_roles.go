package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationGroupRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupRolesDataSource{}
)

type organizationGroupRolesDataSource struct{ client client.AdminClient }

type organizationGroupRolesDataSourceModel struct {
	GroupID types.String                           `tfsdk:"group_id"`
	After   types.String                           `tfsdk:"after"`
	Limit   types.Int64                            `tfsdk:"limit"`
	Order   types.String                           `tfsdk:"order"`
	Items   []organizationGroupRoleDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                             `tfsdk:"has_more"`
	Next    types.String                           `tfsdk:"next"`
}

func NewOrganizationGroupRolesDataSource() datasource.DataSource {
	return &organizationGroupRolesDataSource{}
}

func (d *organizationGroupRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_roles"
}

func (d *organizationGroupRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleAssignmentDataSourceAttributes("group_id", "OpenAI organization group ID.", false)
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists organization role assignments for an OpenAI organization group.",
		Attributes: map[string]datasourceschema.Attribute{
			"group_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization group ID."},
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of role assignments to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Group role assignments returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: attrs}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more assignments are available after this page."},
			"next":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *organizationGroupRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationGroupRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationGroupRoles(ctx, config.GroupID.ValueString(), client.RoleAssignmentListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization group roles", err)
		return
	}
	state := organizationGroupRolesDataSourceModel{GroupID: config.GroupID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]organizationGroupRoleDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, assignment := range page.Items {
		item, diags := organizationGroupRoleResourceModelFromAPI(ctx, &assignment, config.GroupID.ValueString())
		resp.Diagnostics.Append(diags...)
		state.Items = append(state.Items, organizationGroupRoleDataSourceModel(item))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
