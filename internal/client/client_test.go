package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAIAdminClientProjectLifecycle(t *testing.T) {
	var requests []string
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if got := r.Header.Get("Authorization"); got != "Bearer admin-key" {
			t.Fatalf("Authorization header = %q", got)
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /organization/projects":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "App" || body["external_key_id"] != "ext" || body["geography"] != "eu" {
				t.Fatalf("unexpected create body: %#v", body)
			}
			return jsonResponse(map[string]any{"id": "proj_1", "name": "App", "status": "active", "external_key_id": "ext", "created_at": 10}), nil
		case "GET /organization/projects/proj_1":
			return jsonResponse(map[string]any{"id": "proj_1", "name": "App", "status": "active", "created_at": 10}), nil
		case "GET /organization/projects":
			if got := r.URL.Query().Get("after"); got != "proj_0" {
				t.Fatalf("after query = %q", got)
			}
			if got := r.URL.Query().Get("limit"); got != "2" {
				t.Fatalf("limit query = %q", got)
			}
			if got := r.URL.Query().Get("include_archived"); got != "true" {
				t.Fatalf("include_archived query = %q", got)
			}
			return jsonResponse(map[string]any{
				"data": []map[string]any{
					{"id": "proj_1", "name": "App", "status": "active", "created_at": 10},
					{"id": "proj_2", "name": "Archived", "status": "archived", "created_at": 11, "archived_at": 20},
				},
				"has_more": true,
				"last_id":  "proj_2",
			}), nil
		case "POST /organization/projects/proj_1":
			return jsonResponse(map[string]any{"id": "proj_1", "name": "App 2", "status": "active", "created_at": 10}), nil
		case "POST /organization/projects/proj_1/archive":
			return jsonResponse(map[string]any{"id": "proj_1", "name": "App 2", "status": "archived", "created_at": 10, "archived_at": 20}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	created, err := cl.CreateProject(context.Background(), ProjectCreateRequest{Name: "App", ExternalKeyID: "ext", Geography: "eu"})
	if err != nil || created.ID != "proj_1" || created.Name != "App" {
		t.Fatalf("CreateProject() = %#v, %v", created, err)
	}
	if got, err := cl.GetProject(context.Background(), "proj_1"); err != nil || got.Status != "active" {
		t.Fatalf("GetProject() = %#v, %v", got, err)
	}
	list, err := cl.ListProjects(context.Background(), ProjectListRequest{After: "proj_0", Limit: 2, IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(list.Items) != 2 || list.Items[1].ID != "proj_2" || !list.HasMore || list.LastID != "proj_2" {
		t.Fatalf("ListProjects() = %#v", list)
	}
	if got, err := cl.UpdateProject(context.Background(), "proj_1", ProjectUpdateRequest{Name: "App 2"}); err != nil || got.Name != "App 2" {
		t.Fatalf("UpdateProject() = %#v, %v", got, err)
	}
	if got, err := cl.ArchiveProject(context.Background(), "proj_1"); err != nil || got.Status != "archived" || got.ArchivedAt != 20 {
		t.Fatalf("ArchiveProject() = %#v, %v", got, err)
	}
	if len(requests) != 5 {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestOpenAIAdminClientServiceAccountLifecycle(t *testing.T) {
	updateCalls := 0
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		switch r.Method + " " + r.URL.Path {
		case "POST /organization/projects/proj_1/service_accounts":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "svc" || body["create_service_account_only"] != true {
				t.Fatalf("unexpected create body: %#v", body)
			}
			return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "none", "created_at": 11}), nil
		case "POST /organization/projects/proj_1/service_accounts/sa_1":
			updateCalls++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch updateCalls {
			case 1:
				if body["role"] != "member" {
					t.Fatalf("expected member role update: %#v", body)
				}
				return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11}), nil
			case 2:
				if body["name"] != "svc2" || body["role"] != "owner" {
					t.Fatalf("expected name and owner role update: %#v", body)
				}
				return jsonResponse(map[string]any{"id": "sa_1", "name": "svc2", "role": "owner", "created_at": 11}), nil
			default:
				t.Fatalf("unexpected service account update call %d: %#v", updateCalls, body)
			}
		case "GET /organization/projects/proj_1/service_accounts/sa_1":
			return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11}), nil
		case "GET /organization/projects/proj_1/service_accounts":
			if got := r.URL.Query().Get("after"); got != "sa_0" {
				t.Fatalf("after query = %q", got)
			}
			if got := r.URL.Query().Get("limit"); got != "2" {
				t.Fatalf("limit query = %q", got)
			}
			body := map[string]any{
				"data": []map[string]any{
					{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11},
					{"id": "sa_2", "name": "batch", "role": "owner", "created_at": 12},
				},
				"has_more": true,
				"last_id":  "sa_2",
			}
			return jsonResponse(body), nil
		case "DELETE /organization/projects/proj_1/service_accounts/sa_1":
			return jsonResponse(map[string]any{"id": "sa_1", "deleted": true}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	created, err := cl.CreateServiceAccount(context.Background(), "proj_1", ServiceAccountCreateRequest{Name: "svc", Role: "member"})
	if err != nil || created.ID != "sa_1" || created.Role != "member" {
		t.Fatalf("CreateServiceAccount() = %#v, %v", created, err)
	}
	if got, err := cl.GetServiceAccount(context.Background(), "proj_1", "sa_1"); err != nil || got.Name != "svc" {
		t.Fatalf("GetServiceAccount() = %#v, %v", got, err)
	}
	list, err := cl.ListServiceAccounts(context.Background(), "proj_1", ServiceAccountListRequest{After: "sa_0", Limit: 2})
	if err != nil {
		t.Fatalf("ListServiceAccounts() error = %v", err)
	}
	if len(list.Items) != 2 || list.Items[1].ID != "sa_2" || !list.HasMore || list.LastID != "sa_2" {
		t.Fatalf("ListServiceAccounts() = %#v", list)
	}
	updated, err := cl.UpdateServiceAccount(context.Background(), "proj_1", "sa_1", ServiceAccountUpdateRequest{Name: "svc2", Role: "owner"})
	if err != nil || updated.Name != "svc2" || updated.Role != "owner" {
		t.Fatalf("UpdateServiceAccount() = %#v, %v", updated, err)
	}
	if err := cl.DeleteServiceAccount(context.Background(), "proj_1", "sa_1"); err != nil {
		t.Fatalf("DeleteServiceAccount() = %v", err)
	}
}

func TestOpenAIAdminClientAPIKeyLifecycle(t *testing.T) {
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		switch r.Method + " " + r.URL.Path {
		case "POST /organization/projects/proj_1/service_accounts/sa_1/api_keys":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "svc-key" {
				t.Fatalf("unexpected key create body: %#v", body)
			}
			return jsonResponse(map[string]any{"id": "key_1", "name": "svc-key", "value": "sk-created", "created_at": 12}), nil
		case "GET /organization/projects/proj_1/api_keys/key_1":
			return jsonResponse(map[string]any{
				"id": "key_1", "name": "svc-key", "redacted_value": "sk-...", "created_at": 12, "last_used_at": 0,
				"owner_project_access": "active",
				"owner":                map[string]any{"type": "service_account", "service_account": map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11}},
			}), nil
		case "DELETE /organization/projects/proj_1/api_keys/key_1":
			return jsonResponse(map[string]any{"id": "key_1", "deleted": true}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	created, err := cl.CreateServiceAccountAPIKey(context.Background(), "proj_1", "sa_1", ServiceAccountAPIKeyCreateRequest{Name: "svc-key", Scopes: []string{"models.read"}})
	if err != nil || created.Value != "sk-created" {
		t.Fatalf("CreateServiceAccountAPIKey() = %#v, %v", created, err)
	}
	read, err := cl.GetProjectAPIKey(context.Background(), "proj_1", "key_1")
	if err != nil || read.OwnerID != "sa_1" || read.RedactedValue != "sk-..." {
		t.Fatalf("GetProjectAPIKey() = %#v, %v", read, err)
	}
	if err := cl.DeleteProjectAPIKey(context.Background(), "proj_1", "key_1"); err != nil {
		t.Fatalf("DeleteProjectAPIKey() = %v", err)
	}
}

func TestOpenAIAdminClientSettingsHeaders(t *testing.T) {
	cl := newWithHTTPClient(Settings{
		AdminAPIKey:    "admin-key",
		BaseURL:        "https://api.test/",
		OrganizationID: "org_1",
		ProjectID:      "proj_1",
	}, "test-agent", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer admin-key" {
			t.Fatalf("Authorization header = %q", got)
		}
		if got := r.Header.Get("OpenAI-Organization"); got != "org_1" {
			t.Fatalf("OpenAI-Organization header = %q", got)
		}
		if got := r.Header.Get("OpenAI-Project"); got != "proj_1" {
			t.Fatalf("OpenAI-Project header = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Fatalf("User-Agent header = %q", got)
		}
		if got := r.URL.String(); got != "https://api.test/organization/projects" {
			t.Fatalf("request URL = %q", got)
		}
		return jsonResponse(map[string]any{"data": []map[string]any{}, "has_more": false}), nil
	})})

	if _, err := cl.ListProjects(context.Background(), ProjectListRequest{}); err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
}

