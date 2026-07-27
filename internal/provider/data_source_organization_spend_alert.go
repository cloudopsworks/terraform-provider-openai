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
	_ datasource.DataSource              = &organizationSpendAlertDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationSpendAlertDataSource{}
)

type organizationSpendAlertDataSource struct{ client client.AdminClient }

type organizationSpendAlertDataSourceModel struct {
	ID                  types.String                       `tfsdk:"id"`
	ThresholdAmount     types.Int64                        `tfsdk:"threshold_amount"`
	Currency            types.String                       `tfsdk:"currency"`
	Interval            types.String                       `tfsdk:"interval"`
	NotificationChannel spendAlertNotificationChannelModel `tfsdk:"notification_channel"`
}

func NewOrganizationSpendAlertDataSource() datasource.DataSource {
	return &organizationSpendAlertDataSource{}
}

func (d *organizationSpendAlertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_alert"
}

func (d *organizationSpendAlertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization spend alert by ID.", Attributes: organizationSpendAlertDataSourceAttributes(true)}
}

func organizationSpendAlertDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI spend alert ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI spend alert ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":               idAttr,
		"threshold_amount": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Alert threshold amount in cents."},
		"currency":         datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Currency for threshold_amount."},
		"interval":         datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Spend evaluation interval."},
		"notification_channel": datasourceschema.SingleNestedAttribute{Computed: true, MarkdownDescription: "Email notification settings for the spend alert.", Attributes: map[string]datasourceschema.Attribute{
			"recipients":     datasourceschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Email addresses that receive spend alert notifications."},
			"type":           datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Notification channel type."},
			"subject_prefix": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Optional subject prefix for alert emails."},
		}},
	}
}

func (d *organizationSpendAlertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationSpendAlertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationSpendAlertDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := d.client.GetOrganizationSpendAlert(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization spend alert", err)
		return
	}
	state, diags := organizationSpendAlertDataSourceModelFromAPI(ctx, alert)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func organizationSpendAlertDataSourceModelFromAPI(ctx context.Context, alert *client.SpendAlert) (organizationSpendAlertDataSourceModel, diag.Diagnostics) {
	channel, diags := spendAlertNotificationChannelModelFromAPI(ctx, alert.NotificationChannel)
	return organizationSpendAlertDataSourceModel{ID: types.StringValue(alert.ID), ThresholdAmount: types.Int64Value(alert.ThresholdAmount), Currency: stringOrNull(alert.Currency), Interval: stringOrNull(alert.Interval), NotificationChannel: channel}, diags
}
