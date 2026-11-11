package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type partialServiceAccountClient struct {
	*fakeAdminClient
	partial *client.ServiceAccount
}

func (f *partialServiceAccountClient) CreateServiceAccount(context.Context, string, client.ServiceAccountCreateRequest) (*client.ServiceAccount, error) {
	return f.partial, errors.New("role assignment failed after creation")
}

func TestServiceAccountResourcePreservesPartialCreation(t *testing.T) {
	for _, createKey := range []bool{false, true} {
		t.Run(map[bool]string{false: "keyless", true: "bootstrap"}[createKey], func(t *testing.T) {
			ctx := context.Background()
			account := &client.ServiceAccount{ID: "sa_created", Name: "svc", Role: "none", CreatedAt: 11}
			if createKey {
				account.APIKey = &client.ServiceAccountAPIKeyCreateResponse{ID: "key_created", Name: "bootstrap", Value: "sk-test-only", CreatedAt: 12}
			}
			fake := &partialServiceAccountClient{fakeAdminClient: &fakeAdminClient{}, partial: account}
			r := &serviceAccountResource{client: fake}
			schema := resourceSchema(ctx, t, r)
			plan := tfsdk.Plan{Schema: schema}
			if d := plan.Set(ctx, &serviceAccountResourceModel{ProjectID: types.StringValue("proj_1"), Name: types.StringValue("svc"), Role: types.StringValue("member"), CreateAPIKey: types.BoolValue(createKey), Scopes: types.SetNull(types.StringType)}); d.HasError() {
				t.Fatal(d)
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected assignment diagnostic")
			}
			var state serviceAccountResourceModel
			if d := resp.State.Get(ctx, &state); d.HasError() {
				t.Fatal(d)
			}
			if state.ID.ValueString() != "sa_created" || state.ProjectID.ValueString() != "proj_1" || state.Role.ValueString() != "none" {
				t.Fatal("created identity was lost instead of remaining recoverable in state")
			}
			if state.CreateAPIKey.ValueBool() != createKey {
				t.Fatal("credential ownership changed")
			}
			if createKey && (state.APIKeyID.ValueString() != "key_created" || state.APIKeyValue.ValueString() != "sk-test-only") {
				t.Fatal("bootstrap credential was lost")
			}
			if !createKey && !state.APIKeyID.IsNull() {
				t.Fatal("keyless state retained a key")
			}
			if fake.deletedAccount != "" {
				t.Fatal("unexpected deletion of partially created identity")
			}
		})
	}
}
