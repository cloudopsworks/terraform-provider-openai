package provider

import (
	"context"
	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectSpendLimitDataSource{}
	_ datasource.DataSourceWithConfigure = &projectSpendLimitDataSource{}
)

type projectSpendLimitDataSource struct{ client client.AdminClient }
type projectSpendLimitDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	ThresholdAmount   types.Int64  `tfsdk:"threshold_amount"`
	Currency          types.String `tfsdk:"currency"`
	Interval          types.String `tfsdk:"interval"`
	EnforcementStatus types.String `tfsdk:"enforcement_status"`
}

func NewProjectSpendLimitDataSource() datasource.DataSource { return &projectSpendLimitDataSource{} }
func (d *projectSpendLimitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_limit"
}
func (d *projectSpendLimitDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves the current OpenAI project hard spend limit.", Attributes: map[string]datasourceschema.Attribute{"id": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."}, "project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID."}, "threshold_amount": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Hard spend limit amount in cents."}, "currency": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Currency for threshold_amount."}, "interval": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Spend evaluation interval."}, "enforcement_status": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI enforcement status for the hard spend limit."}}}
}
func (d *projectSpendLimitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectSpendLimitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectSpendLimitDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := d.client.GetProjectSpendLimit(ctx, config.ProjectID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project spend limit", err)
		return
	}
	state := projectSpendLimitDataSourceModel{ID: config.ProjectID, ProjectID: config.ProjectID, ThresholdAmount: types.Int64Value(limit.ThresholdAmount), Currency: stringOrNull(limit.Currency), Interval: stringOrNull(limit.Interval), EnforcementStatus: stringOrNull(limit.EnforcementStatus)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
