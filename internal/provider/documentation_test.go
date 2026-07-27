package provider

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderRegisteredSurfacesMatchTerraformDocumentation(t *testing.T) {
	ctx := context.Background()
	p := &openAIProvider{version: "test"}

	resourceHierarchies := map[string]string{
		"openai_admin_api_key":               "Administration > Organization > Admin API Keys",
		"openai_organization_certificate":    "Administration > Organization > Certificates",
		"openai_organization_data_retention": "Administration > Organization > Data Retention",
		"openai_organization_group":          "Administration > Organization > Groups",
		"openai_organization_group_role":     "Administration > Organization > Groups > Roles",
		"openai_organization_group_user":     "Administration > Organization > Groups > Users",
		"openai_organization_invite":         "Administration > Organization > Invites",
		"openai_organization_role":           "Administration > Organization > Roles",
		"openai_organization_spend_alert":    "Administration > Organization > Spend Alerts",
		"openai_organization_spend_limit":    "Administration > Organization > Spend Limit",
		"openai_organization_user_role":      "Administration > Organization > Users > Roles",
		"openai_project":                     "Administration > Organization > Projects",
		"openai_project_api_key":             "Administration > Organization > Projects > Service Accounts > API Keys",
		"openai_project_role":                "Administration > Organization > Projects > Roles",
		"openai_service_account":             "Administration > Organization > Projects > Service Accounts",
	}
	dataSourceHierarchies := map[string]string{
		"openai_admin_api_key":               "Administration > Organization > Admin API Keys",
		"openai_admin_api_keys":              "Administration > Organization > Admin API Keys",
		"openai_organization_certificate":    "Administration > Organization > Certificates",
		"openai_organization_certificates":   "Administration > Organization > Certificates",
		"openai_organization_data_retention": "Administration > Organization > Data Retention",
		"openai_organization_group":          "Administration > Organization > Groups",
		"openai_organization_group_role":     "Administration > Organization > Groups > Roles",
		"openai_organization_group_roles":    "Administration > Organization > Groups > Roles",
		"openai_organization_group_user":     "Administration > Organization > Groups > Users",
		"openai_organization_group_users":    "Administration > Organization > Groups > Users",
		"openai_organization_groups":         "Administration > Organization > Groups",
		"openai_organization_invite":         "Administration > Organization > Invites",
		"openai_organization_invites":        "Administration > Organization > Invites",
		"openai_organization_role":           "Administration > Organization > Roles",
		"openai_organization_roles":          "Administration > Organization > Roles",
		"openai_organization_spend_alert":    "Administration > Organization > Spend Alerts",
		"openai_organization_spend_alerts":   "Administration > Organization > Spend Alerts",
		"openai_organization_spend_limit":    "Administration > Organization > Spend Limit",
		"openai_organization_user":           "Administration > Organization > Users",
		"openai_organization_user_role":      "Administration > Organization > Users > Roles",
		"openai_organization_user_roles":     "Administration > Organization > Users > Roles",
		"openai_organization_users":          "Administration > Organization > Users",
		"openai_project":                     "Administration > Organization > Projects",
		"openai_project_role":                "Administration > Organization > Projects > Roles",
		"openai_project_roles":               "Administration > Organization > Projects > Roles",
		"openai_projects":                    "Administration > Organization > Projects",
		"openai_service_account":             "Administration > Organization > Projects > Service Accounts",
		"openai_service_accounts":            "Administration > Organization > Projects > Service Accounts",
	}

	registeredResources := resourceTypeNames(ctx, p.Resources(ctx))
	assertStringSet(t, "resources", registeredResources, sortedKeys(resourceHierarchies))
	assertSurfaceDocs(t, "resources", registeredResources, resourceHierarchies)

	registeredDataSources := dataSourceTypeNames(ctx, p.DataSources(ctx))
	assertStringSet(t, "data sources", registeredDataSources, sortedKeys(dataSourceHierarchies))
	assertSurfaceDocs(t, "data-sources", registeredDataSources, dataSourceHierarchies)
}

