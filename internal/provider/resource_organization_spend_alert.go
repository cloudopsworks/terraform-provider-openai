package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationSpendAlertResource{}
	_ resource.ResourceWithConfigure   = &organizationSpendAlertResource{}
	_ resource.ResourceWithImportState = &organizationSpendAlertResource{}
)

type organizationSpendAlertResource struct{ client client.AdminClient }

type organizationSpendAlertResourceModel struct {
	ID                  types.String                       `tfsdk:"id"`
	ThresholdAmount     types.Int64                        `tfsdk:"threshold_amount"`
	Currency            types.String                       `tfsdk:"currency"`
	Interval            types.String                       `tfsdk:"interval"`
	NotificationChannel spendAlertNotificationChannelModel `tfsdk:"notification_channel"`
}

func NewOrganizationSpendAlertResource() resource.Resource { return &organizationSpendAlertResource{} }

func (r *organizationSpendAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_alert"
}

func (r *organizationSpendAlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization spend alert. Spend alerts notify email recipients when monthly organization spend reaches the configured threshold amount in cents.",
		Attributes: map[string]resourceschema.Attribute{
			"id":               resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI spend alert ID."},
			"threshold_amount": resourceschema.Int64Attribute{Required: true, MarkdownDescription: "Alert threshold amount in cents. OpenAI accepts zero or greater."},
			"currency":         resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("USD"), MarkdownDescription: "Currency for threshold_amount. OpenAI currently supports USD.", Validators: []validator.String{newStringEnumValidator(spendCurrencies...)}},
			"interval":         resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("month"), MarkdownDescription: "Spend evaluation interval. OpenAI currently supports month.", Validators: []validator.String{newStringEnumValidator(spendIntervals...)}},
			"notification_channel": resourceschema.SingleNestedAttribute{Required: true, MarkdownDescription: "Email notification settings for the spend alert.", Attributes: map[string]resourceschema.Attribute{
				"recipients":     resourceschema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Email addresses that receive spend alert notifications."},
				"type":           resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("email"), MarkdownDescription: "Notification channel type. OpenAI currently supports email.", Validators: []validator.String{newStringEnumValidator("email")}},
				"subject_prefix": resourceschema.StringAttribute{Optional: true, MarkdownDescription: "Optional subject prefix for alert emails."},
			}},
		},
	}
}

func (r *organizationSpendAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, err := configureClient(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("OpenAI provider not configured", err.Error())
		return
	}
	r.client = data.client
}

func (r *organizationSpendAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	channel, diags := spendAlertNotificationChannelFromModel(ctx, plan.NotificationChannel, path.Root("notification_channel"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := r.client.CreateOrganizationSpendAlert(ctx, client.SpendAlertCreateRequest{ThresholdAmount: plan.ThresholdAmount.ValueInt64(), Currency: plan.Currency.ValueString(), Interval: plan.Interval.ValueString(), NotificationChannel: channel})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI organization spend alert", err)
		return
	}
	state, stateDiags := organizationSpendAlertResourceModelFromAPI(ctx, alert)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationSpendAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := r.client.GetOrganizationSpendAlert(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization spend alert", err)
		return
	}
	newState, diags := organizationSpendAlertResourceModelFromAPI(ctx, alert)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationSpendAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	channel, diags := spendAlertNotificationChannelFromModel(ctx, plan.NotificationChannel, path.Root("notification_channel"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := r.client.UpdateOrganizationSpendAlert(ctx, plan.ID.ValueString(), client.SpendAlertUpdateRequest{ThresholdAmount: plan.ThresholdAmount.ValueInt64(), Currency: plan.Currency.ValueString(), Interval: plan.Interval.ValueString(), NotificationChannel: channel})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization spend alert", err)
		return
	}
	state, stateDiags := organizationSpendAlertResourceModelFromAPI(ctx, alert)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationSpendAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganizationSpendAlert(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI organization spend alert", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationSpendAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func organizationSpendAlertResourceModelFromAPI(ctx context.Context, alert *client.SpendAlert) (organizationSpendAlertResourceModel, diag.Diagnostics) {
	channel, diags := spendAlertNotificationChannelModelFromAPI(ctx, alert.NotificationChannel)
	return organizationSpendAlertResourceModel{ID: types.StringValue(alert.ID), ThresholdAmount: types.Int64Value(alert.ThresholdAmount), Currency: stringOrNull(alert.Currency), Interval: stringOrNull(alert.Interval), NotificationChannel: channel}, diags
}
