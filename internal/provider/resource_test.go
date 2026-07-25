package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

type fakeAdminClient struct {
	project        *client.Project
	account        *client.ServiceAccount
	apiKey         *client.APIKey
	createdKey     *client.ServiceAccountAPIKeyCreateResponse
	deletedProject string
	deletedAccount string
	deletedKey     string
	err            error
}

func (f *fakeAdminClient) CreateProject(ctx context.Context, req client.ProjectCreateRequest) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.project = &client.Project{ID: "proj_1", Name: req.Name, ExternalKeyID: req.ExternalKeyID, Status: "active", CreatedAt: 10}
	return f.project, nil
}
func (f *fakeAdminClient) GetProject(ctx context.Context, id string) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.project, nil
}
func (f *fakeAdminClient) UpdateProject(ctx context.Context, id string, req client.ProjectUpdateRequest) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.project.Name = req.Name
	f.project.ExternalKeyID = req.ExternalKeyID
	return f.project, nil
}
func (f *fakeAdminClient) ArchiveProject(ctx context.Context, id string) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.deletedProject = id
	f.project.Status = "archived"
	f.project.ArchivedAt = 20
	return f.project, nil
}
func (f *fakeAdminClient) CreateServiceAccount(ctx context.Context, projectID string, req client.ServiceAccountCreateRequest) (*client.ServiceAccount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.account = &client.ServiceAccount{ID: "sa_1", Name: req.Name, Role: req.Role, CreatedAt: 11}
	return f.account, nil
}
func (f *fakeAdminClient) GetServiceAccount(ctx context.Context, projectID, serviceAccountID string) (*client.ServiceAccount, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.account, nil
}
func (f *fakeAdminClient) UpdateServiceAccount(ctx context.Context, projectID, serviceAccountID string, req client.ServiceAccountUpdateRequest) (*client.ServiceAccount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.account.Name = req.Name
	f.account.Role = req.Role
	return f.account, nil
}
func (f *fakeAdminClient) DeleteServiceAccount(ctx context.Context, projectID, serviceAccountID string) error {
	if f.err != nil {
		return f.err
	}
	f.deletedAccount = serviceAccountID
	return nil
}
func (f *fakeAdminClient) CreateServiceAccountAPIKey(ctx context.Context, projectID, serviceAccountID string, req client.ServiceAccountAPIKeyCreateRequest) (*client.ServiceAccountAPIKeyCreateResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.createdKey = &client.ServiceAccountAPIKeyCreateResponse{ID: "key_1", Name: req.Name, Value: "sk-created", CreatedAt: 12}
	return f.createdKey, nil
}
func (f *fakeAdminClient) GetProjectAPIKey(ctx context.Context, projectID, apiKeyID string) (*client.APIKey, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.apiKey, nil
}
func (f *fakeAdminClient) DeleteProjectAPIKey(ctx context.Context, projectID, apiKeyID string) error {
	if f.err != nil {
		return f.err
	}
	f.deletedKey = apiKeyID
	return nil
}

func TestProjectResourceLifecycleWithMockClient(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAdminClient{}
	r := &projectResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	plan := tfsdk.Plan{Schema: schema}
	respDiags := plan.Set(ctx, &projectResourceModel{Name: types.StringValue("App"), ExternalKeyID: types.StringValue("ext"), Geography: types.StringValue("eu")})
	if respDiags.HasError() {
		t.Fatalf("plan set: %v", respDiags)
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}
	var state projectResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("state get: %v", diags)
	}
	if state.ID.ValueString() != "proj_1" || state.Name.ValueString() != "App" {
		t.Fatalf("unexpected state: %#v", state)
	}

	readResp := resource.ReadResponse{State: createResp.State}
	r.Read(ctx, resource.ReadRequest{State: createResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	deleteResp := resource.DeleteResponse{State: readResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() || fake.deletedProject != "proj_1" {
		t.Fatalf("delete diagnostics=%v deleted=%q", deleteResp.Diagnostics, fake.deletedProject)
	}
}

func TestServiceAccountResourceLifecycleWithMockClient(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAdminClient{}
	r := &serviceAccountResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &serviceAccountResourceModel{ProjectID: types.StringValue("proj_1"), Name: types.StringValue("svc"), Role: types.StringValue("member")}); diags.HasError() {
		t.Fatalf("plan set: %v", diags)
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}
	var state serviceAccountResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("state get: %v", diags)
	}
	if state.ID.ValueString() != "sa_1" || state.Role.ValueString() != "member" {
		t.Fatalf("unexpected state: %#v", state)
	}

	state.Name = types.StringValue("svc2")
	plan2 := tfsdk.Plan{Schema: schema}
	if diags := plan2.Set(ctx, &state); diags.HasError() {
		t.Fatalf("plan2 set: %v", diags)
	}
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan2}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}
	deleteResp := resource.DeleteResponse{State: updateResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: updateResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() || fake.deletedAccount != "sa_1" {
		t.Fatalf("delete diagnostics=%v deleted=%q", deleteResp.Diagnostics, fake.deletedAccount)
	}
}

func TestProjectAPIKeyResourceLifecycleWithMockClient(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAdminClient{}
	r := &projectAPIKeyResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	scopes, diags := setStringValueOrNull(ctx, []string{"models.read"})
	if diags.HasError() {
		t.Fatalf("scopes: %v", diags)
	}
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &projectAPIKeyResourceModel{ProjectID: types.StringValue("proj_1"), ServiceAccountID: types.StringValue("sa_1"), Name: types.StringValue("svc-key"), Scopes: scopes}); diags.HasError() {
		t.Fatalf("plan set: %v", diags)
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}
	var state projectAPIKeyResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("state get: %v", diags)
	}
	if state.ID.ValueString() != "key_1" || state.Value.ValueString() != "sk-created" {
		t.Fatalf("unexpected state: %#v", state)
	}

	fake.apiKey = &client.APIKey{ID: "key_1", Name: "svc-key", RedactedValue: "sk-...", OwnerType: "service_account", OwnerID: "sa_1", OwnerProjectAccess: "active", CreatedAt: 12}
	readResp := resource.ReadResponse{State: createResp.State}
	r.Read(ctx, resource.ReadRequest{State: createResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	var refreshed projectAPIKeyResourceModel
	if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
		t.Fatalf("refreshed get: %v", diags)
	}
	if refreshed.Value.ValueString() != "sk-created" || refreshed.RedactedValue.ValueString() != "sk-..." {
		t.Fatalf("secret not preserved/read metadata missing: %#v", refreshed)
	}
	deleteResp := resource.DeleteResponse{State: readResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() || fake.deletedKey != "key_1" {
		t.Fatalf("delete diagnostics=%v deleted=%q", deleteResp.Diagnostics, fake.deletedKey)
	}
}

func TestResourceClientErrorsAreDiagnostics(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAdminClient{err: errors.New("boom")}
	r := &projectResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &projectResourceModel{Name: types.StringValue("App")}); diags.HasError() {
		t.Fatalf("plan set: %v", diags)
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if !createResp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic")
	}
}

func resourceSchema(ctx context.Context, t *testing.T, r resource.Resource) resourceschema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}
