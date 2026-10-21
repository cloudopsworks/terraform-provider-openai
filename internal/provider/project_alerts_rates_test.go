package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

type projectAlertsRatesClient struct {
	client.AdminClient
	alert                    *client.SpendAlert
	limits                   []client.RateLimit
	pagedLimits              []client.RateLimit
	updatedRateLimitRequest  client.RateLimitUpdateRequest
	updatedSpendAlertRequest client.SpendAlertUpdateRequest
}

func (f *projectAlertsRatesClient) CreateProjectSpendAlert(_ context.Context, _ string, req client.SpendAlertCreateRequest) (*client.SpendAlert, error) {
	f.alert = &client.SpendAlert{ID: "alert_1", ThresholdAmount: req.ThresholdAmount, Currency: req.Currency, Interval: req.Interval, NotificationChannel: req.NotificationChannel}
	return f.alert, nil
}
func (f *projectAlertsRatesClient) GetProjectSpendAlert(_ context.Context, _, _ string) (*client.SpendAlert, error) {
	return f.alert, nil
}
func (f *projectAlertsRatesClient) ListProjectSpendAlerts(_ context.Context, _ string, _ client.SpendAlertListRequest) (*client.SpendAlertListResponse, error) {
	return &client.SpendAlertListResponse{Items: []client.SpendAlert{*f.alert}, LastID: f.alert.ID}, nil
}
func (f *projectAlertsRatesClient) UpdateProjectSpendAlert(_ context.Context, _, id string, req client.SpendAlertUpdateRequest) (*client.SpendAlert, error) {
	f.updatedSpendAlertRequest = req
	f.alert = &client.SpendAlert{ID: id, ThresholdAmount: req.ThresholdAmount, Currency: req.Currency, Interval: req.Interval, NotificationChannel: req.NotificationChannel}
	return f.alert, nil
}
func (f *projectAlertsRatesClient) DeleteProjectSpendAlert(_ context.Context, _, _ string) error {
	return nil
}
func (f *projectAlertsRatesClient) ListProjectRateLimits(_ context.Context, _ string, req client.RateLimitListRequest) (*client.RateLimitListResponse, error) {
	if len(f.pagedLimits) > 0 {
		if req.After == "" {
			return &client.RateLimitListResponse{Items: []client.RateLimit{f.pagedLimits[0]}, HasMore: true, LastID: f.pagedLimits[0].ID}, nil
		}
		return &client.RateLimitListResponse{Items: f.pagedLimits[1:], LastID: f.pagedLimits[len(f.pagedLimits)-1].ID}, nil
	}
	return &client.RateLimitListResponse{Items: f.limits, LastID: f.limits[len(f.limits)-1].ID}, nil
}
func (f *projectAlertsRatesClient) UpdateProjectRateLimit(_ context.Context, _, id string, req client.RateLimitUpdateRequest) (*client.RateLimit, error) {
	f.updatedRateLimitRequest = req
	result := &client.RateLimit{ID: id, Model: "gpt-5"}
	if req.MaxRequestsPer1Minute != nil {
		result.MaxRequestsPer1Minute = *req.MaxRequestsPer1Minute
	}
	return result, nil
}

func TestProjectSpendAlertResourceLifecycleAndSchemas(t *testing.T) {
	ctx := context.Background()
	fake := &projectAlertsRatesClient{}
	recipients, diags := setStringValue(ctx, []string{"finops@example.com"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	r := &projectSpendAlertResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	if _, ok := schema.Attributes["project_id"]; !ok {
		t.Fatal("project_id is missing from project spend alert resource schema")
	}
	plan := tfsdk.Plan{Schema: schema}
	value := projectSpendAlertResourceModel{ProjectID: types.StringValue("proj_1"), ThresholdAmount: types.Int64Value(100), Currency: types.StringValue("USD"), Interval: types.StringValue("month"), NotificationChannel: spendAlertNotificationChannelModel{Recipients: recipients, Type: types.StringValue("email")}}
	if diags := plan.Set(ctx, &value); diags.HasError() {
		t.Fatal(diags)
	}
	create := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	var state projectSpendAlertResourceModel
	if diags := create.State.Get(ctx, &state); diags.HasError() || state.ID.ValueString() != "alert_1" || state.ProjectID.ValueString() != "proj_1" {
		t.Fatalf("state=%#v diagnostics=%v", state, diags)
	}
	state.ThresholdAmount = types.Int64Value(200)
	updatePlan := tfsdk.Plan{Schema: schema}
	if diags := updatePlan.Set(ctx, &state); diags.HasError() {
		t.Fatal(diags)
	}
	update := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: updatePlan}, &update)
	if update.Diagnostics.HasError() {
		t.Fatal(update.Diagnostics)
	}
	ds := &projectSpendAlertsDataSource{client: fake}
	dsSchema := dataSourceSchema(ctx, t, ds)
	if _, ok := dsSchema.Attributes["project_id"]; !ok {
		t.Fatal("project_id is missing from spend-alert list schema")
	}
	configPlan := tfsdk.Plan{Schema: dsSchema}
	if diags := configPlan.Set(ctx, &projectSpendAlertsDataSourceModel{ProjectID: types.StringValue("proj_1")}); diags.HasError() {
		t.Fatal(diags)
	}
	response := datasource.ReadResponse{State: tfsdk.State{Schema: dsSchema}}
	ds.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: dsSchema, Raw: configPlan.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
}

