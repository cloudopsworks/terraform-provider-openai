package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/openai/openai-go/v3"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

type fakeAdminClient struct {
	project                 *client.Project
	projects                []client.Project
	readProjectID           string
	listProjectsReqs        []client.ProjectListRequest
	account                 *client.ServiceAccount
	accounts                []client.ServiceAccount
	listServiceAccountReqs  []serviceAccountListCall
	apiKey                  *client.APIKey
	createdKey              *client.ServiceAccountAPIKeyCreateResponse
	createdProjectReq       client.ProjectCreateRequest
	updatedProjectID        string
	updatedProjectReq       client.ProjectUpdateRequest
	createdAccountProjectID string
	createdAccountReq       client.ServiceAccountCreateRequest
	readAccountProjectID    string
	readAccountID           string
	updatedAccountProjectID string
	updatedAccountID        string
	updatedAccountReq       client.ServiceAccountUpdateRequest
	deletedProject          string
	archiveProjectErr       error
	deletedAccount          string
	deletedKey              string
	err                     error
}

type serviceAccountListCall struct {
	projectID string
	req       client.ServiceAccountListRequest
}

func (f *fakeAdminClient) CreateProject(ctx context.Context, req client.ProjectCreateRequest) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.createdProjectReq = req
	f.project = &client.Project{ID: "proj_1", Name: req.Name, ExternalKeyID: req.ExternalKeyID, Status: "active", CreatedAt: 10}
	return f.project, nil
}
func (f *fakeAdminClient) GetProject(ctx context.Context, id string) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.readProjectID = id
	if f.project != nil && f.project.ID == id {
		return f.project, nil
	}
	for i := range f.projects {
		if f.projects[i].ID == id {
			return &f.projects[i], nil
		}
	}
	return f.project, nil
}
func (f *fakeAdminClient) ListProjects(ctx context.Context, req client.ProjectListRequest) (*client.ProjectListResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.listProjectsReqs = append(f.listProjectsReqs, req)
	projects := append([]client.Project(nil), f.projects...)
	if len(projects) == 0 && f.project != nil {
		projects = append(projects, *f.project)
	}
	if len(projects) == 0 {
		return &client.ProjectListResponse{}, nil
	}
	return &client.ProjectListResponse{Items: projects, LastID: projects[len(projects)-1].ID}, nil
}
func (f *fakeAdminClient) UpdateProject(ctx context.Context, id string, req client.ProjectUpdateRequest) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.updatedProjectID = id
	f.updatedProjectReq = req
	f.project.Name = req.Name
	f.project.ExternalKeyID = req.ExternalKeyID
	return f.project, nil
}
func (f *fakeAdminClient) ArchiveProject(ctx context.Context, id string) (*client.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.archiveProjectErr != nil {
		return nil, f.archiveProjectErr
	}
	f.deletedProject = id
	if f.project == nil {
		return nil, nil
	}
	f.project.Status = "archived"
	f.project.ArchivedAt = 20
	return f.project, nil
}
func (f *fakeAdminClient) CreateServiceAccount(ctx context.Context, projectID string, req client.ServiceAccountCreateRequest) (*client.ServiceAccount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.createdAccountProjectID = projectID
	f.createdAccountReq = req
	f.account = &client.ServiceAccount{ID: "sa_1", Name: req.Name, Role: req.Role, CreatedAt: 11}
	return f.account, nil
}
func (f *fakeAdminClient) GetServiceAccount(ctx context.Context, projectID, serviceAccountID string) (*client.ServiceAccount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.readAccountProjectID = projectID
	f.readAccountID = serviceAccountID
	if f.account == nil || projectID != f.createdAccountProjectID || serviceAccountID != f.account.ID {
		return nil, errors.New("mock service account not found")
	}
	return f.account, nil
}
func (f *fakeAdminClient) ListServiceAccounts(ctx context.Context, projectID string, req client.ServiceAccountListRequest) (*client.ServiceAccountListResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.listServiceAccountReqs = append(f.listServiceAccountReqs, serviceAccountListCall{projectID: projectID, req: req})
	accounts := append([]client.ServiceAccount(nil), f.accounts...)
	if len(accounts) == 0 && f.account != nil {
		accounts = append(accounts, *f.account)
	}
	if len(accounts) == 0 {
		return &client.ServiceAccountListResponse{}, nil
	}
	return &client.ServiceAccountListResponse{Items: accounts, LastID: accounts[len(accounts)-1].ID}, nil
}
func (f *fakeAdminClient) UpdateServiceAccount(ctx context.Context, projectID, serviceAccountID string, req client.ServiceAccountUpdateRequest) (*client.ServiceAccount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.updatedAccountProjectID = projectID
	f.updatedAccountID = serviceAccountID
	f.updatedAccountReq = req
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
	if state.ID.ValueString() != "proj_1" || state.Name.ValueString() != "App" || state.ExternalKeyID.ValueString() != "ext" {
		t.Fatalf("unexpected state: %#v", state)
	}
	if state.Geography.ValueString() != "eu" {
		t.Fatalf("geography was not preserved in state: %#v", state)
	}
	if state.Status.ValueString() != "active" || state.CreatedAt.ValueInt64() != 10 || !state.ArchivedAt.IsNull() {
		t.Fatalf("computed project attributes were not mapped: %#v", state)
	}
	if fake.createdProjectReq != (client.ProjectCreateRequest{Name: "App", ExternalKeyID: "ext", Geography: "eu"}) {
		t.Fatalf("unexpected create request: %#v", fake.createdProjectReq)
	}

	state.Name = types.StringValue("App 2")
	state.ExternalKeyID = types.StringValue("ext2")
	plan2 := tfsdk.Plan{Schema: schema}
	if diags := plan2.Set(ctx, &state); diags.HasError() {
		t.Fatalf("plan2 set: %v", diags)
	}
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan2}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}
	if fake.updatedProjectID != "proj_1" || fake.updatedProjectReq != (client.ProjectUpdateRequest{Name: "App 2", ExternalKeyID: "ext2", Geography: "eu"}) {
		t.Fatalf("unexpected update id=%q request=%#v", fake.updatedProjectID, fake.updatedProjectReq)
	}

	readResp := resource.ReadResponse{State: updateResp.State}
	r.Read(ctx, resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	var refreshed projectResourceModel
	if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
		t.Fatalf("refreshed get: %v", diags)
	}
	if refreshed.Name.ValueString() != "App 2" || refreshed.ExternalKeyID.ValueString() != "ext2" || refreshed.Geography.ValueString() != "eu" {
		t.Fatalf("project read state mapping failed: %#v", refreshed)
	}
	deleteResp := resource.DeleteResponse{State: readResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: readResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() || fake.deletedProject != "proj_1" {
		t.Fatalf("delete diagnostics=%v deleted=%q", deleteResp.Diagnostics, fake.deletedProject)
	}
}

