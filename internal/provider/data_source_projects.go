package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectsDataSource{}
)

type projectsDataSource struct {
	client client.AdminClient
}

type projectsDataSourceModel struct {
	After           types.String           `tfsdk:"after"`
	Limit           types.Int64            `tfsdk:"limit"`
	IncludeArchived types.Bool             `tfsdk:"include_archived"`
	Items           []projectListItemModel `tfsdk:"items"`
	HasMore         types.Bool             `tfsdk:"has_more"`
	LastID          types.String           `tfsdk:"last_id"`
}

type projectListItemModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ExternalKeyID types.String `tfsdk:"external_key_id"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	ArchivedAt    types.Int64  `tfsdk:"archived_at"`
}

func NewProjectsDataSource() datasource.DataSource { return &projectsDataSource{} }

func (d *projectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization projects through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":            datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination. Use the last_id from a previous response to fetch the next page."},
			"limit":            datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of projects to return. The OpenAI API supports values from 1 through 100."},
			"include_archived": datasourceschema.BoolAttribute{Optional: true, MarkdownDescription: "Include archived projects. Archived projects are excluded by default."},
			"items": datasourceschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Normalized project collection returned by the API.",
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: map[string]datasourceschema.Attribute{
						"id":              datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID."},
						"name":            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project display name."},
						"external_key_id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "External key ID associated with the project, when present."},
						"status":          datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project status, such as active or archived."},
						"created_at":      datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for project creation."},
						"archived_at":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for project archival when archived."},
					},
				},
			},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more projects are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last project in the returned page, for use as the after cursor."},
		},
	}
}

func (d *projectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := d.client.ListProjects(ctx, client.ProjectListRequest{
		After:           stringValue(config.After),
		Limit:           int64Value(config.Limit),
		IncludeArchived: boolValue(config.IncludeArchived),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI projects", err)
		return
	}

	state := projectsDataSourceModel{
		After:           config.After,
		Limit:           config.Limit,
		IncludeArchived: config.IncludeArchived,
		Items:           make([]projectListItemModel, 0, len(page.Items)),
		HasMore:         types.BoolValue(page.HasMore),
		LastID:          stringOrNull(page.LastID),
	}
	for _, project := range page.Items {
		state.Items = append(state.Items, projectListItemModel{
			ID:            types.StringValue(project.ID),
			Name:          stringOrNull(project.Name),
			ExternalKeyID: stringOrNull(project.ExternalKeyID),
			Status:        stringOrNull(project.Status),
			CreatedAt:     int64OrNull(project.CreatedAt),
			ArchivedAt:    int64OrNull(project.ArchivedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