func TestTerraformLogEnablesOpenAIDebugLog(t *testing.T) {
	tests := map[string]bool{
		"TRACE":   true,
		"trace":   true,
		" trace ": true,
		"JSON":    true,
		"DEBUG":   false,
		"INFO":    false,
		"WARN":    false,
		"ERROR":   false,
		"OFF":     false,
		"":        false,
	}
	for value, want := range tests {
		if got := terraformLogEnablesOpenAIDebugLog(value); got != want {
			t.Fatalf("terraformLogEnablesOpenAIDebugLog(%q) = %t, want %t", value, got, want)
		}
	}
}

func TestOpenAIAdminClientDebugLogEnabledByTFLogTrace(t *testing.T) {
	t.Setenv("TF_LOG", "TRACE")
	var logBuffer bytes.Buffer
	restoreDefaultLogger(t, &logBuffer)

	cl := testClient(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer admin-key" {
			t.Fatalf("Authorization header = %q", got)
		}
		return jsonResponse(map[string]any{"data": []map[string]any{}, "has_more": false}), nil
	})
	if _, err := cl.ListProjects(context.Background(), ProjectListRequest{}); err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	logs := logBuffer.String()
	if !strings.Contains(logs, "Request Content:") || !strings.Contains(logs, "Response Content:") {
		t.Fatalf("expected OpenAI debug request/response logs, got %q", logs)
	}
	if !strings.Contains(logs, "GET /organization/projects") {
		t.Fatalf("expected request path in debug logs, got %q", logs)
	}
	if !strings.Contains(logs, "Authorization: ***") {
		t.Fatalf("expected redacted authorization header in debug logs, got %q", logs)
	}
	if strings.Contains(logs, "admin-key") {
		t.Fatalf("debug logs leaked admin key: %q", logs)
	}
}

