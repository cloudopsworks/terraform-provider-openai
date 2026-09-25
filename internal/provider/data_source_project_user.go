package provider

import (
	"context"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectUserDataSource{}
	_ datasource.DataSourceWithConfigure = &projectUserDataSource{}
)

type projectUserDataSource struct{ client client.AdminClient }

type projectUserDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	ID        types.String `tfsdk:"id"`
	UserID    types.String `tfsdk:"user_id"`
	Email     types.String `tfsdk:"email"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	AddedAt   types.Int64  `tfsdk:"added_at"`
}

func NewProjectUserDataSource() datasource.DataSource { return &projectUserDataSource{} }
func (d *projectUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}
func (d *projectUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Retrieves one OpenAI project user by ID.",
		Attributes:          projectUserDataSourceAttributes(true),
	}
}

func projectUserDataSourceAttributes(required bool) map[string]datasourceschema.Attribute {
	id := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI user ID."}
	projectID := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID."}
	if required {
		id = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI user ID."}
		projectID = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}
	}
	return map[string]datasourceschema.Attribute{"id": id, "project_id": projectID, "user_id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI user ID."}, "email": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User email address."}, "name": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User display name."}, "role": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project role."}, "added_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when access was granted."}}
}
func (d *projectUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := d.client.GetProjectUser(ctx, config.ProjectID.ValueString(), config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project user", err)
		return
	}
	state := projectUserDataSourceModelFromAPI(user, config.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func projectUserDataSourceModelFromAPI(user *client.ProjectUser, projectID types.String) projectUserDataSourceModel {
	return projectUserDataSourceModel{ProjectID: projectID, ID: types.StringValue(user.ID), UserID: types.StringValue(user.ID), Email: stringOrNull(user.Email), Name: stringOrNull(user.Name), Role: stringOrNull(user.Role), AddedAt: int64OrNull(user.AddedAt)}
}
