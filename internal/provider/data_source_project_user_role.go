package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectUserRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &projectUserRoleDataSource{}
)

type projectUserRoleDataSource struct{ client client.AdminClient }

type projectUserRoleDataSourceModel projectUserRoleResourceModel

func NewProjectUserRoleDataSource() datasource.DataSource {
	return &projectUserRoleDataSource{}
}

func (d *projectUserRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user_role"
}

func (d *projectUserRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleAssignmentDataSourceAttributes("user_id", "OpenAI project user ID.", true)
	attrs["project_id"] = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI project user role assignment by project ID, user ID, and role ID.", Attributes: attrs}
}

func (d *projectUserRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectUserRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectUserRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := d.client.GetProjectUserRole(ctx, config.ProjectID.ValueString(), config.UserID.ValueString(), config.RoleID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project user role", err)
		return
	}
	state, diags := projectUserRoleResourceModelFromAPI(ctx, assignment, config.ProjectID.ValueString(), config.UserID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dsState := projectUserRoleDataSourceModel(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &dsState)...)
}
