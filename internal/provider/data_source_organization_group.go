package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationGroupDataSource{}
)

type organizationGroupDataSource struct{ client client.AdminClient }

type organizationGroupDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	GroupType     types.String `tfsdk:"group_type"`
	IsScimManaged types.Bool   `tfsdk:"is_scim_managed"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
}

func NewOrganizationGroupDataSource() datasource.DataSource { return &organizationGroupDataSource{} }

func (d *organizationGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group"
}

func (d *organizationGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization group by ID.", Attributes: organizationGroupDataSourceAttributes(true)}
}

func organizationGroupDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI organization group ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization group ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":              idAttr,
		"name":            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group display name."},
		"group_type":      datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group type returned by OpenAI."},
		"is_scim_managed": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the group is managed through SCIM."},
		"created_at":      datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for group creation."},
	}
}

func (d *organizationGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := d.client.GetOrganizationGroup(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization group", err)
		return
	}
	state := organizationGroupDataSourceModelFromAPI(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func organizationGroupDataSourceModelFromAPI(group *client.OrganizationGroup) organizationGroupDataSourceModel {
	return organizationGroupDataSourceModel{ID: types.StringValue(group.ID), Name: stringOrNull(group.Name), GroupType: stringOrNull(group.GroupType), IsScimManaged: types.BoolValue(group.IsScimManaged), CreatedAt: int64OrNull(group.CreatedAt)}
}
