package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationUserRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationUserRolesDataSource{}
)

type organizationUserRolesDataSource struct{ client client.AdminClient }

type organizationUserRolesDataSourceModel struct {
	UserID  types.String                          `tfsdk:"user_id"`
	After   types.String                          `tfsdk:"after"`
	Limit   types.Int64                           `tfsdk:"limit"`
	Order   types.String                          `tfsdk:"order"`
	Items   []organizationUserRoleDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                            `tfsdk:"has_more"`
	Next    types.String                          `tfsdk:"next"`
}

func NewOrganizationUserRolesDataSource() datasource.DataSource {
	return &organizationUserRolesDataSource{}
}

func (d *organizationUserRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_roles"
}

func (d *organizationUserRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleAssignmentDataSourceAttributes("user_id", "OpenAI organization user ID.", false)
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists organization role assignments for an OpenAI organization user.",
		Attributes: map[string]datasourceschema.Attribute{
			"user_id":  datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization user ID."},
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of role assignments to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "User role assignments returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: attrs}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more assignments are available after this page."},
			"next":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *organizationUserRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationUserRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationUserRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationUserRoles(ctx, config.UserID.ValueString(), client.RoleAssignmentListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization user roles", err)
		return
	}
	state := organizationUserRolesDataSourceModel{UserID: config.UserID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]organizationUserRoleDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, assignment := range page.Items {
		item, diags := organizationUserRoleResourceModelFromAPI(ctx, &assignment, config.UserID.ValueString())
		resp.Diagnostics.Append(diags...)
		state.Items = append(state.Items, organizationUserRoleDataSourceModel(item))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
