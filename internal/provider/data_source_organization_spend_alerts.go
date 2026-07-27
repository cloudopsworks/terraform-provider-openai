package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationSpendAlertsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationSpendAlertsDataSource{}
)

type organizationSpendAlertsDataSource struct{ client client.AdminClient }

type organizationSpendAlertsDataSourceModel struct {
	After   types.String                            `tfsdk:"after"`
	Before  types.String                            `tfsdk:"before"`
	Limit   types.Int64                             `tfsdk:"limit"`
	Order   types.String                            `tfsdk:"order"`
	Items   []organizationSpendAlertDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                              `tfsdk:"has_more"`
	LastID  types.String                            `tfsdk:"last_id"`
}

func NewOrganizationSpendAlertsDataSource() datasource.DataSource {
	return &organizationSpendAlertsDataSource{}
}

func (d *organizationSpendAlertsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_alerts"
}

func (d *organizationSpendAlertsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization spend alerts through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for the next page."},
			"before":   datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for the previous page."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of spend alerts to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order by creation time: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Spend alerts returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: organizationSpendAlertDataSourceAttributes(false)}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more spend alerts are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last spend alert in the page."},
		},
	}
}

func (d *organizationSpendAlertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationSpendAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationSpendAlertsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationSpendAlerts(ctx, client.SpendAlertListRequest{After: stringValue(config.After), Before: stringValue(config.Before), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization spend alerts", err)
		return
	}
	state := organizationSpendAlertsDataSourceModel{After: config.After, Before: config.Before, Limit: config.Limit, Order: config.Order, Items: make([]organizationSpendAlertDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for _, alert := range page.Items {
		item, diags := organizationSpendAlertDataSourceModelFromAPI(ctx, &alert)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Items = append(state.Items, item)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