func TestProjectModelFromAPIMapsProjectResponseToTerraformState(t *testing.T) {
	ctx := context.Background()
	schema := resourceSchema(ctx, t, NewProjectResource())
	prior := projectResourceModel{
		Geography: types.StringValue("eu"),
	}

	mapped := projectModelFromAPI(&client.Project{
		ID:            "proj_123",
		Name:          "Platform",
		ExternalKeyID: "external-123",
		Status:        "archived",
		CreatedAt:     1710000000,
		ArchivedAt:    1720000000,
	}, prior)

	state := tfsdk.State{Schema: schema}
	if diags := state.Set(ctx, &mapped); diags.HasError() {
		t.Fatalf("state set: %v", diags)
	}

	var got projectResourceModel
	if diags := state.Get(ctx, &got); diags.HasError() {
		t.Fatalf("state get: %v", diags)
	}

	if got.ID.ValueString() != "proj_123" ||
		got.Name.ValueString() != "Platform" ||
		got.ExternalKeyID.ValueString() != "external-123" ||
		got.Geography.ValueString() != "eu" ||
		got.Status.ValueString() != "archived" ||
		got.CreatedAt.ValueInt64() != 1710000000 ||
		got.ArchivedAt.ValueInt64() != 1720000000 {
		t.Fatalf("project API response was not mapped to Terraform state: %#v", got)
	}
}

