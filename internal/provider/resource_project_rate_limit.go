package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &projectRateLimitResource{}
	_ resource.ResourceWithConfigure   = &projectRateLimitResource{}
	_ resource.ResourceWithImportState = &projectRateLimitResource{}
)

type projectRateLimitResource struct{ client client.AdminClient }

type projectRateLimitResourceModel struct {
	ID                          types.String `tfsdk:"id"`
	ProjectID                   types.String `tfsdk:"project_id"`
	Model                       types.String `tfsdk:"model"`
	MaxRequestsPer1Minute       types.Int64  `tfsdk:"max_requests_per_1_minute"`
	MaxTokensPer1Minute         types.Int64  `tfsdk:"max_tokens_per_1_minute"`
	Batch1DayMaxInputTokens     types.Int64  `tfsdk:"batch_1_day_max_input_tokens"`
	MaxAudioMegabytesPer1Minute types.Int64  `tfsdk:"max_audio_megabytes_per_1_minute"`
	MaxImagesPer1Minute         types.Int64  `tfsdk:"max_images_per_1_minute"`
	MaxRequestsPer1Day          types.Int64  `tfsdk:"max_requests_per_1_day"`
}

func NewProjectRateLimitResource() resource.Resource { return &projectRateLimitResource{} }

func (r *projectRateLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_rate_limit"
}

func (r *projectRateLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI project rate limit. OpenAI exposes updates but no create, delete, or reset endpoint; Terraform updates an existing rate-limit entry and removes only its state on destroy.",
		Attributes:          projectRateLimitResourceAttributes(),
	}
}

func projectRateLimitResourceAttributes() map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"id":                               resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI rate-limit ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"project_id":                       resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the rate limit.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"model":                            resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Model governed by the rate limit."},
		"max_requests_per_1_minute":        resourceschema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum requests per minute. Omit to leave this cap unchanged."},
		"max_tokens_per_1_minute":          resourceschema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum tokens per minute. Omit to leave this cap unchanged."},
		"batch_1_day_max_input_tokens":     resourceschema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum batch input tokens per day. Omit to leave this cap unchanged."},
		"max_audio_megabytes_per_1_minute": resourceschema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum audio megabytes per minute. Omit to leave this cap unchanged."},
		"max_images_per_1_minute":          resourceschema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum images per minute. Omit to leave this cap unchanged."},
		"max_requests_per_1_day":           resourceschema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Maximum requests per day. Omit to leave this cap unchanged."},
	}
}

func (r *projectRateLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectRateLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectRateLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project rate limit", err)
		return
	}
	state := projectRateLimitResourceModelFromAPI(limit, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectRateLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectRateLimitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, found, err := r.find(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project rate limit", err)
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	newState := projectRateLimitResourceModelFromAPI(limit, state.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *projectRateLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectRateLimitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project rate limit", err)
		return
	}
	state := projectRateLimitResourceModelFromAPI(limit, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectRateLimitResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("OpenAI project rate limit was not reset", "OpenAI exposes list and update for project rate limits but no delete or reset endpoint. Terraform is removing only its state object.")
	resp.State.RemoveResource(ctx)
}

func (r *projectRateLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, rateLimitID, err := parseTwoPartImportID(req.ID, "project_id", "rate_limit_id")
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(rateLimitID))...)
}

func (r *projectRateLimitResource) update(ctx context.Context, plan projectRateLimitResourceModel) (*client.RateLimit, error) {
	return r.client.UpdateProjectRateLimit(ctx, plan.ProjectID.ValueString(), plan.ID.ValueString(), projectRateLimitUpdateRequestFromModel(plan))
}

func (r *projectRateLimitResource) find(ctx context.Context, projectID, rateLimitID string) (*client.RateLimit, bool, error) {
	return findProjectRateLimit(ctx, r.client, projectID, rateLimitID)
}

func findProjectRateLimit(ctx context.Context, apiClient client.AdminClient, projectID, rateLimitID string) (*client.RateLimit, bool, error) {
	var after string
	for {
		page, err := apiClient.ListProjectRateLimits(ctx, projectID, client.RateLimitListRequest{After: after})
		if err != nil {
			return nil, false, err
		}
		for i := range page.Items {
			if page.Items[i].ID == rateLimitID {
				return &page.Items[i], true, nil
			}
		}
		if !page.HasMore || page.LastID == "" || page.LastID == after {
			return nil, false, nil
		}
		after = page.LastID
	}
}

func projectRateLimitUpdateRequestFromModel(model projectRateLimitResourceModel) client.RateLimitUpdateRequest {
	return client.RateLimitUpdateRequest{
		Batch1DayMaxInputTokens:     int64PointerOrNil(model.Batch1DayMaxInputTokens),
		MaxAudioMegabytesPer1Minute: int64PointerOrNil(model.MaxAudioMegabytesPer1Minute),
		MaxImagesPer1Minute:         int64PointerOrNil(model.MaxImagesPer1Minute),
		MaxRequestsPer1Day:          int64PointerOrNil(model.MaxRequestsPer1Day),
		MaxRequestsPer1Minute:       int64PointerOrNil(model.MaxRequestsPer1Minute),
		MaxTokensPer1Minute:         int64PointerOrNil(model.MaxTokensPer1Minute),
	}
}

func int64PointerOrNil(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	result := value.ValueInt64()
	return &result
}

func projectRateLimitResourceModelFromAPI(limit *client.RateLimit, projectID types.String) projectRateLimitResourceModel {
	return projectRateLimitResourceModel{ID: types.StringValue(limit.ID), ProjectID: projectID, Model: stringOrNull(limit.Model), MaxRequestsPer1Minute: types.Int64Value(limit.MaxRequestsPer1Minute), MaxTokensPer1Minute: types.Int64Value(limit.MaxTokensPer1Minute), Batch1DayMaxInputTokens: optionalInt64OrNull(limit.Batch1DayMaxInputTokens), MaxAudioMegabytesPer1Minute: optionalInt64OrNull(limit.MaxAudioMegabytesPer1Minute), MaxImagesPer1Minute: optionalInt64OrNull(limit.MaxImagesPer1Minute), MaxRequestsPer1Day: optionalInt64OrNull(limit.MaxRequestsPer1Day)}
}

func optionalInt64OrNull(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}
