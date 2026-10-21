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
	_ datasource.DataSource              = &projectSpendAlertDataSource{}
	_ datasource.DataSourceWithConfigure = &projectSpendAlertDataSource{}
)

type projectSpendAlertDataSource struct{ client client.AdminClient }

type projectSpendAlertDataSourceModel struct {
	ProjectID           types.String                       `tfsdk:"project_id"`
	ID                  types.String                       `tfsdk:"id"`
	ThresholdAmount     types.Int64                        `tfsdk:"threshold_amount"`
	Currency            types.String                       `tfsdk:"currency"`
	Interval            types.String                       `tfsdk:"interval"`
	NotificationChannel spendAlertNotificationChannelModel `tfsdk:"notification_channel"`
}

func NewProjectSpendAlertDataSource() datasource.DataSource {
	return &projectSpendAlertDataSource{}
}

func (d *projectSpendAlertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_alert"
}

func (d *projectSpendAlertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := projectSpendAlertDataSourceAttributes(true)
	attrs["project_id"] = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the spend alert."}
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI project spend alert by ID.", Attributes: attrs}
}

func projectSpendAlertDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
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

func (d *projectSpendAlertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectSpendAlertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectSpendAlertDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := d.client.GetProjectSpendAlert(ctx, config.ProjectID.ValueString(), config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project spend alert", err)
		return
	}
	state, diags := projectSpendAlertDataSourceModelFromAPI(ctx, alert, config.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func projectSpendAlertDataSourceModelFromAPI(ctx context.Context, alert *client.SpendAlert, projectID types.String) (projectSpendAlertDataSourceModel, diag.Diagnostics) {
	channel, diags := spendAlertNotificationChannelModelFromAPI(ctx, alert.NotificationChannel)
	return projectSpendAlertDataSourceModel{ProjectID: projectID, ID: types.StringValue(alert.ID), ThresholdAmount: types.Int64Value(alert.ThresholdAmount), Currency: stringOrNull(alert.Currency), Interval: stringOrNull(alert.Interval), NotificationChannel: channel}, diags
}
