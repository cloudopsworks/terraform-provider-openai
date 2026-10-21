package provider

import (
	"context"

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
	_ resource.Resource                = &projectSpendLimitResource{}
	_ resource.ResourceWithConfigure   = &projectSpendLimitResource{}
	_ resource.ResourceWithImportState = &projectSpendLimitResource{}
)

type projectSpendLimitResource struct{ client client.AdminClient }
type projectSpendLimitResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	ThresholdAmount   types.Int64  `tfsdk:"threshold_amount"`
	Currency          types.String `tfsdk:"currency"`
	Interval          types.String `tfsdk:"interval"`
	EnforcementStatus types.String `tfsdk:"enforcement_status"`
}

func NewProjectSpendLimitResource() resource.Resource { return &projectSpendLimitResource{} }
func (r *projectSpendLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_spend_limit"
}
func (r *projectSpendLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI project hard spend limit. Destroy deletes the project spend limit.", Attributes: map[string]resourceschema.Attribute{
		"id":                 resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."},
		"project_id":         resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"threshold_amount":   resourceschema.Int64Attribute{Required: true, MarkdownDescription: "Hard spend limit amount in cents."},
		"currency":           resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("USD"), MarkdownDescription: "Currency for threshold_amount. OpenAI currently supports USD.", Validators: []validator.String{newStringEnumValidator(spendCurrencies...)}},
		"interval":           resourceschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("month"), MarkdownDescription: "Spend evaluation interval. OpenAI currently supports month.", Validators: []validator.String{newStringEnumValidator(spendIntervals...)}},
		"enforcement_status": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI enforcement status for the hard spend limit."},
	}}
}
func (r *projectSpendLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectSpendLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project spend limit", err)
		return
	}
	state := projectSpendLimitResourceModelFromAPI(limit, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectSpendLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectSpendLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.client.GetProjectSpendLimit(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project spend limit", err)
		return
	}
	state = projectSpendLimitResourceModelFromAPI(limit, state.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectSpendLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectSpendLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project spend limit", err)
		return
	}
	state := projectSpendLimitResourceModelFromAPI(limit, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectSpendLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectSpendLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectSpendLimit(ctx, state.ProjectID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI project spend limit", err)
		return
	}
	resp.State.RemoveResource(ctx)
}
func (r *projectSpendLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}
func (r *projectSpendLimitResource) update(ctx context.Context, plan projectSpendLimitResourceModel) (*client.SpendLimit, error) {
	return r.client.UpdateProjectSpendLimit(ctx, plan.ProjectID.ValueString(), client.SpendLimitUpdateRequest{ThresholdAmount: plan.ThresholdAmount.ValueInt64(), Currency: plan.Currency.ValueString(), Interval: plan.Interval.ValueString()})
}
func projectSpendLimitResourceModelFromAPI(limit *client.SpendLimit, projectID types.String) projectSpendLimitResourceModel {
	return projectSpendLimitResourceModel{ID: projectID, ProjectID: projectID, ThresholdAmount: types.Int64Value(limit.ThresholdAmount), Currency: stringOrNull(limit.Currency), Interval: stringOrNull(limit.Interval), EnforcementStatus: stringOrNull(limit.EnforcementStatus)}
}
