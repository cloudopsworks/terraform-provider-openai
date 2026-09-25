package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectUserRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &projectUserRolesDataSource{}
)

type projectUserRolesDataSource struct{ client client.AdminClient }

type projectUserRolesDataSourceModel struct {
	ProjectID types.String                     `tfsdk:"project_id"`
	UserID    types.String                     `tfsdk:"user_id"`
	After     types.String                     `tfsdk:"after"`
	Limit     types.Int64                      `tfsdk:"limit"`
	Order     types.String                     `tfsdk:"order"`
	Items     []projectUserRoleDataSourceModel `tfsdk:"items"`
	HasMore   types.Bool                       `tfsdk:"has_more"`
	Next      types.String                     `tfsdk:"next"`
}

func NewProjectUserRolesDataSource() datasource.DataSource {
	return &projectUserRolesDataSource{}
}

func (d *projectUserRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user_roles"
}

func (d *projectUserRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleAssignmentDataSourceAttributes("user_id", "OpenAI project user ID.", false)
	attrs["project_id"] = datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID."}
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists project role assignments for an OpenAI project user.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."},
			"user_id":    datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project user ID."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of role assignments to return."},
			"order":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":      datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "User role assignments returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: attrs}},
			"has_more":   datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more assignments are available after this page."},
			"next":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor returned by the API."},
		},
	}
}

func (d *projectUserRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectUserRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectUserRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectUserRoles(ctx, config.ProjectID.ValueString(), config.UserID.ValueString(), client.RoleAssignmentListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project user roles", err)
		return
	}
	state := projectUserRolesDataSourceModel{ProjectID: config.ProjectID, UserID: config.UserID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]projectUserRoleDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for _, assignment := range page.Items {
		item, diags := projectUserRoleResourceModelFromAPI(ctx, &assignment, config.ProjectID.ValueString(), config.UserID.ValueString())
		resp.Diagnostics.Append(diags...)
		state.Items = append(state.Items, projectUserRoleDataSourceModel(item))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
