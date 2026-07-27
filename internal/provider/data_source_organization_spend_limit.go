package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationSpendLimitDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationSpendLimitDataSource{}
)

type organizationSpendLimitDataSource struct{ client client.AdminClient }

type organizationSpendLimitDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	ThresholdAmount   types.Int64  `tfsdk:"threshold_amount"`
	Currency          types.String `tfsdk:"currency"`
	Interval          types.String `tfsdk:"interval"`
	EnforcementStatus types.String `tfsdk:"enforcement_status"`
}

func NewOrganizationSpendLimitDataSource() datasource.DataSource {
	return &organizationSpendLimitDataSource{}
}

func (d *organizationSpendLimitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_limit"
}

func (d *organizationSpendLimitDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Retrieves the current OpenAI organization hard spend limit.",
		Attributes: map[string]datasourceschema.Attribute{
			"id":                 datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID. Always organization."},
			"threshold_amount":   datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Hard spend limit amount in cents."},
			"currency":           datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Currency for threshold_amount."},
			"interval":           datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Spend evaluation interval."},
			"enforcement_status": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI enforcement status for the hard spend limit."},
		},
	}
}

func (d *organizationSpendLimitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationSpendLimitDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	limit, err := d.client.GetOrganizationSpendLimit(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization spend limit", err)
		return
	}
	state := organizationSpendLimitDataSourceModel{ID: types.StringValue(organizationSingletonID), ThresholdAmount: types.Int64Value(limit.ThresholdAmount), Currency: stringOrNull(limit.Currency), Interval: stringOrNull(limit.Interval), EnforcementStatus: stringOrNull(limit.EnforcementStatus)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
