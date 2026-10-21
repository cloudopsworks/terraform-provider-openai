package provider

import (
	"context"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectGroupsDataSource{}
)

type projectGroupsDataSource struct{ client client.AdminClient }

type projectGroupsDataSourceModel struct {
	ProjectID types.String                  `tfsdk:"project_id"`
	After     types.String                  `tfsdk:"after"`
	Limit     types.Int64                   `tfsdk:"limit"`
	Order     types.String                  `tfsdk:"order"`
	Items     []projectGroupDataSourceModel `tfsdk:"items"`
	HasMore   types.Bool                    `tfsdk:"has_more"`
	Next      types.String                  `tfsdk:"next"`
}

func NewProjectGroupsDataSource() datasource.DataSource { return &projectGroupsDataSource{} }
func (d *projectGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_groups"
}
func (d *projectGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI project group memberships.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of group memberships."},
			"order":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order: asc or desc."},
			"items":      datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Project group memberships.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: projectGroupDataSourceAttributes(false)}},
			"has_more":   datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more memberships are available."},
			"next":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Next cursor."},
		},
	}
}

func (d *projectGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectGroups(ctx, config.ProjectID.ValueString(), client.ProjectGroupListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project groups", err)
		return
	}
	state := projectGroupsDataSourceModel{ProjectID: config.ProjectID, After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]projectGroupDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), Next: stringOrNull(page.Next)}
	for i := range page.Items {
		state.Items = append(state.Items, projectGroupDataSourceModelFromAPI(&page.Items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
