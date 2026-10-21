package provider

import (
	"context"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &projectUsersDataSource{}
)

type projectUsersDataSource struct{ client client.AdminClient }

type projectUsersDataSourceModel struct {
	ProjectID types.String                 `tfsdk:"project_id"`
	After     types.String                 `tfsdk:"after"`
	Limit     types.Int64                  `tfsdk:"limit"`
	Items     []projectUserDataSourceModel `tfsdk:"items"`
	HasMore   types.Bool                   `tfsdk:"has_more"`
	LastID    types.String                 `tfsdk:"last_id"`
}

func NewProjectUsersDataSource() datasource.DataSource { return &projectUsersDataSource{} }
func (d *projectUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_users"
}
func (d *projectUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI project users.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of users."},
			"items":      datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Project users.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: projectUserDataSourceAttributes(false)}},
			"has_more":   datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more users are available."},
			"last_id":    datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Last user ID."},
		},
	}
}

func (d *projectUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectUsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectUsers(ctx, config.ProjectID.ValueString(), client.ProjectUserListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project users", err)
		return
	}
	state := projectUsersDataSourceModel{ProjectID: config.ProjectID, After: config.After, Limit: config.Limit, Items: make([]projectUserDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for i := range page.Items {
		state.Items = append(state.Items, projectUserDataSourceModelFromAPI(&page.Items[i], config.ProjectID))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
