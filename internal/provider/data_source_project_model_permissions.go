package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectModelPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectModelPermissionsDataSource{}
)

type projectModelPermissionsDataSource struct{ client client.AdminClient }
type projectModelPermissionsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Mode      types.String `tfsdk:"mode"`
	ModelIDs  types.Set    `tfsdk:"model_ids"`
}

func NewProjectModelPermissionsDataSource() datasource.DataSource {
	return &projectModelPermissionsDataSource{}
}
func (d *projectModelPermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_model_permissions"
}
func (d *projectModelPermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves current OpenAI project model permissions.", Attributes: map[string]datasourceschema.Attribute{"id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."}, "project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}, "mode": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Whether the listed models are allowed or denied."}, "model_ids": datasourceschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Enabled model IDs."}}}
}
func (d *projectModelPermissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectModelPermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectModelPermissionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, err := d.client.GetProjectModelPermissions(ctx, config.ProjectID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project model permissions", err)
		return
	}
	state, diags := projectModelPermissionsDataSourceModelFromAPI(ctx, permissions, config.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func projectModelPermissionsDataSourceModelFromAPI(ctx context.Context, permissions *client.ModelPermissions, projectID types.String) (projectModelPermissionsDataSourceModel, diag.Diagnostics) {
	modelIDs, diags := setStringValue(ctx, permissions.ModelIDs)
	return projectModelPermissionsDataSourceModel{ID: projectID, ProjectID: projectID, Mode: stringOrNull(permissions.Mode), ModelIDs: modelIDs}, diags
}