func TestTerraformDocumentationMirrorsOpenAIAdminAPIHierarchy(t *testing.T) {
	repoRoot := repositoryRootFromProviderPackage()
	expectations := map[string][]string{
		"docs/api-groups/README.md": {
			"# OpenAI Administration API hierarchy",
			"Administration > Organization > Invites",
			"Administration > Organization > Projects > Service Accounts > API Keys",
			"Administration > Organization > Projects > Model Permissions",
		},
		"README.yaml": {
			"OpenAI API hierarchy represented by this provider:",
			"Administration > Organization > Groups > Roles",
			"Administration > Organization > Projects > Service Accounts > API Keys",
			"Administration > Organization > Projects > Model Permissions",
		},
		"docs/index.md": {
			"OpenAI API hierarchy represented by this provider:",
			"Administration > Organization > Data Retention",
			"Administration > Organization > Projects > Service Accounts > API Keys",
			"Administration > Organization > Projects > Hosted Tool Permissions",
		},
	}

	for relativePath, expected := range expectations {
		t.Run(relativePath, func(t *testing.T) {
			content := readRepoFile(t, repoRoot, relativePath)
			for _, want := range expected {
				if !strings.Contains(content, want) {
					t.Fatalf("%s is missing %q", relativePath, want)
				}
			}
		})
	}
}

func resourceTypeNames(ctx context.Context, factories []func() resource.Resource) []string {
	names := make([]string, 0, len(factories))
	for _, factory := range factories {
		instance := factory()
		var resp resource.MetadataResponse
		instance.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "openai"}, &resp)
		names = append(names, resp.TypeName)
	}
	sort.Strings(names)
	return names
}

func dataSourceTypeNames(ctx context.Context, factories []func() datasource.DataSource) []string {
	names := make([]string, 0, len(factories))
	for _, factory := range factories {
		instance := factory()
		var resp datasource.MetadataResponse
		instance.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "openai"}, &resp)
		names = append(names, resp.TypeName)
	}
	sort.Strings(names)
	return names
}

func assertSurfaceDocs(t *testing.T, docsDir string, typeNames []string, hierarchies map[string]string) {
	t.Helper()
	repoRoot := repositoryRootFromProviderPackage()
	for _, typeName := range typeNames {
		t.Run(docsDir+"/"+typeName, func(t *testing.T) {
			content := readRepoFile(t, repoRoot, filepath.Join("docs", docsDir, typeName+".md"))
			if !strings.HasPrefix(content, "# "+typeName+"\n") {
				t.Fatalf("documentation title mismatch for %s", typeName)
			}
			if !strings.Contains(content, "API group:") {
				t.Fatalf("documentation for %s is missing API group", typeName)
			}
			expectedHierarchy := "OpenAI API hierarchy: `" + hierarchies[typeName] + "`."
			if !strings.Contains(content, expectedHierarchy) {
				t.Fatalf("documentation for %s is missing hierarchy line %q", typeName, expectedHierarchy)
			}
		})
	}
}

func assertStringSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") == strings.Join(want, "\n") {
		return
	}
	gotSet := make(map[string]bool, len(got))
	wantSet := make(map[string]bool, len(want))
	for _, value := range got {
		gotSet[value] = true
	}
	for _, value := range want {
		wantSet[value] = true
	}
	var missing, extra []string
	for _, value := range want {
		if !gotSet[value] {
			missing = append(missing, value)
		}
	}
	for _, value := range got {
		if !wantSet[value] {
			extra = append(extra, value)
		}
	}
	t.Fatalf("registered %s mismatch\nmissing: %v\nextra: %v\ngot: %v\nwant: %v", label, missing, extra, got, want)
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func repositoryRootFromProviderPackage() string {
	return filepath.Clean(filepath.Join("..", ".."))
}

func readRepoFile(t *testing.T, repoRoot, relativePath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(content)
}
