package provider

import (
	"context"
	"testing"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProjectAccessRoleResourceSchemas(t *testing.T) {
	ctx := context.Background()
	for name, r := range map[string]resource.Resource{
		"user":  NewProjectUserRoleResource(),
		"group": NewProjectGroupRoleResource(),
	} {
		t.Run(name, func(t *testing.T) {
			schema := resourceSchema(ctx, t, r)
			for _, key := range []string{"project_id", name + "_id", "role_id"} {
				attribute, ok := schema.Attributes[key].(resourceschema.StringAttribute)
				if !ok || !attribute.Required || attribute.Computed || len(attribute.PlanModifiers) != 1 {
					t.Fatalf("%s schema mismatch: %#v", key, schema.Attributes[key])
				}
			}
		})
	}
}

func TestProjectAccessRoleDataSourceSchemas(t *testing.T) {
	ctx := context.Background()
	for name, d := range map[string]datasource.DataSource{
		"user":  NewProjectUserRoleDataSource(),
		"group": NewProjectGroupRoleDataSource(),
	} {
		t.Run(name, func(t *testing.T) {
			schema := dataSourceSchema(ctx, t, d)
			for _, key := range []string{"project_id", name + "_id", "role_id"} {
				attribute, ok := schema.Attributes[key].(datasourceschema.StringAttribute)
				if !ok || !attribute.Required || attribute.Computed {
					t.Fatalf("%s schema mismatch: %#v", key, schema.Attributes[key])
				}
			}
		})
	}
}

func TestProjectAccessRoleListDataSourceSchemas(t *testing.T) {
	ctx := context.Background()
	for name, d := range map[string]datasource.DataSource{
		"user":  NewProjectUserRolesDataSource(),
		"group": NewProjectGroupRolesDataSource(),
	} {
		t.Run(name, func(t *testing.T) {
			schema := dataSourceSchema(ctx, t, d)
			for _, key := range []string{"project_id", name + "_id"} {
				attribute, ok := schema.Attributes[key].(datasourceschema.StringAttribute)
				if !ok || !attribute.Required || attribute.Computed {
					t.Fatalf("%s schema mismatch: %#v", key, schema.Attributes[key])
				}
			}
			if _, ok := schema.Attributes["items"].(datasourceschema.ListNestedAttribute); !ok {
				t.Fatalf("items schema mismatch: %T", schema.Attributes["items"])
			}
		})
	}
}

func TestProjectAccessRoleModelsUseThreePartIdentity(t *testing.T) {
	ctx := context.Background()
	assignment := &client.RoleAssignment{Role: client.Role{ID: "role_1", Name: "Reader", Permissions: []string{"project.read"}}}
	user, diags := projectUserRoleResourceModelFromAPI(ctx, assignment, "proj_1", "user_1")
	if diags.HasError() || user.ID != types.StringValue("proj_1/user_1/role_1") || user.ProjectID != types.StringValue("proj_1") {
		t.Fatalf("user model = %#v, diags=%v", user, diags)
	}
	group, diags := projectGroupRoleResourceModelFromAPI(ctx, assignment, "proj_1", "group_1")
	if diags.HasError() || group.ID != types.StringValue("proj_1/group_1/role_1") || group.ProjectID != types.StringValue("proj_1") {
		t.Fatalf("group model = %#v, diags=%v", group, diags)
	}
}