func TestProjectRateLimitResourceLifecycleAndSchemas(t *testing.T) {
	ctx := context.Background()
	fake := &projectAlertsRatesClient{limits: []client.RateLimit{{ID: "rl_1", Model: "gpt-5", MaxRequestsPer1Minute: 10}}}
	r := &projectRateLimitResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	if attribute, ok := schema.Attributes["id"]; !ok || attribute == nil {
		t.Fatal("id is missing from project rate-limit resource schema")
	}
	plan := tfsdk.Plan{Schema: schema}
	value := projectRateLimitResourceModel{ID: types.StringValue("rl_1"), ProjectID: types.StringValue("proj_1"), MaxRequestsPer1Minute: types.Int64Value(0)}
	if diags := plan.Set(ctx, &value); diags.HasError() {
		t.Fatal(diags)
	}
	create := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	if fake.updatedRateLimitRequest.MaxRequestsPer1Minute == nil || *fake.updatedRateLimitRequest.MaxRequestsPer1Minute != 0 {
		t.Fatalf("zero value was not sent as an explicit rate limit: %#v", fake.updatedRateLimitRequest)
	}
	deleteResponse := resource.DeleteResponse{State: create.State}
	r.Delete(ctx, resource.DeleteRequest{State: create.State}, &deleteResponse)
	if deleteResponse.Diagnostics.HasError() || len(deleteResponse.Diagnostics) == 0 {
		t.Fatalf("delete diagnostics=%v", deleteResponse.Diagnostics)
	}
	ds := &projectRateLimitsDataSource{client: fake}
	dsSchema := dataSourceSchema(ctx, t, ds)
	configPlan := tfsdk.Plan{Schema: dsSchema}
	if diags := configPlan.Set(ctx, &projectRateLimitsDataSourceModel{ProjectID: types.StringValue("proj_1")}); diags.HasError() {
		t.Fatal(diags)
	}
	response := datasource.ReadResponse{State: tfsdk.State{Schema: dsSchema}}
	ds.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: dsSchema, Raw: configPlan.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
}

func TestProjectRateLimitReadFindsLaterPage(t *testing.T) {
	fake := &projectAlertsRatesClient{pagedLimits: []client.RateLimit{{ID: "rl_first"}, {ID: "rl_target", Model: "gpt-5"}}}
	limit, found, err := (&projectRateLimitResource{client: fake}).find(context.Background(), "proj_1", "rl_target")
	if err != nil || !found || limit.ID != "rl_target" {
		t.Fatalf("limit=%#v found=%t err=%v", limit, found, err)
	}
	dataSource := &projectRateLimitDataSource{client: fake}
	schema := dataSourceSchema(context.Background(), t, dataSource)
	configPlan := tfsdk.Plan{Schema: schema}
	if diags := configPlan.Set(context.Background(), &projectRateLimitDataSourceModel{ID: types.StringValue("rl_target"), ProjectID: types.StringValue("proj_1")}); diags.HasError() {
		t.Fatal(diags)
	}
	response := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	dataSource.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: schema, Raw: configPlan.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var state projectRateLimitDataSourceModel
	if diags := response.State.Get(context.Background(), &state); diags.HasError() || !state.Batch1DayMaxInputTokens.IsNull() {
		t.Fatalf("state=%#v diagnostics=%v", state, diags)
	}
}

func TestProjectRateLimitAbsentCapsRemainNullAndAreNotUpdated(t *testing.T) {
	model := projectRateLimitResourceModel{MaxRequestsPer1Minute: types.Int64Value(10), MaxTokensPer1Minute: types.Int64Value(20)}
	request := projectRateLimitUpdateRequestFromModel(model)
	if request.MaxRequestsPer1Minute == nil || request.MaxTokensPer1Minute == nil || request.Batch1DayMaxInputTokens != nil || request.MaxAudioMegabytesPer1Minute != nil || request.MaxImagesPer1Minute != nil || request.MaxRequestsPer1Day != nil {
		t.Fatalf("unexpected update request for absent caps: %#v", request)
	}
	state := projectRateLimitResourceModelFromAPI(&client.RateLimit{ID: "rl_1", MaxRequestsPer1Minute: 10, MaxTokensPer1Minute: 20}, types.StringValue("proj_1"))
	if !state.Batch1DayMaxInputTokens.IsNull() || !state.MaxAudioMegabytesPer1Minute.IsNull() || !state.MaxImagesPer1Minute.IsNull() || !state.MaxRequestsPer1Day.IsNull() {
		t.Fatalf("optional caps should be null: %#v", state)
	}
}

func TestProjectSpendAlertOmittedSubjectPrefixClearsOnUpdate(t *testing.T) {
	ctx := context.Background()
	fake := &projectAlertsRatesClient{}
	recipients, diags := setStringValue(ctx, []string{"finops@example.com"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	r := &projectSpendAlertResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	plan := tfsdk.Plan{Schema: schema}
	value := projectSpendAlertResourceModel{ID: types.StringValue("alert_1"), ProjectID: types.StringValue("proj_1"), ThresholdAmount: types.Int64Value(100), Currency: types.StringValue("USD"), Interval: types.StringValue("month"), NotificationChannel: spendAlertNotificationChannelModel{Recipients: recipients, Type: types.StringValue("email"), SubjectPrefix: types.StringNull()}}
	if diags := plan.Set(ctx, &value); diags.HasError() {
		t.Fatal(diags)
	}
	response := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	if fake.updatedSpendAlertRequest.NotificationChannel.SubjectPrefix != "" {
		t.Fatalf("omitted prefix was not normalized for update: %#v", fake.updatedSpendAlertRequest)
	}
}
