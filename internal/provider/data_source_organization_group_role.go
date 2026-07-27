package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationGroupRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupRoleDataSource{}
)

type organizationGroupRoleDataSource struct{ client client.AdminClient }

type organizationGroupRoleDataSourceModel organizationGroupRoleResourceModel

func NewOrganizationGroupRoleDataSource() datasource.DataSource {
	return &organizationGroupRoleDataSource{}
}

func (d *organizationGroupRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group_role"
}

func (d *organizationGroupRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization group role assignment by group ID and role ID.", Attributes: roleAssignmentDataSourceAttributes("group_id", "OpenAI organization group ID.", true)}
}

func (d *organizationGroupRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationGroupRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := d.client.GetOrganizationGroupRole(ctx, config.GroupID.ValueString(), config.RoleID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization group role", err)
		return
	}
	state, diags := organizationGroupRoleResourceModelFromAPI(ctx, assignment, config.GroupID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dsState := organizationGroupRoleDataSourceModel(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &dsState)...)
}
