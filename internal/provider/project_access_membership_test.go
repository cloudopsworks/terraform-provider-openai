package provider

import (
	"context"
	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

type projectAccessClient struct {
	client.AdminClient
	user         *client.ProjectUser
	group        *client.ProjectGroup
	updatedRole  string
	deletedUser  bool
	deletedGroup bool
}

func (f *projectAccessClient) CreateProjectUser(_ context.Context, projectID string, req client.ProjectUserCreateRequest) (*client.ProjectUser, error) {
	f.user = &client.ProjectUser{ID: "user_1", Email: req.Email, Name: "User", Role: req.Role, AddedAt: 1}
	return f.user, nil
}
func (f *projectAccessClient) GetProjectUser(_ context.Context, _, _ string) (*client.ProjectUser, error) {
	return f.user, nil
}
func (f *projectAccessClient) ListProjectUsers(_ context.Context, _ string, _ client.ProjectUserListRequest) (*client.ProjectUserListResponse, error) {
	return &client.ProjectUserListResponse{Items: []client.ProjectUser{*f.user}, LastID: f.user.ID}, nil
}
func (f *projectAccessClient) UpdateProjectUser(_ context.Context, _, id string, req client.ProjectUserUpdateRequest) (*client.ProjectUser, error) {
	f.updatedRole = req.Role
	f.user = &client.ProjectUser{ID: id, Email: f.user.Email, Name: f.user.Name, Role: req.Role, AddedAt: f.user.AddedAt}
	return f.user, nil
}
func (f *projectAccessClient) DeleteProjectUser(_ context.Context, _, _ string) error {
	f.deletedUser = true
	return nil
}
func (f *projectAccessClient) CreateProjectGroup(_ context.Context, projectID string, req client.ProjectGroupCreateRequest) (*client.ProjectGroup, error) {
	f.group = &client.ProjectGroup{ID: projectID + "/" + req.GroupID, ProjectID: projectID, GroupID: req.GroupID, GroupName: "Engineering", GroupType: "group", CreatedAt: 1}
	return f.group, nil
}
func (f *projectAccessClient) GetProjectGroup(_ context.Context, _, _ string, _ client.ProjectGroupGetRequest) (*client.ProjectGroup, error) {
	return f.group, nil
}
func (f *projectAccessClient) ListProjectGroups(_ context.Context, _ string, _ client.ProjectGroupListRequest) (*client.ProjectGroupListResponse, error) {
	return &client.ProjectGroupListResponse{Items: []client.ProjectGroup{*f.group}, Next: ""}, nil
}
func (f *projectAccessClient) DeleteProjectGroup(_ context.Context, _, _ string) error {
	f.deletedGroup = true
	return nil
}
func TestProjectUserMembershipLifecycle(t *testing.T) {
	ctx := context.Background()
	fake := &projectAccessClient{}
	r := &projectUserResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	if _, ok := schema.Attributes["project_id"]; !ok {
		t.Fatal("project_id missing")
	}
	plan := tfsdk.Plan{Schema: schema}
	input := projectUserResourceModel{ProjectID: types.StringValue("proj_1"), Email: types.StringValue("user@example.com"), Role: types.StringValue("member")}
	if d := plan.Set(ctx, &input); d.HasError() {
		t.Fatal(d)
	}
	create := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	var state projectUserResourceModel
	if d := create.State.Get(ctx, &state); d.HasError() || state.ID.ValueString() != "user_1" {
		t.Fatalf("state=%#v d=%v", state, d)
	}
	state.Role = types.StringValue("owner")
	updatePlan := tfsdk.Plan{Schema: schema}
	if d := updatePlan.Set(ctx, &state); d.HasError() {
		t.Fatal(d)
	}
	update := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: updatePlan}, &update)
	if update.Diagnostics.HasError() || fake.updatedRole != "owner" {
		t.Fatalf("diagnostics=%v role=%q", update.Diagnostics, fake.updatedRole)
	}
	del := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &del)
	if del.Diagnostics.HasError() || !fake.deletedUser {
		t.Fatalf("diagnostics=%v deleted=%t", del.Diagnostics, fake.deletedUser)
	}
}
func TestProjectGroupMembershipLifecycle(t *testing.T) {
	ctx := context.Background()
	fake := &projectAccessClient{}
	r := &projectGroupResource{client: fake}
	schema := resourceSchema(ctx, t, r)
	if _, ok := schema.Attributes["group_type"]; !ok {
		t.Fatal("group_type missing")
	}
	plan := tfsdk.Plan{Schema: schema}
	input := projectGroupResourceModel{ProjectID: types.StringValue("proj_1"), GroupID: types.StringValue("grp_1"), Role: types.StringValue("member")}
	if d := plan.Set(ctx, &input); d.HasError() {
		t.Fatal(d)
	}
	create := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &create)
	if create.Diagnostics.HasError() {
		t.Fatal(create.Diagnostics)
	}
	var state projectGroupResourceModel
	if d := create.State.Get(ctx, &state); d.HasError() || state.ID.ValueString() != "proj_1/grp_1" || state.GroupType.ValueString() != "group" {
		t.Fatalf("state=%#v d=%v", state, d)
	}
	del := resource.DeleteResponse{State: create.State}
	r.Delete(ctx, resource.DeleteRequest{State: create.State}, &del)
	if del.Diagnostics.HasError() || !fake.deletedGroup {
		t.Fatalf("diagnostics=%v deleted=%t", del.Diagnostics, fake.deletedGroup)
	}
}

func TestProjectGroupImportIsRejectedBecauseRoleCannotBeRecovered(t *testing.T) {
	ctx := context.Background()
	response := resource.ImportStateResponse{State: tfsdk.State{Schema: resourceSchema(ctx, t, &projectGroupResource{})}}
	(&projectGroupResource{}).ImportState(ctx, resource.ImportStateRequest{ID: "proj_1/grp_1"}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("expected project group import to be rejected")
	}
}
