package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestProjectPolicyResourceSchemaContracts(t *testing.T) {
	ctx := context.Background()
	resources := map[string]struct {
		resource resource.Resource
		required []string
		computed []string
	}{
		"data retention":          {NewProjectDataRetentionResource(), []string{"project_id", "type"}, []string{"id"}},
		"spend limit":             {NewProjectSpendLimitResource(), []string{"project_id", "threshold_amount"}, []string{"id", "enforcement_status"}},
		"model permissions":       {NewProjectModelPermissionsResource(), []string{"project_id", "mode", "model_ids"}, []string{"id"}},
		"hosted tool permissions": {NewProjectHostedToolPermissionsResource(), []string{"project_id", "code_interpreter", "file_search", "image_generation", "mcp", "web_search"}, []string{"id"}},
	}
	for name, tc := range resources {
		t.Run(name, func(t *testing.T) {
			schema := resourceSchema(ctx, t, tc.resource)
			for _, attribute := range tc.required {
				assertResourceAttributeFlags(t, schema.Attributes, attribute, true, false)
			}
			for _, attribute := range tc.computed {
				assertResourceAttributeFlags(t, schema.Attributes, attribute, false, true)
			}
			assertProjectIDRequiresReplacement(t, schema.Attributes)
		})
	}
}

func TestProjectPolicyDataSourceSchemaContracts(t *testing.T) {
	ctx := context.Background()
	dataSources := map[string]struct {
		dataSource datasource.DataSource
		computed   []string
	}{
		"data retention":          {NewProjectDataRetentionDataSource(), []string{"id", "type"}},
		"spend limit":             {NewProjectSpendLimitDataSource(), []string{"id", "threshold_amount", "currency", "interval", "enforcement_status"}},
		"model permissions":       {NewProjectModelPermissionsDataSource(), []string{"id", "mode", "model_ids"}},
		"hosted tool permissions": {NewProjectHostedToolPermissionsDataSource(), []string{"id", "code_interpreter", "file_search", "image_generation", "mcp", "web_search"}},
	}
	for name, tc := range dataSources {
		t.Run(name, func(t *testing.T) {
			schema := dataSourceSchema(ctx, t, tc.dataSource)
			assertDataSourceAttributeFlags(t, schema.Attributes, "project_id", true, false)
			for _, attribute := range tc.computed {
				assertDataSourceAttributeFlags(t, schema.Attributes, attribute, false, true)
			}
		})
	}
}

func assertResourceAttributeFlags(t *testing.T, attributes map[string]resourceschema.Attribute, name string, required, computed bool) {
	t.Helper()
	attribute, ok := attributes[name]
	if !ok {
		t.Fatalf("missing %s", name)
	}
	switch value := attribute.(type) {
	case resourceschema.StringAttribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	case resourceschema.Int64Attribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	case resourceschema.BoolAttribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	case resourceschema.SetAttribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	default:
		t.Fatalf("unexpected %s type: %T", name, attribute)
	}
}

func assertProjectIDRequiresReplacement(t *testing.T, attributes map[string]resourceschema.Attribute) {
	t.Helper()
	projectID, ok := attributes["project_id"].(resourceschema.StringAttribute)
	if !ok || len(projectID.PlanModifiers) != 1 {
		t.Fatalf("project_id must require replacement: %#v", attributes["project_id"])
	}
}

func assertDataSourceAttributeFlags(t *testing.T, attributes map[string]datasourceschema.Attribute, name string, required, computed bool) {
	t.Helper()
	attribute, ok := attributes[name]
	if !ok {
		t.Fatalf("missing %s", name)
	}
	switch value := attribute.(type) {
	case datasourceschema.StringAttribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	case datasourceschema.Int64Attribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	case datasourceschema.BoolAttribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	case datasourceschema.SetAttribute:
		if value.Required != required || value.Computed != computed {
			t.Fatalf("%s flags: %#v", name, value)
		}
	default:
		t.Fatalf("unexpected %s type: %T", name, attribute)
	}
}
