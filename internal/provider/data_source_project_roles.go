package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &projectRolesDataSource{}
)

type projectRolesDataSource struct{ client client.AdminClient }

type projectRolesDataSourceModel struct {
	ProjectID types.String    `tfsdk:"project_id"`
	After     types.String    `tfsdk:"after"`
	Limit     types.Int64     `tfsdk:"limit"`
	Order     types.String    `tfsdk:"order"`
	Items     []roleItemModel `tfsdk:"items"`
	HasMore   types.Bool      `tfsdk:"has_more"`
	Next      types.String    `tfsdk:"next"`
}

func NewProjectRolesDataSource() datasource.DataSource { return &projectRolesDataSource{} }

func (d *projectRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_roles"
}

func (d *projectRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI project roles through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the roles."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of roles to return."},
			"order":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":      datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Project roles returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: roleDataSourceAttributes(false)}},
			"has_more":   datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more roles are available after this page."},
			"next":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *projectRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectRoles(ctx, config.ProjectID.ValueString(), client.RoleListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project roles", err)
		return
	}
	state := projectRolesDataSourceModel{ProjectID: config.ProjectID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]roleItemModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, role := range page.Items {
		item, diags := roleItemModelFromAPI(ctx, role)
		resp.Diagnostics.Append(diags...)
		state.Items = append(state.Items, item)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
