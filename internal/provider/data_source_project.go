package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

type projectDataSource struct {
	client client.AdminClient
}

type projectDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ExternalKeyID types.String `tfsdk:"external_key_id"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	ArchivedAt    types.Int64  `tfsdk:"archived_at"`
}

func NewProjectDataSource() datasource.DataSource { return &projectDataSource{} }

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Looks up an OpenAI organization project by ID through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"id":              datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID. Lookup uses the project retrieve endpoint."},
			"name":            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project display name."},
			"external_key_id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "External key ID associated with the project, when present."},
			"status":          datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project status, such as active or archived."},
			"created_at":      datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for project creation."},
			"archived_at":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for project archival when archived."},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if stringValue(config.ID) == "" {
		resp.Diagnostics.AddError("Missing project lookup value", "Set id to look up an OpenAI project.")
		return
	}

	project, err := d.client.GetProject(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project", err)
		return
	}

	state := projectDataSourceModel{
		ID:            types.StringValue(project.ID),
		Name:          stringOrNull(project.Name),
		ExternalKeyID: stringOrNull(project.ExternalKeyID),
		Status:        stringOrNull(project.Status),
		CreatedAt:     int64OrNull(project.CreatedAt),
		ArchivedAt:    int64OrNull(project.ArchivedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
