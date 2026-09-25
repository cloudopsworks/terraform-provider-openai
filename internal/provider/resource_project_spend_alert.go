package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &projectSpendAlertResource{}
	_ resource.ResourceWithConfigure   = &projectSpendAlertResource{}
	_ resource.ResourceWithImportState = &projectSpendAlertResource{}
)

type projectSpendAlertResource struct{ client client.AdminClient }

type projectSpendAlertResourceModel struct {
	ID                  types.String                       `tfsdk:"id"`
	ProjectID           types.String                       `tfsdk:"project_id"`
	ThresholdAmount     types.Int64                        `tfsdk:"threshold_amount"`
	Currency            types.String                       `tfsdk:"currency"`
	Interval            types.String                       `tfsdk:"interval"`
	NotificationChannel spendAlertNotificationChannelModel `tfsdk:"notification_channel"`
}

func NewProjectSpendAlertResource() resource.Resource { return &projectSpendAlertResource{} }

func (r *projectSpendAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_alert"
}

func (r *projectSpendAlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI project spend alert. Spend alerts notify email recipients when monthly project spend reaches the configured threshold amount in cents.",
		Attributes: map[string]resourceschema.Attribute{
			"id":               resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI spend alert ID."},
			"project_id":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the spend alert.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
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

func (r *projectSpendAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectSpendAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.NotificationChannel.SubjectPrefix.IsNull() && plan.NotificationChannel.SubjectPrefix.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("notification_channel").AtName("subject_prefix"), "Invalid OpenAI spend alert subject prefix", "Omit subject_prefix to clear it; an explicitly empty prefix is not supported.")
		return
	}
	channel, diags := spendAlertNotificationChannelFromModel(ctx, plan.NotificationChannel, path.Root("notification_channel"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := r.client.CreateProjectSpendAlert(ctx, plan.ProjectID.ValueString(), client.SpendAlertCreateRequest{ThresholdAmount: plan.ThresholdAmount.ValueInt64(), Currency: plan.Currency.ValueString(), Interval: plan.Interval.ValueString(), NotificationChannel: channel})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI project spend alert", err)
		return
	}
	state, stateDiags := projectSpendAlertResourceModelFromAPI(ctx, alert, plan.ProjectID)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectSpendAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := r.client.GetProjectSpendAlert(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project spend alert", err)
		return
	}
	newState, diags := projectSpendAlertResourceModelFromAPI(ctx, alert, state.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *projectSpendAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectSpendAlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.NotificationChannel.SubjectPrefix.IsNull() && plan.NotificationChannel.SubjectPrefix.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("notification_channel").AtName("subject_prefix"), "Invalid OpenAI spend alert subject prefix", "Omit subject_prefix to clear it; an explicitly empty prefix is not supported.")
		return
	}
	channel, diags := spendAlertNotificationChannelFromModel(ctx, plan.NotificationChannel, path.Root("notification_channel"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	alert, err := r.client.UpdateProjectSpendAlert(ctx, plan.ProjectID.ValueString(), plan.ID.ValueString(), client.SpendAlertUpdateRequest{ThresholdAmount: plan.ThresholdAmount.ValueInt64(), Currency: plan.Currency.ValueString(), Interval: plan.Interval.ValueString(), NotificationChannel: channel})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project spend alert", err)
		return
	}
	state, stateDiags := projectSpendAlertResourceModelFromAPI(ctx, alert, plan.ProjectID)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectSpendAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectSpendAlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectSpendAlert(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI project spend alert", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *projectSpendAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, alertID, err := parseTwoPartImportID(req.ID, "project_id", "spend_alert_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(alertID))...)
}

func projectSpendAlertResourceModelFromAPI(ctx context.Context, alert *client.SpendAlert, projectID types.String) (projectSpendAlertResourceModel, diag.Diagnostics) {
	channel, diags := spendAlertNotificationChannelModelFromAPI(ctx, alert.NotificationChannel)
	return projectSpendAlertResourceModel{ID: types.StringValue(alert.ID), ProjectID: projectID, ThresholdAmount: types.Int64Value(alert.ThresholdAmount), Currency: stringOrNull(alert.Currency), Interval: stringOrNull(alert.Interval), NotificationChannel: channel}, diags
}