func TestOpenAIAdminClientDebugLogDisabledBelowTrace(t *testing.T) {
	t.Setenv("TF_LOG", "DEBUG")
	var logBuffer bytes.Buffer
	restoreDefaultLogger(t, &logBuffer)

	cl := testClient(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(map[string]any{"data": []map[string]any{}, "has_more": false}), nil
	})
	if _, err := cl.ListProjects(context.Background(), ProjectListRequest{}); err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if logs := logBuffer.String(); logs != "" {
		t.Fatalf("expected no OpenAI debug logs for TF_LOG=DEBUG, got %q", logs)
	}
}

func restoreDefaultLogger(t *testing.T, output io.Writer) {
	t.Helper()
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	previousPrefix := log.Prefix()
	log.SetOutput(output)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
	})
}

func TestValidationErrors(t *testing.T) {
	if _, err := mapProject(nil); err == nil {
		t.Fatal("expected nil project error")
	}
	if err := validateServiceAccount(&ServiceAccount{}); err == nil {
		t.Fatal("expected invalid service account")
	}
	if err := validateServiceAccountAPIKeyCreate(&ServiceAccountAPIKeyCreateResponse{ID: "key", Name: "n"}); err == nil {
		t.Fatal("expected missing value")
	}
	if err := validateAPIKey(&APIKey{Name: "n"}); err == nil {
		t.Fatal("expected missing id")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testClient(fn roundTripFunc) *OpenAIAdminClient {
	return newWithHTTPClient(Settings{AdminAPIKey: "admin-key", BaseURL: "https://api.test"}, "test-agent", &http.Client{Transport: fn})
}

func jsonResponse(value any) *http.Response {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		panic(err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(&body),
	}
}

func TestErrorHelpersWithNonAPIError(t *testing.T) {
	err := assertErr("boom")
	if IsNotFound(err) {
		t.Fatal("non-API error should not be not-found")
	}
	if got := ErrorSummary(err); got != "boom" {
		t.Fatalf("ErrorSummary() = %q", got)
	}
	if got := ErrorSummary(nil); got != "" {
		t.Fatalf("nil ErrorSummary() = %q", got)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

func TestOpenAIAdminClientAdminAPIKeyLifecycle(t *testing.T) {
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		switch r.Method + " " + r.URL.Path {
		case "POST /organization/admin_api_keys":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "admin-key" || body["expires_in_seconds"] != float64(3600) {
				t.Fatalf("unexpected admin key create body: %#v", body)
			}
			scopes, ok := body["scopes"].([]any)
			if !ok || len(scopes) != 2 || scopes[0] != "organization.users.read" || scopes[1] != "organization.projects.read" {
				t.Fatalf("unexpected admin key scopes: %#v", body["scopes"])
			}
			return jsonResponse(adminAPIKeyJSON("admin_key_1", "admin-key", "sk-admin-created")), nil
		case "GET /organization/admin_api_keys/admin_key_1":
			return jsonResponse(adminAPIKeyJSON("admin_key_1", "admin-key", "")), nil
		case "GET /organization/admin_api_keys":
			if got := r.URL.Query().Get("after"); got != "admin_key_0" {
				t.Fatalf("after query = %q", got)
			}
			if got := r.URL.Query().Get("limit"); got != "2" {
				t.Fatalf("limit query = %q", got)
			}
			if got := r.URL.Query().Get("order"); got != "desc" {
				t.Fatalf("order query = %q", got)
			}
			return jsonResponse(map[string]any{"data": []map[string]any{adminAPIKeyJSON("admin_key_1", "admin-key", "")}, "has_more": false}), nil
		case "DELETE /organization/admin_api_keys/admin_key_1":
			return jsonResponse(map[string]any{"id": "admin_key_1", "deleted": true}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	created, err := cl.CreateAdminAPIKey(context.Background(), AdminAPIKeyCreateRequest{Name: "admin-key", ExpiresInSeconds: 3600, Scopes: []string{"organization.users.read", "organization.projects.read"}})
	if err != nil || created.Value != "sk-admin-created" || created.OwnerID != "user_1" {
		t.Fatalf("CreateAdminAPIKey() = %#v, %v", created, err)
	}
	got, err := cl.GetAdminAPIKey(context.Background(), "admin_key_1")
	if err != nil || got.RedactedValue != "sk-admin..." {
		t.Fatalf("GetAdminAPIKey() = %#v, %v", got, err)
	}
	list, err := cl.ListAdminAPIKeys(context.Background(), AdminAPIKeyListRequest{After: "admin_key_0", Limit: 2, Order: "desc"})
	if err != nil || len(list.Items) != 1 || list.LastID != "admin_key_1" {
		t.Fatalf("ListAdminAPIKeys() = %#v, %v", list, err)
	}
	if err := cl.DeleteAdminAPIKey(context.Background(), "admin_key_1"); err != nil {
		t.Fatalf("DeleteAdminAPIKey() = %v", err)
	}
}

func TestOpenAIAdminClientAdminAPIKeyDeleteRequiresRevocationConfirmation(t *testing.T) {
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		switch r.Method + " " + r.URL.Path {
		case "DELETE /organization/admin_api_keys/admin_key_1":
			return jsonResponse(map[string]any{"id": "admin_key_1", "deleted": false}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	err := cl.DeleteAdminAPIKey(context.Background(), "admin_key_1")
	if err == nil || !strings.Contains(err.Error(), "deleted=false") {
		t.Fatalf("DeleteAdminAPIKey() error = %v, want deleted=false confirmation error", err)
	}
}

func TestOpenAIAdminClientAdminAPIKeyDeleteRejectsIDMismatch(t *testing.T) {
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		switch r.Method + " " + r.URL.Path {
		case "DELETE /organization/admin_api_keys/admin_key_1":
			return jsonResponse(map[string]any{"id": "admin_key_2", "deleted": true}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	err := cl.DeleteAdminAPIKey(context.Background(), "admin_key_1")
	if err == nil || !strings.Contains(err.Error(), "id mismatch") {
		t.Fatalf("DeleteAdminAPIKey() error = %v, want id mismatch error", err)
	}
}

func TestOpenAIAdminClientOrganizationIdentitySurfaces(t *testing.T) {
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		switch r.Method + " " + r.URL.Path {
		case "GET /organization/users/user_1":
			return jsonResponse(organizationUserJSON("user_1")), nil
		case "GET /organization/users":
			if got := r.URL.Query()["emails[]"]; len(got) != 1 || got[0] != "user@example.com" {
				t.Fatalf("emails query = %#v raw=%s", r.URL.Query(), r.URL.RawQuery)
			}
			return jsonResponse(map[string]any{"data": []map[string]any{organizationUserJSON("user_1")}, "has_more": false, "last_id": "user_1"}), nil
		case "POST /organization/groups":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "Engineering" {
				t.Fatalf("unexpected group create body: %#v", body)
			}
			return jsonResponse(groupJSON("group_1", "Engineering")), nil
		case "GET /organization/groups/group_1":
			return jsonResponse(groupJSON("group_1", "Engineering")), nil
		case "GET /organization/groups":
			return jsonResponse(map[string]any{"data": []map[string]any{groupJSON("group_1", "Engineering")}, "has_more": false, "next": ""}), nil
		case "POST /organization/groups/group_1":
			return jsonResponse(map[string]any{"id": "group_1", "name": "Platform", "is_scim_managed": false, "created_at": 10}), nil
		case "DELETE /organization/groups/group_1":
			return jsonResponse(map[string]any{"id": "group_1", "deleted": true}), nil
		case "POST /organization/groups/group_1/users":
			return jsonResponse(map[string]any{"group_id": "group_1", "user_id": "user_1"}), nil
		case "GET /organization/groups/group_1/users/user_1":
			return jsonResponse(groupUserJSON("user_1")), nil
		case "GET /organization/groups/group_1/users":
			return jsonResponse(map[string]any{"data": []map[string]any{{"id": "user_1", "name": "User", "email": "user@example.com"}}, "has_more": false}), nil
		case "DELETE /organization/groups/group_1/users/user_1":
			return jsonResponse(map[string]any{"deleted": true}), nil
		case "POST /organization/roles":
			return jsonResponse(roleJSON("role_1", "OrgRole", "api.organization")), nil
		case "GET /organization/roles/role_1":
			return jsonResponse(roleJSON("role_1", "OrgRole", "api.organization")), nil
		case "GET /organization/roles":
			return jsonResponse(map[string]any{"data": []map[string]any{roleJSON("role_1", "OrgRole", "api.organization")}, "has_more": false}), nil
		case "POST /organization/roles/role_1":
			return jsonResponse(roleJSON("role_1", "OrgRoleUpdated", "api.organization")), nil
		case "DELETE /organization/roles/role_1":
			return jsonResponse(map[string]any{"id": "role_1", "deleted": true}), nil
		case "POST /projects/proj_1/roles":
			return jsonResponse(roleJSON("role_project_1", "ProjectRole", "api.project")), nil
		case "GET /projects/proj_1/roles/role_project_1":
			return jsonResponse(roleJSON("role_project_1", "ProjectRole", "api.project")), nil
		case "GET /projects/proj_1/roles":
			return jsonResponse(map[string]any{"data": []map[string]any{roleJSON("role_project_1", "ProjectRole", "api.project")}, "has_more": false}), nil
		case "POST /projects/proj_1/roles/role_project_1":
			return jsonResponse(roleJSON("role_project_1", "ProjectRoleUpdated", "api.project")), nil
		case "DELETE /projects/proj_1/roles/role_project_1":
			return jsonResponse(map[string]any{"id": "role_project_1", "deleted": true}), nil
		case "POST /organization/users/user_1/roles":
			return jsonResponse(map[string]any{"object": "user.role", "role": roleJSON("role_1", "OrgRole", "api.organization"), "user": organizationUserJSON("user_1")}), nil
		case "GET /organization/users/user_1/roles/role_1":
			return jsonResponse(roleAssignmentJSON("role_1", "OrgRole", "api.organization")), nil
		case "GET /organization/users/user_1/roles":
			return jsonResponse(map[string]any{"data": []map[string]any{roleAssignmentJSON("role_1", "OrgRole", "api.organization")}, "has_more": false}), nil
		case "DELETE /organization/users/user_1/roles/role_1":
			return jsonResponse(map[string]any{"deleted": true, "object": "user.role.deleted"}), nil
		case "POST /organization/groups/group_1/roles":
			return jsonResponse(map[string]any{"object": "group.role", "role": roleJSON("role_1", "OrgRole", "api.organization"), "group": groupJSON("group_1", "Engineering")}), nil
		case "GET /organization/groups/group_1/roles/role_1":
			return jsonResponse(roleAssignmentJSON("role_1", "OrgRole", "api.organization")), nil
		case "GET /organization/groups/group_1/roles":
			return jsonResponse(map[string]any{"data": []map[string]any{roleAssignmentJSON("role_1", "OrgRole", "api.organization")}, "has_more": false}), nil
		case "DELETE /organization/groups/group_1/roles/role_1":
			return jsonResponse(map[string]any{"deleted": true, "object": "group.role.deleted"}), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return nil, nil
	})

	if user, err := cl.GetOrganizationUser(context.Background(), "user_1"); err != nil || user.Email != "user@example.com" {
		t.Fatalf("GetOrganizationUser() = %#v, %v", user, err)
	}
	if users, err := cl.ListOrganizationUsers(context.Background(), OrganizationUserListRequest{Emails: []string{"user@example.com"}}); err != nil || len(users.Items) != 1 {
		t.Fatalf("ListOrganizationUsers() = %#v, %v", users, err)
	}
	if group, err := cl.CreateOrganizationGroup(context.Background(), OrganizationGroupCreateRequest{Name: "Engineering"}); err != nil || group.ID != "group_1" {
		t.Fatalf("CreateOrganizationGroup() = %#v, %v", group, err)
	}
	if group, err := cl.GetOrganizationGroup(context.Background(), "group_1"); err != nil || group.Name != "Engineering" {
		t.Fatalf("GetOrganizationGroup() = %#v, %v", group, err)
	}
	if groups, err := cl.ListOrganizationGroups(context.Background(), OrganizationGroupListRequest{}); err != nil || len(groups.Items) != 1 {
		t.Fatalf("ListOrganizationGroups() = %#v, %v", groups, err)
	}
	if group, err := cl.UpdateOrganizationGroup(context.Background(), "group_1", OrganizationGroupUpdateRequest{Name: "Platform"}); err != nil || group.Name != "Platform" {
		t.Fatalf("UpdateOrganizationGroup() = %#v, %v", group, err)
	}
	if err := cl.DeleteOrganizationGroup(context.Background(), "group_1"); err != nil {
		t.Fatalf("DeleteOrganizationGroup() = %v", err)
	}
	if member, err := cl.CreateOrganizationGroupUser(context.Background(), "group_1", OrganizationGroupUserCreateRequest{UserID: "user_1"}); err != nil || member.ID != "user_1" {
		t.Fatalf("CreateOrganizationGroupUser() = %#v, %v", member, err)
	}
	if member, err := cl.GetOrganizationGroupUser(context.Background(), "group_1", "user_1"); err != nil || member.UserType != "user" {
		t.Fatalf("GetOrganizationGroupUser() = %#v, %v", member, err)
	}
	if members, err := cl.ListOrganizationGroupUsers(context.Background(), "group_1", OrganizationGroupUserListRequest{}); err != nil || len(members.Items) != 1 {
		t.Fatalf("ListOrganizationGroupUsers() = %#v, %v", members, err)
	}
	if err := cl.DeleteOrganizationGroupUser(context.Background(), "group_1", "user_1"); err != nil {
		t.Fatalf("DeleteOrganizationGroupUser() = %v", err)
	}
	if role, err := cl.CreateOrganizationRole(context.Background(), RoleCreateRequest{Name: "OrgRole", Permissions: []string{"organization.users.read"}}); err != nil || role.ResourceType != "api.organization" {
		t.Fatalf("CreateOrganizationRole() = %#v, %v", role, err)
	}
	if role, err := cl.GetOrganizationRole(context.Background(), "role_1"); err != nil || role.Name != "OrgRole" {
		t.Fatalf("GetOrganizationRole() = %#v, %v", role, err)
	}
	if roles, err := cl.ListOrganizationRoles(context.Background(), RoleListRequest{}); err != nil || len(roles.Items) != 1 {
		t.Fatalf("ListOrganizationRoles() = %#v, %v", roles, err)
	}
	if role, err := cl.UpdateOrganizationRole(context.Background(), "role_1", RoleUpdateRequest{Name: "OrgRoleUpdated"}); err != nil || role.Name != "OrgRoleUpdated" {
		t.Fatalf("UpdateOrganizationRole() = %#v, %v", role, err)
	}
	if err := cl.DeleteOrganizationRole(context.Background(), "role_1"); err != nil {
		t.Fatalf("DeleteOrganizationRole() = %v", err)
	}
	if role, err := cl.CreateProjectRole(context.Background(), "proj_1", RoleCreateRequest{Name: "ProjectRole", Permissions: []string{"project.api_keys.read"}}); err != nil || role.ResourceType != "api.project" {
		t.Fatalf("CreateProjectRole() = %#v, %v", role, err)
	}
	if role, err := cl.GetProjectRole(context.Background(), "proj_1", "role_project_1"); err != nil || role.Name != "ProjectRole" {
		t.Fatalf("GetProjectRole() = %#v, %v", role, err)
	}
	if roles, err := cl.ListProjectRoles(context.Background(), "proj_1", RoleListRequest{}); err != nil || len(roles.Items) != 1 {
		t.Fatalf("ListProjectRoles() = %#v, %v", roles, err)
	}
	if role, err := cl.UpdateProjectRole(context.Background(), "proj_1", "role_project_1", RoleUpdateRequest{Name: "ProjectRoleUpdated"}); err != nil || role.Name != "ProjectRoleUpdated" {
		t.Fatalf("UpdateProjectRole() = %#v, %v", role, err)
	}
	if err := cl.DeleteProjectRole(context.Background(), "proj_1", "role_project_1"); err != nil {
		t.Fatalf("DeleteProjectRole() = %v", err)
	}
	if assignment, err := cl.CreateOrganizationUserRole(context.Background(), "user_1", RoleAssignmentCreateRequest{RoleID: "role_1"}); err != nil || assignment.PrincipalType != "user" {
		t.Fatalf("CreateOrganizationUserRole() = %#v, %v", assignment, err)
	}
	if assignment, err := cl.GetOrganizationUserRole(context.Background(), "user_1", "role_1"); err != nil || assignment.CreatedBy != "user_1" {
		t.Fatalf("GetOrganizationUserRole() = %#v, %v", assignment, err)
	}
	if assignments, err := cl.ListOrganizationUserRoles(context.Background(), "user_1", RoleAssignmentListRequest{}); err != nil || len(assignments.Items) != 1 {
		t.Fatalf("ListOrganizationUserRoles() = %#v, %v", assignments, err)
	}
	if err := cl.DeleteOrganizationUserRole(context.Background(), "user_1", "role_1"); err != nil {
		t.Fatalf("DeleteOrganizationUserRole() = %v", err)
	}
	if assignment, err := cl.CreateOrganizationGroupRole(context.Background(), "group_1", RoleAssignmentCreateRequest{RoleID: "role_1"}); err != nil || assignment.PrincipalType != "group" {
		t.Fatalf("CreateOrganizationGroupRole() = %#v, %v", assignment, err)
	}
	if assignment, err := cl.GetOrganizationGroupRole(context.Background(), "group_1", "role_1"); err != nil || assignment.ID != "role_1" {
		t.Fatalf("GetOrganizationGroupRole() = %#v, %v", assignment, err)
	}
	if assignments, err := cl.ListOrganizationGroupRoles(context.Background(), "group_1", RoleAssignmentListRequest{}); err != nil || len(assignments.Items) != 1 {
		t.Fatalf("ListOrganizationGroupRoles() = %#v, %v", assignments, err)
	}
	if err := cl.DeleteOrganizationGroupRole(context.Background(), "group_1", "role_1"); err != nil {
		t.Fatalf("DeleteOrganizationGroupRole() = %v", err)
	}
}

func adminAPIKeyJSON(id, name, value string) map[string]any {
	body := map[string]any{"id": id, "name": name, "redacted_value": "sk-admin...", "created_at": 21, "expires_at": 0, "last_used_at": 0, "owner": map[string]any{"type": "user", "id": "user_1", "name": "Owner", "role": "owner"}}
	if value != "" {
		body["value"] = value
	}
	return body
}

func organizationUserJSON(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "User", "email": "user@example.com", "role": "reader", "added_at": 22,
		"user":     map[string]any{"id": id, "name": "User", "email": "user@example.com", "enabled": true},
		"projects": map[string]any{"data": []map[string]any{{"id": "proj_1", "name": "Project", "role": "member"}}},
	}
}

func groupJSON(id, name string) map[string]any {
	return map[string]any{"id": id, "name": name, "group_type": "group", "is_scim_managed": false, "created_at": 23}
}

func groupUserJSON(id string) map[string]any {
	return map[string]any{"id": id, "name": "User", "email": "user@example.com", "is_service_account": false, "picture": "", "user_type": "user"}
}

func roleJSON(id, name, resourceType string) map[string]any {
	permission := "organization.users.read"
	if resourceType == "api.project" {
		permission = "project.api_keys.read"
	}
	return map[string]any{"id": id, "name": name, "description": "role", "permissions": []string{permission}, "predefined_role": false, "resource_type": resourceType}
}

func roleAssignmentJSON(id, name, resourceType string) map[string]any {
	body := roleJSON(id, name, resourceType)
	body["assignment_sources"] = []map[string]any{{"principal_id": "group_1", "principal_type": "group"}}
	body["created_at"] = 24
	body["updated_at"] = 25
	body["created_by"] = "user_1"
	body["created_by_user_obj"] = map[string]any{}
	body["metadata"] = map[string]any{}
	return body
}
