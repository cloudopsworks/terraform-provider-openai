package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectGroupRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &projectGroupRoleDataSource{}
)

type projectGroupRoleDataSource struct{ client client.AdminClient }

type projectGroupRoleDataSourceModel projectGroupRoleResourceModel

func NewProjectGroupRoleDataSource() datasource.DataSource {
	return &projectGroupRoleDataSource{}
}

func (d *projectGroupRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group_role"
}

func (d *projectGroupRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleAssignmentDataSourceAttributes("group_id", "OpenAI project group ID.", true)
	attrs["project_id"] = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI project group role assignment by project ID, group ID, and role ID.", Attributes: attrs}
}

func (d *projectGroupRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectGroupRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectGroupRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := d.client.GetProjectGroupRole(ctx, config.ProjectID.ValueString(), config.GroupID.ValueString(), config.RoleID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project group role", err)
		return
	}
	state, diags := projectGroupRoleResourceModelFromAPI(ctx, assignment, config.ProjectID.ValueString(), config.GroupID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dsState := projectGroupRoleDataSourceModel(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &dsState)...)
}
