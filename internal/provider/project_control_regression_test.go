package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// projectControlRegressionClient deliberately embeds the full interface so each
// test only needs to define the remote call it observes.
type projectControlRegressionClient struct {
	client.AdminClient
	spendLimitDeleteErr         error
	modelPermissionsDeleteErr   error
	spendLimitDeleteCalls       int
	modelPermissionsDeleteCalls int
}

func (f *projectControlRegressionClient) DeleteProjectSpendLimit(_ context.Context, _ string) error {
	f.spendLimitDeleteCalls++
	return f.spendLimitDeleteErr
}

func (f *projectControlRegressionClient) DeleteProjectModelPermissions(_ context.Context, _ string) error {
	f.modelPermissionsDeleteCalls++
	return f.modelPermissionsDeleteErr
}

func TestProjectStateOnlyPolicyDeletesRemoveStateAndWarn(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name   string
		state  any
		delete func(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
		schema resource.Resource
	}{
		{
			name:   "data retention",
			state:  &projectDataRetentionResourceModel{ID: types.StringValue("proj_1"), ProjectID: types.StringValue("proj_1"), Type: types.StringValue("none")},
			delete: (&projectDataRetentionResource{}).Delete,
			schema: &projectDataRetentionResource{},
		},
		{
			name:   "hosted tool permissions",
			state:  &projectHostedToolPermissionsResourceModel{ID: types.StringValue("proj_1"), ProjectID: types.StringValue("proj_1")},
			delete: (&projectHostedToolPermissionsResource{}).Delete,
			schema: &projectHostedToolPermissionsResource{},
		},
		{
			name:   "rate limit",
			state:  &projectRateLimitResourceModel{ID: types.StringValue("rl_1"), ProjectID: types.StringValue("proj_1"), Model: types.StringValue("gpt-5")},
			delete: (&projectRateLimitResource{}).Delete,
			schema: &projectRateLimitResource{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := resourceSchema(ctx, t, tt.schema)
			state := tfsdk.State{Schema: schema}
			if diags := state.Set(ctx, tt.state); diags.HasError() {
				t.Fatal(diags)
			}
			response := resource.DeleteResponse{State: state}
			tt.delete(ctx, resource.DeleteRequest{State: state}, &response)
			if response.Diagnostics.HasError() || len(response.Diagnostics) != 1 {
				t.Fatalf("expected exactly one state-only deletion warning, got %v", response.Diagnostics)
			}
			if !response.State.Raw.IsNull() {
				t.Fatal("state-only delete did not remove the Terraform state")
			}
		})
	}
}

func TestProjectRemotePolicyDeleteErrorsKeepState(t *testing.T) {
	ctx := context.Background()
	deleteErr := errors.New("remote delete failed")
	tests := []struct {
		name   string
		state  any
		delete func(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
		schema resource.Resource
		calls  func(*projectControlRegressionClient) int
	}{
		{
			name:   "spend limit",
			state:  &projectSpendLimitResourceModel{ID: types.StringValue("proj_1"), ProjectID: types.StringValue("proj_1"), ThresholdAmount: types.Int64Value(100), Currency: types.StringValue("USD"), Interval: types.StringValue("month")},
			schema: &projectSpendLimitResource{},
			calls:  func(f *projectControlRegressionClient) int { return f.spendLimitDeleteCalls },
		},
		{
			name:   "model permissions",
			state:  &projectModelPermissionsResourceModel{ID: types.StringValue("proj_1"), ProjectID: types.StringValue("proj_1"), Mode: types.StringValue("allow_list"), ModelIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("gpt-5")})},
			schema: &projectModelPermissionsResource{},
			calls:  func(f *projectControlRegressionClient) int { return f.modelPermissionsDeleteCalls },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &projectControlRegressionClient{spendLimitDeleteErr: deleteErr, modelPermissionsDeleteErr: deleteErr}
			if tt.name == "spend limit" {
				tt.delete = (&projectSpendLimitResource{client: fake}).Delete
			} else {
				tt.delete = (&projectModelPermissionsResource{client: fake}).Delete
			}
			schema := resourceSchema(ctx, t, tt.schema)
			state := tfsdk.State{Schema: schema}
			if diags := state.Set(ctx, tt.state); diags.HasError() {
				t.Fatal(diags)
			}
			response := resource.DeleteResponse{State: state}
			tt.delete(ctx, resource.DeleteRequest{State: state}, &response)
			if !response.Diagnostics.HasError() || tt.calls(fake) != 1 {
				t.Fatalf("diagnostics=%v calls=%d", response.Diagnostics, tt.calls(fake))
			}
			if response.State.Raw.IsNull() {
				t.Fatal("state was removed despite remote delete failure")
			}
		})
	}
}

