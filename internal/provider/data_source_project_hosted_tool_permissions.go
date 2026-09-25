package provider

import (
	"context"
	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectHostedToolPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectHostedToolPermissionsDataSource{}
)

type projectHostedToolPermissionsDataSource struct{ client client.AdminClient }
type projectHostedToolPermissionsDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	CodeInterpreter types.Bool   `tfsdk:"code_interpreter"`
	FileSearch      types.Bool   `tfsdk:"file_search"`
	ImageGeneration types.Bool   `tfsdk:"image_generation"`
	Mcp             types.Bool   `tfsdk:"mcp"`
	WebSearch       types.Bool   `tfsdk:"web_search"`
}

func NewProjectHostedToolPermissionsDataSource() datasource.DataSource {
	return &projectHostedToolPermissionsDataSource{}
}
func (d *projectHostedToolPermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_hosted_tool_permissions"
}
func (d *projectHostedToolPermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves current OpenAI project hosted-tool permissions.", Attributes: map[string]datasourceschema.Attribute{"id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."}, "project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}, "code_interpreter": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether Code Interpreter is enabled."}, "file_search": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether File Search is enabled."}, "image_generation": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether Image Generation is enabled."}, "mcp": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether MCP is enabled."}, "web_search": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether Web Search is enabled."}}}
}
func (d *projectHostedToolPermissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectHostedToolPermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectHostedToolPermissionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, err := d.client.GetProjectHostedToolPermissions(ctx, config.ProjectID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project hosted-tool permissions", err)
		return
	}
	state := projectHostedToolPermissionsDataSourceModel{ID: config.ProjectID, ProjectID: config.ProjectID, CodeInterpreter: types.BoolValue(permissions.CodeInterpreter), FileSearch: types.BoolValue(permissions.FileSearch), ImageGeneration: types.BoolValue(permissions.ImageGeneration), Mcp: types.BoolValue(permissions.Mcp), WebSearch: types.BoolValue(permissions.WebSearch)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
