package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &projectRoleDataSource{}
)

type projectRoleDataSource struct{ client client.AdminClient }

type projectRoleDataSourceModel struct {
	ProjectID      types.String `tfsdk:"project_id"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Permissions    types.Set    `tfsdk:"permissions"`
	PredefinedRole types.Bool   `tfsdk:"predefined_role"`
	ResourceType   types.String `tfsdk:"resource_type"`
}

func NewProjectRoleDataSource() datasource.DataSource { return &projectRoleDataSource{} }

func (d *projectRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_role"
}

func (d *projectRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := roleDataSourceAttributes(true)
	attrs["project_id"] = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the role."}
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI project role by ID.", Attributes: attrs}
}

func (d *projectRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := d.client.GetProjectRole(ctx, config.ProjectID.ValueString(), config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project role", err)
		return
	}
	item, diags := roleItemModelFromAPI(ctx, *role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := projectRoleDataSourceModel{ProjectID: config.ProjectID, ID: item.ID, Name: item.Name, Description: item.Description, Permissions: item.Permissions, PredefinedRole: item.PredefinedRole, ResourceType: item.ResourceType}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