func TestProjectResourceDeleteWithMockClient(t *testing.T) {
	ctx := context.Background()
	schema := resourceSchema(ctx, t, &projectResource{})
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(ctx, &projectResourceModel{ID: types.StringValue("proj_1"), Name: types.StringValue("App")}); diags.HasError() {
		t.Fatalf("state set: %v", diags)
	}

	tests := map[string]struct {
		fake             *fakeAdminClient
		wantDiagnostic   bool
		wantArchiveCall  bool
		wantResourceGone bool
	}{
		"archives project and removes state": {
			fake:             &fakeAdminClient{project: &client.Project{ID: "proj_1", Name: "App", Status: "active"}},
			wantArchiveCall:  true,
			wantResourceGone: true,
		},
		"treats missing project as deleted": {
			fake:             &fakeAdminClient{archiveProjectErr: &openai.Error{StatusCode: http.StatusNotFound, Message: "not found"}},
			wantResourceGone: true,
		},
		"keeps state when archive fails": {
			fake:           &fakeAdminClient{archiveProjectErr: errors.New("boom")},
			wantDiagnostic: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := &projectResource{client: tt.fake}
			deleteResp := resource.DeleteResponse{State: state}

			r.Delete(ctx, resource.DeleteRequest{State: state}, &deleteResp)

			if deleteResp.Diagnostics.HasError() != tt.wantDiagnostic {
				t.Fatalf("diagnostics=%v, wantDiagnostic=%t", deleteResp.Diagnostics, tt.wantDiagnostic)
			}
			if (tt.fake.deletedProject == "proj_1") != tt.wantArchiveCall {
				t.Fatalf("archive call deleted=%q, want call=%t", tt.fake.deletedProject, tt.wantArchiveCall)
			}
			if deleteResp.State.Raw.IsNull() != tt.wantResourceGone {
				t.Fatalf("state removed=%t, want=%t", deleteResp.State.Raw.IsNull(), tt.wantResourceGone)
			}
		})
	}
}

func TestProjectDataSourcesWithMockClient(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAdminClient{projects: []client.Project{
		{ID: "proj_1", Name: "App", Status: "active", CreatedAt: 10},
		{ID: "proj_2", Name: "Archive", ExternalKeyID: "ext_2", Status: "archived", CreatedAt: 11, ArchivedAt: 20},
	}}

	listDS := &projectsDataSource{client: fake}
	listSchema := dataSourceSchema(ctx, t, listDS)
	listConfigPlan := tfsdk.Plan{Schema: listSchema}
	if diags := listConfigPlan.Set(ctx, &projectsDataSourceModel{
		After:           types.StringValue("proj_0"),
		Limit:           types.Int64Value(2),
		IncludeArchived: types.BoolValue(true),
	}); diags.HasError() {
		t.Fatalf("list config set: %v", diags)
	}
	listResp := datasource.ReadResponse{State: tfsdk.State{Schema: listSchema}}
	listDS.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: listSchema, Raw: listConfigPlan.Raw}}, &listResp)
	if listResp.Diagnostics.HasError() {
		t.Fatalf("list diagnostics: %v", listResp.Diagnostics)
	}
	if len(fake.listProjectsReqs) != 1 || fake.listProjectsReqs[0] != (client.ProjectListRequest{After: "proj_0", Limit: 2, IncludeArchived: true}) {
		t.Fatalf("unexpected list requests: %#v", fake.listProjectsReqs)
	}
	var listState projectsDataSourceModel
	if diags := listResp.State.Get(ctx, &listState); diags.HasError() {
		t.Fatalf("list state get: %v", diags)
	}
	if len(listState.Items) != 2 || listState.Items[1].ID.ValueString() != "proj_2" || listState.LastID.ValueString() != "proj_2" {
		t.Fatalf("unexpected list state: %#v", listState)
	}

	singleDS := &projectDataSource{client: fake}
	singleSchema := dataSourceSchema(ctx, t, singleDS)
	singleConfigPlan := tfsdk.Plan{Schema: singleSchema}
	if diags := singleConfigPlan.Set(ctx, &projectDataSourceModel{ID: types.StringValue("proj_2")}); diags.HasError() {
		t.Fatalf("single config set: %v", diags)
	}
	singleResp := datasource.ReadResponse{State: tfsdk.State{Schema: singleSchema}}
	singleDS.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: singleSchema, Raw: singleConfigPlan.Raw}}, &singleResp)
	if singleResp.Diagnostics.HasError() {
		t.Fatalf("single diagnostics: %v", singleResp.Diagnostics)
	}
	if len(fake.listProjectsReqs) != 1 {
		t.Fatalf("single lookup should not list projects: %#v", fake.listProjectsReqs)
	}
	if fake.readProjectID != "proj_2" {
		t.Fatalf("single lookup read project id = %q, want proj_2", fake.readProjectID)
	}
	var singleState projectDataSourceModel
	if diags := singleResp.State.Get(ctx, &singleState); diags.HasError() {
		t.Fatalf("single state get: %v", diags)
	}
	if singleState.ID.ValueString() != "proj_2" ||
		singleState.Name.ValueString() != "Archive" ||
		singleState.ExternalKeyID.ValueString() != "ext_2" ||
		singleState.Status.ValueString() != "archived" ||
		singleState.CreatedAt.ValueInt64() != 11 ||
		singleState.ArchivedAt.ValueInt64() != 20 {
		t.Fatalf("unexpected single state: %#v", singleState)
	}

	missingIDConfigPlan := tfsdk.Plan{Schema: singleSchema}
	if diags := missingIDConfigPlan.Set(ctx, &projectDataSourceModel{ID: types.StringNull()}); diags.HasError() {
		t.Fatalf("missing id config set: %v", diags)
	}
	missingIDResp := datasource.ReadResponse{State: tfsdk.State{Schema: singleSchema}}
	singleDS.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: singleSchema, Raw: missingIDConfigPlan.Raw}}, &missingIDResp)
	if !missingIDResp.Diagnostics.HasError() {
		t.Fatal("expected missing id diagnostics")
	}
	if len(fake.listProjectsReqs) != 1 {
		t.Fatalf("missing id should not list projects: %#v", fake.listProjectsReqs)
	}
}

