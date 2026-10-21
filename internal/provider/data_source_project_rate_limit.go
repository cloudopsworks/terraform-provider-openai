package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &projectRateLimitDataSource{}
	_ datasource.DataSourceWithConfigure = &projectRateLimitDataSource{}
)

type projectRateLimitDataSource struct{ client client.AdminClient }

type projectRateLimitDataSourceModel struct {
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

func NewProjectRateLimitDataSource() datasource.DataSource { return &projectRateLimitDataSource{} }
func (d *projectRateLimitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_rate_limit"
}
func (d *projectRateLimitDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI project rate limit by ID.", Attributes: projectRateLimitDataSourceAttributes(true)}
}
func projectRateLimitDataSourceAttributes(identifierRequired bool) map[string]datasourceschema.Attribute {
	id := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI rate-limit ID."}
	projectID := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID that owns the rate limit."}
	if identifierRequired {
		id = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI rate-limit ID."}
		projectID = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the rate limit."}
	}
	return map[string]datasourceschema.Attribute{
		"id": id, "project_id": projectID,
		"model":                            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Model governed by the rate limit."},
		"max_requests_per_1_minute":        datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum requests per minute."},
		"max_tokens_per_1_minute":          datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum tokens per minute."},
		"batch_1_day_max_input_tokens":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum batch input tokens per day."},
		"max_audio_megabytes_per_1_minute": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum audio megabytes per minute."},
		"max_images_per_1_minute":          datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum images per minute."},
		"max_requests_per_1_day":           datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum requests per day."},
	}
}
func (d *projectRateLimitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectRateLimitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectRateLimitDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit, found, err := findProjectRateLimit(ctx, d.client, config.ProjectID.ValueString(), config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI project rate limits", err)
		return
	}
	if found {
		state := projectRateLimitDataSourceModelFromAPI(limit, config.ProjectID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}
	resp.Diagnostics.AddError("OpenAI project rate limit not found", fmt.Sprintf("No project rate limit with ID %q exists in project %q.", config.ID.ValueString(), config.ProjectID.ValueString()))
}
func projectRateLimitDataSourceModelFromAPI(limit *client.RateLimit, projectID types.String) projectRateLimitDataSourceModel {
	return projectRateLimitDataSourceModel{ID: types.StringValue(limit.ID), ProjectID: projectID, Model: stringOrNull(limit.Model), MaxRequestsPer1Minute: types.Int64Value(limit.MaxRequestsPer1Minute), MaxTokensPer1Minute: types.Int64Value(limit.MaxTokensPer1Minute), Batch1DayMaxInputTokens: optionalInt64OrNull(limit.Batch1DayMaxInputTokens), MaxAudioMegabytesPer1Minute: optionalInt64OrNull(limit.MaxAudioMegabytesPer1Minute), MaxImagesPer1Minute: optionalInt64OrNull(limit.MaxImagesPer1Minute), MaxRequestsPer1Day: optionalInt64OrNull(limit.MaxRequestsPer1Day)}
}
