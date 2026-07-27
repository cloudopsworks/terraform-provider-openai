package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationUserRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationUserRoleDataSource{}
)

type organizationUserRoleDataSource struct{ client client.AdminClient }

type organizationUserRoleDataSourceModel organizationUserRoleResourceModel

func NewOrganizationUserRoleDataSource() datasource.DataSource {
	return &organizationUserRoleDataSource{}
}

func (d *organizationUserRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user_role"
}

func (d *organizationUserRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization user role assignment by user ID and role ID.", Attributes: roleAssignmentDataSourceAttributes("user_id", "OpenAI organization user ID.", true)}
}

func (d *organizationUserRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationUserRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationUserRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := d.client.GetOrganizationUserRole(ctx, config.UserID.ValueString(), config.RoleID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization user role", err)
		return
	}
	state, diags := organizationUserRoleResourceModelFromAPI(ctx, assignment, config.UserID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dsState := organizationUserRoleDataSourceModel(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &dsState)...)
}
