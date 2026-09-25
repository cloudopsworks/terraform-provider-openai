package provider

import (
	"context"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &projectGroupDataSource{}
)

type projectGroupDataSource struct{ client client.AdminClient }

type projectGroupDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	GroupID   types.String `tfsdk:"group_id"`
	GroupName types.String `tfsdk:"group_name"`
	GroupType types.String `tfsdk:"group_type"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewProjectGroupDataSource() datasource.DataSource { return &projectGroupDataSource{} }
func (d *projectGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_group"
}
func (d *projectGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Retrieves one OpenAI project group membership.",
		Attributes:          projectGroupDataSourceAttributes(true),
	}
}

func projectGroupDataSourceAttributes(required bool) map[string]datasourceschema.Attribute {
	projectID := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID."}
	groupID := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI group ID."}
	groupType := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group type."}
	if required {
		projectID = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}
		groupID = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI group ID."}
		groupType = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "Group type."}
	}
	return map[string]datasourceschema.Attribute{"id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Composite project/group membership ID."}, "project_id": projectID, "group_id": groupID, "group_name": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group display name."}, "group_type": groupType, "created_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when membership was created."}}
}
func (d *projectGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := d.client.GetProjectGroup(ctx, config.ProjectID.ValueString(), config.GroupID.ValueString(), client.ProjectGroupGetRequest{GroupType: config.GroupType.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project group", err)
		return
	}
	state := projectGroupDataSourceModelFromAPI(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func projectGroupDataSourceModelFromAPI(group *client.ProjectGroup) projectGroupDataSourceModel {
	return projectGroupDataSourceModel{ID: types.StringValue(group.ProjectID + "/" + group.GroupID), ProjectID: types.StringValue(group.ProjectID), GroupID: types.StringValue(group.GroupID), GroupName: stringOrNull(group.GroupName), GroupType: stringOrNull(group.GroupType), CreatedAt: int64OrNull(group.CreatedAt)}
}