func TestProjectDataSourceSchemaContract(t *testing.T) {
	ctx := context.Background()
	schema := dataSourceSchema(ctx, t, NewProjectDataSource())

	stringAttr := func(name string) datasourceschema.StringAttribute {
		t.Helper()
		attr, ok := schema.Attributes[name].(datasourceschema.StringAttribute)
		if !ok {
			t.Fatalf("%s was not a string attribute: %T", name, schema.Attributes[name])
		}
		return attr
	}
	int64Attr := func(name string) datasourceschema.Int64Attribute {
		t.Helper()
		attr, ok := schema.Attributes[name].(datasourceschema.Int64Attribute)
		if !ok {
			t.Fatalf("%s was not an int64 attribute: %T", name, schema.Attributes[name])
		}
		return attr
	}

	if attr := stringAttr("id"); !attr.Required || attr.Optional || attr.Computed {
		t.Fatalf("id schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("name"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("name schema flags mismatch: %#v", attr)
	}
	if _, ok := schema.Attributes["include_archived"]; ok {
		t.Fatal("singular project data source must not expose include_archived")
	}
	if attr := stringAttr("external_key_id"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("external_key_id schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("status"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("status schema flags mismatch: %#v", attr)
	}
	if attr := int64Attr("created_at"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("created_at schema flags mismatch: %#v", attr)
	}
	if attr := int64Attr("archived_at"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("archived_at schema flags mismatch: %#v", attr)
	}
}

func TestProjectResourceSchemaContract(t *testing.T) {
	ctx := context.Background()
	schema := resourceSchema(ctx, t, NewProjectResource())

	stringAttr := func(name string) resourceschema.StringAttribute {
		t.Helper()
		attr, ok := schema.Attributes[name].(resourceschema.StringAttribute)
		if !ok {
			t.Fatalf("%s was not a string attribute: %T", name, schema.Attributes[name])
		}
		return attr
	}
	int64Attr := func(name string) resourceschema.Int64Attribute {
		t.Helper()
		attr, ok := schema.Attributes[name].(resourceschema.Int64Attribute)
		if !ok {
			t.Fatalf("%s was not an int64 attribute: %T", name, schema.Attributes[name])
		}
		return attr
	}

	if attr := stringAttr("id"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("id schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("name"); !attr.Required || attr.Optional || attr.Computed {
		t.Fatalf("name schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("external_key_id"); !attr.Optional || !attr.Computed || attr.Required {
		t.Fatalf("external_key_id schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("geography"); !attr.Optional || attr.Required || attr.Computed || len(attr.PlanModifiers) != 1 {
		t.Fatalf("geography schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("status"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("status schema flags mismatch: %#v", attr)
	}
	if attr := int64Attr("created_at"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("created_at schema flags mismatch: %#v", attr)
	}
	if attr := int64Attr("archived_at"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("archived_at schema flags mismatch: %#v", attr)
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
	if state.ProjectID.ValueString() != "proj_1" || state.Name.ValueString() != "svc" || state.CreatedAt.ValueInt64() != 11 {
		t.Fatalf("service account state mapping failed: %#v", state)
	}
	if fake.createdAccountProjectID != "proj_1" || fake.createdAccountReq != (client.ServiceAccountCreateRequest{Name: "svc", Role: "member"}) {
		t.Fatalf("unexpected create project=%q request=%#v", fake.createdAccountProjectID, fake.createdAccountReq)
	}

	state.Name = types.StringValue("svc2")
	state.Role = types.StringValue("owner")
	plan2 := tfsdk.Plan{Schema: schema}
	if diags := plan2.Set(ctx, &state); diags.HasError() {
		t.Fatalf("plan2 set: %v", diags)
	}
	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan2}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}
	if fake.updatedAccountProjectID != "proj_1" || fake.updatedAccountID != "sa_1" || fake.updatedAccountReq != (client.ServiceAccountUpdateRequest{Name: "svc2", Role: "owner"}) {
		t.Fatalf("unexpected update project=%q id=%q request=%#v", fake.updatedAccountProjectID, fake.updatedAccountID, fake.updatedAccountReq)
	}

	readResp := resource.ReadResponse{State: updateResp.State}
	r.Read(ctx, resource.ReadRequest{State: updateResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	if fake.readAccountProjectID != "proj_1" || fake.readAccountID != "sa_1" {
		t.Fatalf("unexpected read project=%q id=%q", fake.readAccountProjectID, fake.readAccountID)
	}
	var refreshed serviceAccountResourceModel
	if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
		t.Fatalf("refreshed get: %v", diags)
	}
	if refreshed.ID.ValueString() != "sa_1" || refreshed.Name.ValueString() != "svc2" || refreshed.Role.ValueString() != "owner" || refreshed.ProjectID.ValueString() != "proj_1" || refreshed.CreatedAt.ValueInt64() != 11 {
		t.Fatalf("service account read state mapping failed: %#v", refreshed)
	}

	importResp := resource.ImportStateResponse{State: emptyServiceAccountState(ctx, t, schema)}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "proj_1/sa_1"}, &importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("import diagnostics: %v", importResp.Diagnostics)
	}
	importReadResp := resource.ReadResponse{State: importResp.State}
	r.Read(ctx, resource.ReadRequest{State: importResp.State}, &importReadResp)
	if importReadResp.Diagnostics.HasError() {
		t.Fatalf("import read diagnostics: %v", importReadResp.Diagnostics)
	}
	var imported serviceAccountResourceModel
	if diags := importReadResp.State.Get(ctx, &imported); diags.HasError() {
		t.Fatalf("imported get: %v", diags)
	}
	if imported.ID.ValueString() != "sa_1" || imported.ProjectID.ValueString() != "proj_1" || imported.Name.ValueString() != "svc2" || imported.Role.ValueString() != "owner" {
		t.Fatalf("service account import/read state mapping failed: %#v", imported)
	}

	deleteResp := resource.DeleteResponse{State: importReadResp.State}
	r.Delete(ctx, resource.DeleteRequest{State: importReadResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() || fake.deletedAccount != "sa_1" {
		t.Fatalf("delete diagnostics=%v deleted=%q", deleteResp.Diagnostics, fake.deletedAccount)
	}
}

func TestServiceAccountResourceReadDriftAndNotFoundWithMockClient(t *testing.T) {
	ctx := context.Background()
	schema := resourceSchema(ctx, t, &serviceAccountResource{})
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(ctx, &serviceAccountResourceModel{
		ID:        types.StringValue("sa_1"),
		ProjectID: types.StringValue("proj_1"),
		Name:      types.StringValue("svc"),
		Role:      types.StringValue("member"),
		CreatedAt: types.Int64Value(11),
	}); diags.HasError() {
		t.Fatalf("state set: %v", diags)
	}

	t.Run("refreshes drifted remote attributes", func(t *testing.T) {
		fake := &fakeAdminClient{
			createdAccountProjectID: "proj_1",
			account:                 &client.ServiceAccount{ID: "sa_1", Name: "svc-drift", Role: "owner", CreatedAt: 12},
		}
		r := &serviceAccountResource{client: fake}
		readResp := resource.ReadResponse{State: state}

		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

		if readResp.Diagnostics.HasError() {
			t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
		}
		var refreshed serviceAccountResourceModel
		if diags := readResp.State.Get(ctx, &refreshed); diags.HasError() {
			t.Fatalf("refreshed get: %v", diags)
		}
		if refreshed.Name.ValueString() != "svc-drift" || refreshed.Role.ValueString() != "owner" || refreshed.CreatedAt.ValueInt64() != 12 {
			t.Fatalf("drifted service account was not refreshed: %#v", refreshed)
		}
	})

	t.Run("removes state when remote service account is not found", func(t *testing.T) {
		fake := &fakeAdminClient{err: &openai.Error{StatusCode: http.StatusNotFound, Message: "not found"}}
		r := &serviceAccountResource{client: fake}
		readResp := resource.ReadResponse{State: state}

		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

		if readResp.Diagnostics.HasError() {
			t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
		}
		if !readResp.State.Raw.IsNull() {
			t.Fatalf("expected service account state to be removed after not found, got: %s", readResp.State.Raw.String())
		}
	})
}

func TestServiceAccountResourceDeleteWithMockClient(t *testing.T) {
	ctx := context.Background()
	schema := resourceSchema(ctx, t, &serviceAccountResource{})
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(ctx, &serviceAccountResourceModel{
		ID:        types.StringValue("sa_1"),
		ProjectID: types.StringValue("proj_1"),
		Name:      types.StringValue("svc"),
		Role:      types.StringValue("member"),
		CreatedAt: types.Int64Value(11),
	}); diags.HasError() {
		t.Fatalf("state set: %v", diags)
	}

	tests := map[string]struct {
		fake             *fakeAdminClient
		wantDiagnostic   bool
		wantDeleteCall   bool
		wantResourceGone bool
	}{
		"deletes service account and removes state": {
			fake:             &fakeAdminClient{},
			wantDeleteCall:   true,
			wantResourceGone: true,
		},
		"treats missing service account as deleted": {
			fake:             &fakeAdminClient{err: &openai.Error{StatusCode: http.StatusNotFound, Message: "not found"}},
			wantResourceGone: true,
		},
		"keeps state when delete fails": {
			fake:           &fakeAdminClient{err: errors.New("boom")},
			wantDiagnostic: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := &serviceAccountResource{client: tt.fake}
			deleteResp := resource.DeleteResponse{State: state}

			r.Delete(ctx, resource.DeleteRequest{State: state}, &deleteResp)

			if deleteResp.Diagnostics.HasError() != tt.wantDiagnostic {
				t.Fatalf("diagnostics=%v, wantDiagnostic=%t", deleteResp.Diagnostics, tt.wantDiagnostic)
			}
			if (tt.fake.deletedAccount == "sa_1") != tt.wantDeleteCall {
				t.Fatalf("delete call deleted=%q, want call=%t", tt.fake.deletedAccount, tt.wantDeleteCall)
			}
			if deleteResp.State.Raw.IsNull() != tt.wantResourceGone {
				t.Fatalf("state removed=%t, want=%t", deleteResp.State.Raw.IsNull(), tt.wantResourceGone)
			}
		})
	}
}

func TestServiceAccountDataSourcesWithMockClient(t *testing.T) {
	ctx := context.Background()
	fake := &fakeAdminClient{accounts: []client.ServiceAccount{
		{ID: "sa_1", Name: "svc", Role: "member", CreatedAt: 11},
		{ID: "sa_2", Name: "ops", Role: "owner", CreatedAt: 12},
	}}

	listDS := &serviceAccountsDataSource{client: fake}
	listSchema := dataSourceSchema(ctx, t, listDS)
	listConfigPlan := tfsdk.Plan{Schema: listSchema}
	if diags := listConfigPlan.Set(ctx, &serviceAccountsDataSourceModel{
		ProjectID: types.StringValue("proj_1"),
		After:     types.StringValue("sa_0"),
		Limit:     types.Int64Value(2),
	}); diags.HasError() {
		t.Fatalf("list config set: %v", diags)
	}
	listResp := datasource.ReadResponse{State: tfsdk.State{Schema: listSchema}}
	listDS.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: listSchema, Raw: listConfigPlan.Raw}}, &listResp)
	if listResp.Diagnostics.HasError() {
		t.Fatalf("list diagnostics: %v", listResp.Diagnostics)
	}
	if len(fake.listServiceAccountReqs) != 1 || fake.listServiceAccountReqs[0] != (serviceAccountListCall{projectID: "proj_1", req: client.ServiceAccountListRequest{After: "sa_0", Limit: 2}}) {
		t.Fatalf("unexpected list requests: %#v", fake.listServiceAccountReqs)
	}
	var listState serviceAccountsDataSourceModel
	if diags := listResp.State.Get(ctx, &listState); diags.HasError() {
		t.Fatalf("list state get: %v", diags)
	}
	if len(listState.Items) != 2 || listState.Items[1].ID.ValueString() != "sa_2" || listState.LastID.ValueString() != "sa_2" {
		t.Fatalf("unexpected list state: %#v", listState)
	}

	singleDS := &serviceAccountDataSource{client: fake}
	singleSchema := dataSourceSchema(ctx, t, singleDS)
	singleConfigPlan := tfsdk.Plan{Schema: singleSchema}
	if diags := singleConfigPlan.Set(ctx, &serviceAccountDataSourceModel{ProjectID: types.StringValue("proj_1"), Name: types.StringValue("ops")}); diags.HasError() {
		t.Fatalf("single config set: %v", diags)
	}
	singleResp := datasource.ReadResponse{State: tfsdk.State{Schema: singleSchema}}
	singleDS.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: singleSchema, Raw: singleConfigPlan.Raw}}, &singleResp)
	if singleResp.Diagnostics.HasError() {
		t.Fatalf("single diagnostics: %v", singleResp.Diagnostics)
	}
	if len(fake.listServiceAccountReqs) != 2 || fake.listServiceAccountReqs[1] != (serviceAccountListCall{projectID: "proj_1", req: client.ServiceAccountListRequest{Limit: 100}}) {
		t.Fatalf("unexpected lookup list requests: %#v", fake.listServiceAccountReqs)
	}
	var singleState serviceAccountDataSourceModel
	if diags := singleResp.State.Get(ctx, &singleState); diags.HasError() {
		t.Fatalf("single state get: %v", diags)
	}
	if singleState.ID.ValueString() != "sa_2" || singleState.Role.ValueString() != "owner" {
		t.Fatalf("unexpected single state: %#v", singleState)
	}

	fake.account = &client.ServiceAccount{ID: "sa_1", Name: "svc", Role: "member", CreatedAt: 11}
	fake.createdAccountProjectID = "proj_1"
	idConfigPlan := tfsdk.Plan{Schema: singleSchema}
	if diags := idConfigPlan.Set(ctx, &serviceAccountDataSourceModel{ProjectID: types.StringValue("proj_1"), ID: types.StringValue("sa_1")}); diags.HasError() {
		t.Fatalf("id config set: %v", diags)
	}
	idResp := datasource.ReadResponse{State: tfsdk.State{Schema: singleSchema}}
	singleDS.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: singleSchema, Raw: idConfigPlan.Raw}}, &idResp)
	if idResp.Diagnostics.HasError() {
		t.Fatalf("id diagnostics: %v", idResp.Diagnostics)
	}
	if len(fake.listServiceAccountReqs) != 2 {
		t.Fatalf("id lookup should not list service accounts: %#v", fake.listServiceAccountReqs)
	}
	if fake.readAccountProjectID != "proj_1" || fake.readAccountID != "sa_1" {
		t.Fatalf("unexpected read project=%q id=%q", fake.readAccountProjectID, fake.readAccountID)
	}
}

