package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectGroupRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &projectGroupRolesDataSource{}
)

type projectGroupRolesDataSource struct{ client client.AdminClient }

type projectGroupRolesDataSourceModel struct {
	ProjectID types.String                      `tfsdk:"project_id"`
	GroupID   types.String                      `tfsdk:"group_id"`
	After     types.String                      `tfsdk:"after"`
	Limit     types.Int64                       `tfsdk:"limit"`
	Order     types.String                      `tfsdk:"order"`
	Items     []projectGroupRoleDataSourceModel `tfsdk:"items"`
	HasMore   types.Bool                        `tfsdk:"has_more"`
	Next      types.String                      `tfsdk:"next"`
}

func NewProjectGroupRolesDataSource() datasource.DataSource {
	return &projectGroupRolesDataSource{}
}

func (d *projectGroupRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group_roles"
}

func (d *projectGroupRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleAssignmentDataSourceAttributes("group_id", "OpenAI project group ID.", false)
	attrs["project_id"] = datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID."}
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists project role assignments for an OpenAI project group.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."},
			"group_id":   datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project group ID."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of role assignments to return."},
			"order":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":      datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Group role assignments returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: attrs}},
			"has_more":   datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more assignments are available after this page."},
			"next":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *projectGroupRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectGroupRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectGroupRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectGroupRoles(ctx, config.ProjectID.ValueString(), config.GroupID.ValueString(), client.RoleAssignmentListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project group roles", err)
		return
	}
	state := projectGroupRolesDataSourceModel{ProjectID: config.ProjectID, GroupID: config.GroupID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]projectGroupRoleDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, assignment := range page.Items {
		item, diags := projectGroupRoleResourceModelFromAPI(ctx, &assignment, config.ProjectID.ValueString(), config.GroupID.ValueString())
		resp.Diagnostics.Append(diags...)
		state.Items = append(state.Items, projectGroupRoleDataSourceModel(item))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
