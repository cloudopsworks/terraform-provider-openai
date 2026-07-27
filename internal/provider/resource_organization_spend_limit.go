package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationSpendLimitResource{}
	_ resource.ResourceWithConfigure   = &organizationSpendLimitResource{}
	_ resource.ResourceWithImportState = &organizationSpendLimitResource{}
)

type organizationSpendLimitResource struct{ client client.AdminClient }

type organizationSpendLimitResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ThresholdAmount   types.Int64  `tfsdk:"threshold_amount"`
	Currency          types.String `tfsdk:"currency"`
	Interval          types.String `tfsdk:"interval"`
	EnforcementStatus types.String `tfsdk:"enforcement_status"`
}

func NewOrganizationSpendLimitResource() resource.Resource { return &organizationSpendLimitResource{} }

func (r *organizationSpendLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_spend_limit"
}

func (r *organizationSpendLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization hard spend limit. Create and update both call OpenAI's spend-limit update endpoint; destroy deletes the organization spend limit.",
		Attributes: map[string]resourceschema.Attribute{
			"id":                 resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID. Always organization."},
			"threshold_amount":   resourceschema.Int64Attribute{Required: true, MarkdownDescription: "Hard spend limit amount in cents."},
			"currency":           resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("USD"), MarkdownDescription: "Currency for threshold_amount. OpenAI currently supports USD.", Validators: []validator.String{newStringEnumValidator(spendCurrencies...)}},
			"interval":           resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("month"), MarkdownDescription: "Spend evaluation interval. OpenAI currently supports month.", Validators: []validator.String{newStringEnumValidator(spendIntervals...)}},
			"enforcement_status": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI enforcement status for the hard spend limit: inactive or enforcing."},
		},
	}
}

func (r *organizationSpendLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationSpendLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization spend limit", err)
		return
	}
	state := organizationSpendLimitResourceModelFromAPI(limit)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationSpendLimitResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	limit, err := r.client.GetOrganizationSpendLimit(ctx)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization spend limit", err)
		return
	}
	state := organizationSpendLimitResourceModelFromAPI(limit)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationSpendLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization spend limit", err)
		return
	}
	state := organizationSpendLimitResourceModelFromAPI(limit)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationSpendLimitResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeleteOrganizationSpendLimit(ctx); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI organization spend limit", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationSpendLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *organizationSpendLimitResource) update(ctx context.Context, plan organizationSpendLimitResourceModel) (*client.SpendLimit, error) {
	return r.client.UpdateOrganizationSpendLimit(ctx, client.SpendLimitUpdateRequest{ThresholdAmount: plan.ThresholdAmount.ValueInt64(), Currency: plan.Currency.ValueString(), Interval: plan.Interval.ValueString()})
}

func organizationSpendLimitResourceModelFromAPI(limit *client.SpendLimit) organizationSpendLimitResourceModel {
	return organizationSpendLimitResourceModel{ID: types.StringValue(organizationSingletonID), ThresholdAmount: types.Int64Value(limit.ThresholdAmount), Currency: stringOrNull(limit.Currency), Interval: stringOrNull(limit.Interval), EnforcementStatus: stringOrNull(limit.EnforcementStatus)}
}