func TestServiceAccountResourceSchemaContract(t *testing.T) {
	ctx := context.Background()
	r := NewServiceAccountResource()
	schema := resourceSchema(ctx, t, r)
	if schema.Version != 1 {
		t.Fatalf("schema version = %d, want 1", schema.Version)
	}
	if _, ok := r.(resource.ResourceWithUpgradeState); !ok {
		t.Fatal("service account resource must implement ResourceWithUpgradeState")
	}

	stringAttr := func(name string) resourceschema.StringAttribute {
		t.Helper()
		attr, ok := schema.Attributes[name].(resourceschema.StringAttribute)
		if !ok {
			t.Fatalf("%s was not a string attribute: %T", name, schema.Attributes[name])
		}
		return attr
	}
	int64Attr := func(name string) resourceschema.Int64Attribute {
		t.Helper()
		attr, ok := schema.Attributes[name].(resourceschema.Int64Attribute)
		if !ok {
			t.Fatalf("%s was not an int64 attribute: %T", name, schema.Attributes[name])
		}
		return attr
	}

	if attr := stringAttr("id"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("id schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("project_id"); !attr.Required || attr.Optional || attr.Computed || len(attr.PlanModifiers) != 1 {
		t.Fatalf("project_id schema flags mismatch: %#v", attr)
	}
	if attr := stringAttr("name"); !attr.Required || attr.Optional || attr.Computed {
		t.Fatalf("name schema flags mismatch: %#v", attr)
	}
	role := stringAttr("role")
	if !role.Optional || !role.Computed || role.Required || role.Default == nil || len(role.Validators) != 1 {
		t.Fatalf("role schema flags mismatch: %#v", role)
	}
	validationResp := validator.StringResponse{}
	role.Validators[0].ValidateString(ctx, validator.StringRequest{Path: path.Root("role"), ConfigValue: types.StringValue("reader")}, &validationResp)
	if !validationResp.Diagnostics.HasError() {
		t.Fatal("expected role validator diagnostic")
	}
	if attr := int64Attr("created_at"); !attr.Computed || attr.Required || attr.Optional {
		t.Fatalf("created_at schema flags mismatch: %#v", attr)
	}
}

func TestServiceAccountResourceImportState(t *testing.T) {
	ctx := context.Background()
	r := &serviceAccountResource{}
	schema := resourceSchema(ctx, t, r)
	resp := resource.ImportStateResponse{State: emptyServiceAccountState(ctx, t, schema)}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "proj_1/sa_1"}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("import diagnostics: %v", resp.Diagnostics)
	}
	var state serviceAccountResourceModel
	if diags := resp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("state get: %v", diags)
	}
	if state.ProjectID.ValueString() != "proj_1" || state.ID.ValueString() != "sa_1" {
		t.Fatalf("unexpected import state: %#v", state)
	}

	badResp := resource.ImportStateResponse{State: emptyServiceAccountState(ctx, t, schema)}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "proj_1"}, &badResp)
	if !badResp.Diagnostics.HasError() {
		t.Fatal("expected invalid import diagnostic")
	}
}

