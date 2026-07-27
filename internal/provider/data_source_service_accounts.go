package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &serviceAccountsDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceAccountsDataSource{}
)

type serviceAccountsDataSource struct {
	client client.AdminClient
}

type serviceAccountsDataSourceModel struct {
	ProjectID types.String                  `tfsdk:"project_id"`
	After     types.String                  `tfsdk:"after"`
	Limit     types.Int64                   `tfsdk:"limit"`
	Items     []serviceAccountListItemModel `tfsdk:"items"`
	HasMore   types.Bool                    `tfsdk:"has_more"`
	LastID    types.String                  `tfsdk:"last_id"`
}

type serviceAccountListItemModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewServiceAccountsDataSource() datasource.DataSource { return &serviceAccountsDataSource{} }

func (d *serviceAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (d *serviceAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI project service accounts through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the service accounts."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination. Use the last_id from a previous response to fetch the next page."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of service accounts to return. The OpenAI API supports values from 1 through 100."},
			"items": datasourceschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Normalized service account collection returned by the API.",
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: map[string]datasourceschema.Attribute{
						"id":         datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI service account ID."},
						"name":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Service account name."},
						"role":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project role for the service account."},
						"created_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for service-account creation."},
					},
				},
			},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more service accounts are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last service account in the returned page, for use as the after cursor."},
		},
	}
}

func (d *serviceAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceAccountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := d.client.ListServiceAccounts(ctx, config.ProjectID.ValueString(), client.ServiceAccountListRequest{
		After: stringValue(config.After),
		Limit: int64Value(config.Limit),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI service accounts", err)
		return
	}

	state := serviceAccountsDataSourceModel{
		ProjectID: config.ProjectID,
		After:     config.After,
		Limit:     config.Limit,
		Items:     make([]serviceAccountListItemModel, 0, len(page.Items)),
		HasMore:   types.BoolValue(page.HasMore),
		LastID:    stringOrNull(page.LastID),
	}
	for _, account := range page.Items {
		state.Items = append(state.Items, serviceAccountListItemModel{
			ID:        types.StringValue(account.ID),
			Name:      stringOrNull(account.Name),
			Role:      stringOrNull(account.Role),
			CreatedAt: int64OrNull(account.CreatedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