func TestProjectSpendAlertImportUsesCompositeIdentity(t *testing.T) {
	ctx := context.Background()
	schema := resourceSchema(ctx, t, &projectSpendAlertResource{})
	response := resource.ImportStateResponse{State: emptyProjectSpendAlertImportState(ctx, t, schema)}
	(&projectSpendAlertResource{}).ImportState(ctx, resource.ImportStateRequest{ID: "proj_1/alert_1"}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var state projectSpendAlertResourceModel
	if diags := response.State.Get(ctx, &state); diags.HasError() {
		t.Fatal(diags)
	}
	if state.ProjectID.ValueString() != "proj_1" || state.ID.ValueString() != "alert_1" {
		t.Fatalf("unexpected import state: %#v", state)
	}
}

func TestProjectAccessRoleImportsUseCompositeIdentity(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name        string
		id          string
		importState func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
		schema      resource.Resource
		assert      func(*testing.T, tfsdk.State)
	}{
		{
			name: "user", id: "proj_1/user_1/role_1", schema: &projectUserRoleResource{},
			importState: (&projectUserRoleResource{}).ImportState,
			assert: func(t *testing.T, state tfsdk.State) {
				var model projectUserRoleResourceModel
				if diags := state.Get(ctx, &model); diags.HasError() {
					t.Fatal(diags)
				}
				if model.ProjectID.ValueString() != "proj_1" || model.UserID.ValueString() != "user_1" || model.RoleID.ValueString() != "role_1" || model.ID.ValueString() != "proj_1/user_1/role_1" {
					t.Fatalf("unexpected state: %#v", model)
				}
			},
		},
		{
			name: "group", id: "proj_1/group_1/role_1", schema: &projectGroupRoleResource{},
			importState: (&projectGroupRoleResource{}).ImportState,
			assert: func(t *testing.T, state tfsdk.State) {
				var model projectGroupRoleResourceModel
				if diags := state.Get(ctx, &model); diags.HasError() {
					t.Fatal(diags)
				}
				if model.ProjectID.ValueString() != "proj_1" || model.GroupID.ValueString() != "group_1" || model.RoleID.ValueString() != "role_1" || model.ID.ValueString() != "proj_1/group_1/role_1" {
					t.Fatalf("unexpected state: %#v", model)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := resource.ImportStateResponse{State: emptyProjectAccessRoleImportState(ctx, t, resourceSchema(ctx, t, tt.schema), tt.name)}
			tt.importState(ctx, resource.ImportStateRequest{ID: tt.id}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			tt.assert(t, response.State)
		})
	}
}

func emptyProjectSpendAlertImportState(ctx context.Context, t *testing.T, schema resourceschema.Schema) tfsdk.State {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	model := projectSpendAlertResourceModel{
		ID: types.StringNull(), ProjectID: types.StringNull(), ThresholdAmount: types.Int64Null(), Currency: types.StringNull(), Interval: types.StringNull(),
		NotificationChannel: spendAlertNotificationChannelModel{Recipients: types.SetNull(types.StringType), Type: types.StringNull(), SubjectPrefix: types.StringNull()},
	}
	if diags := plan.Set(ctx, &model); diags.HasError() {
		t.Fatalf("empty spend-alert state: %v", diags)
	}
	return tfsdk.State{Schema: schema, Raw: plan.Raw}
}

func emptyProjectAccessRoleImportState(ctx context.Context, t *testing.T, schema resourceschema.Schema, principal string) tfsdk.State {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	if principal == "user" {
		model := projectUserRoleResourceModel{ID: types.StringNull(), ProjectID: types.StringNull(), UserID: types.StringNull(), RoleID: types.StringNull(), Name: types.StringNull(), Description: types.StringNull(), Permissions: types.SetNull(types.StringType), PredefinedRole: types.BoolNull(), ResourceType: types.StringNull(), CreatedAt: types.Int64Null(), UpdatedAt: types.Int64Null(), CreatedBy: types.StringNull(), AssignmentSources: nil}
		if diags := plan.Set(ctx, &model); diags.HasError() {
			t.Fatalf("empty user-role state: %v", diags)
		}
	} else {
		model := projectGroupRoleResourceModel{ID: types.StringNull(), ProjectID: types.StringNull(), GroupID: types.StringNull(), RoleID: types.StringNull(), Name: types.StringNull(), Description: types.StringNull(), Permissions: types.SetNull(types.StringType), PredefinedRole: types.BoolNull(), ResourceType: types.StringNull(), CreatedAt: types.Int64Null(), UpdatedAt: types.Int64Null(), CreatedBy: types.StringNull(), AssignmentSources: nil}
		if diags := plan.Set(ctx, &model); diags.HasError() {
			t.Fatalf("empty group-role state: %v", diags)
		}
	}
	return tfsdk.State{Schema: schema, Raw: plan.Raw}
}
