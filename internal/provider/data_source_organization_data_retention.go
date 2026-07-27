package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationDataRetentionDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataRetentionDataSource{}
)

type organizationDataRetentionDataSource struct{ client client.AdminClient }

type organizationDataRetentionDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func NewOrganizationDataRetentionDataSource() datasource.DataSource {
	return &organizationDataRetentionDataSource{}
}

func (d *organizationDataRetentionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_data_retention"
}

func (d *organizationDataRetentionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Retrieves current OpenAI organization data retention controls.",
		Attributes: map[string]datasourceschema.Attribute{
			"id":   datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID. Always organization."},
			"type": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Organization data retention type."},
		},
	}
}

func (d *organizationDataRetentionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationDataRetentionDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	retention, err := d.client.GetOrganizationDataRetention(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization data retention", err)
		return
	}
	state := organizationDataRetentionDataSourceModel{ID: types.StringValue(organizationSingletonID), Type: stringOrNull(retention.Type)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
