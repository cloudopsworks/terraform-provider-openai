package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectDataRetentionDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataRetentionDataSource{}
)

type projectDataRetentionDataSource struct{ client client.AdminClient }
type projectDataRetentionDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Type      types.String `tfsdk:"type"`
}

func NewProjectDataRetentionDataSource() datasource.DataSource {
	return &projectDataRetentionDataSource{}
}
func (d *projectDataRetentionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_data_retention"
}
func (d *projectDataRetentionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves current OpenAI project data retention controls.", Attributes: map[string]datasourceschema.Attribute{"id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."}, "project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}, "type": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project data retention type."}}}
}
func (d *projectDataRetentionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectDataRetentionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectDataRetentionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retention, err := d.client.GetProjectDataRetention(ctx, config.ProjectID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project data retention", err)
		return
	}
	state := projectDataRetentionDataSourceModel{ID: config.ProjectID, ProjectID: config.ProjectID, Type: stringOrNull(retention.Type)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
