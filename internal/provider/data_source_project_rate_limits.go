package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectRateLimitsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectRateLimitsDataSource{}
)

type projectRateLimitsDataSource struct{ client client.AdminClient }
type projectRateLimitsDataSourceModel struct {
	ProjectID types.String                      `tfsdk:"project_id"`
	After     types.String                      `tfsdk:"after"`
	Before    types.String                      `tfsdk:"before"`
	Limit     types.Int64                       `tfsdk:"limit"`
	Items     []projectRateLimitDataSourceModel `tfsdk:"items"`
	HasMore   types.Bool                        `tfsdk:"has_more"`
	LastID    types.String                      `tfsdk:"last_id"`
}

func NewProjectRateLimitsDataSource() datasource.DataSource { return &projectRateLimitsDataSource{} }
func (d *projectRateLimitsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_rate_limits"
}
func (d *projectRateLimitsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Lists OpenAI project rate limits through the Administration API.", Attributes: map[string]datasourceschema.Attribute{
		"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the rate limits."}, "after": datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for the next page."}, "before": datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for the previous page."}, "limit": datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of rate limits to return."}, "items": datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Rate limits returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: projectRateLimitDataSourceAttributes(false)}}, "has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more rate limits are available after this page."}, "last_id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last rate limit in the page."}}}
}
func (d *projectRateLimitsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectRateLimitsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectRateLimitsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectRateLimits(ctx, config.ProjectID.ValueString(), client.RateLimitListRequest{After: stringValue(config.After), Before: stringValue(config.Before), Limit: int64Value(config.Limit)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project rate limits", err)
		return
	}
	state := projectRateLimitsDataSourceModel{ProjectID: config.ProjectID, After: config.After, Before: config.Before, Limit: config.Limit, Items: make([]projectRateLimitDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for i := range page.Items {
		state.Items = append(state.Items, projectRateLimitDataSourceModelFromAPI(&page.Items[i], config.ProjectID))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
