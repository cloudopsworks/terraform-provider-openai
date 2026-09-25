package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectSpendAlertsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectSpendAlertsDataSource{}
)

type projectSpendAlertsDataSource struct{ client client.AdminClient }

type projectSpendAlertsDataSourceModel struct {
	ProjectID types.String                           `tfsdk:"project_id"`
	After     types.String                           `tfsdk:"after"`
	Before    types.String                           `tfsdk:"before"`
	Limit     types.Int64                            `tfsdk:"limit"`
	Order     types.String                           `tfsdk:"order"`
	Items     []projectSpendAlertItemDataSourceModel `tfsdk:"items"`
	HasMore   types.Bool                             `tfsdk:"has_more"`
	LastID    types.String                           `tfsdk:"last_id"`
}

type projectSpendAlertItemDataSourceModel struct {
	ID                  types.String                       `tfsdk:"id"`
	ThresholdAmount     types.Int64                        `tfsdk:"threshold_amount"`
	Currency            types.String                       `tfsdk:"currency"`
	Interval            types.String                       `tfsdk:"interval"`
	NotificationChannel spendAlertNotificationChannelModel `tfsdk:"notification_channel"`
}

func projectSpendAlertItemDataSourceModelFromAPI(ctx context.Context, alert *client.SpendAlert) (projectSpendAlertItemDataSourceModel, diag.Diagnostics) {
	channel, diags := spendAlertNotificationChannelModelFromAPI(ctx, alert.NotificationChannel)
	return projectSpendAlertItemDataSourceModel{ID: types.StringValue(alert.ID), ThresholdAmount: types.Int64Value(alert.ThresholdAmount), Currency: stringOrNull(alert.Currency), Interval: stringOrNull(alert.Interval), NotificationChannel: channel}, diags
}

func NewProjectSpendAlertsDataSource() datasource.DataSource {
	return &projectSpendAlertsDataSource{}
}

func (d *projectSpendAlertsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_alerts"
}

func (d *projectSpendAlertsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI project spend alerts through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the spend alerts."},
			"after":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for the next page."},
			"before":     datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for the previous page."},
			"limit":      datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of spend alerts to return."},
			"order":      datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order by creation time: asc or desc."},
			"items":      datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Spend alerts returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: projectSpendAlertDataSourceAttributes(false)}},
			"has_more":   datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more spend alerts are available after this page."},
			"last_id":    datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last spend alert in the page."},
		},
	}
}

func (d *projectSpendAlertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectSpendAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectSpendAlertsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListProjectSpendAlerts(ctx, config.ProjectID.ValueString(), client.SpendAlertListRequest{After: stringValue(config.After), Before: stringValue(config.Before), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project spend alerts", err)
		return
	}
	state := projectSpendAlertsDataSourceModel{ProjectID: config.ProjectID, After: config.After, Before: config.Before, Limit: config.Limit, Order: config.Order, Items: make([]projectSpendAlertItemDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for _, alert := range page.Items {
		item, diags := projectSpendAlertItemDataSourceModelFromAPI(ctx, &alert)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Items = append(state.Items, item)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