func TestServiceAccountResourceUpgradeStatePreservesV0State(t *testing.T) {
	ctx := context.Background()
	r := &serviceAccountResource{}
	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("missing v0 state upgrader")
	}
	if upgrader.PriorSchema == nil || upgrader.PriorSchema.Version != 0 {
		t.Fatalf("unexpected prior schema: %#v", upgrader.PriorSchema)
	}
	prior := tfsdk.State{Schema: *upgrader.PriorSchema}
	if diags := prior.Set(ctx, &serviceAccountResourceModel{
		ID:        types.StringValue("sa_1"),
		ProjectID: types.StringValue("proj_1"),
		Name:      types.StringValue("svc"),
		Role:      types.StringValue("owner"),
		CreatedAt: types.Int64Value(11),
	}); diags.HasError() {
		t.Fatalf("prior set: %v", diags)
	}

	resp := resource.UpgradeStateResponse{}
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &prior}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade diagnostics: %v", resp.Diagnostics)
	}
	var upgraded serviceAccountResourceModel
	if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("upgraded get: %v", diags)
	}
	if upgraded.ID.ValueString() != "sa_1" || upgraded.ProjectID.ValueString() != "proj_1" || upgraded.Name.ValueString() != "svc" || upgraded.Role.ValueString() != "owner" || upgraded.CreatedAt.ValueInt64() != 11 {
		t.Fatalf("unexpected upgraded state: %#v", upgraded)
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

func dataSourceSchema(ctx context.Context, t *testing.T, ds datasource.DataSource) datasourceschema.Schema {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(ctx, datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}
