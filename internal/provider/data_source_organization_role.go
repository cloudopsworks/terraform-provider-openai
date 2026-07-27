package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationRoleDataSource{}
)

type organizationRoleDataSource struct{ client client.AdminClient }

type organizationRoleDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Permissions    types.Set    `tfsdk:"permissions"`
	PredefinedRole types.Bool   `tfsdk:"predefined_role"`
	ResourceType   types.String `tfsdk:"resource_type"`
}

func NewOrganizationRoleDataSource() datasource.DataSource { return &organizationRoleDataSource{} }

func (d *organizationRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_role"
}

func (d *organizationRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization role by ID.", Attributes: roleDataSourceAttributes(true)}
}

func roleDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI role ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI role ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":              idAttr,
		"name":            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Role name."},
		"description":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Role description."},
		"permissions":     datasourceschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Permissions granted by the role."},
		"predefined_role": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OpenAI predefines and manages the role."},
		"resource_type":   datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Resource type the role applies to."},
	}
}

func (d *organizationRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := d.client.GetOrganizationRole(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization role", err)
		return
	}
	item, diags := roleItemModelFromAPI(ctx, *role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := organizationRoleDataSourceModel(item)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
