package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIAdminClientProjectLifecycle(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			writeJSON(w, map[string]any{"id": "proj_1", "name": "App", "status": "active", "external_key_id": "ext", "created_at": 10})
		case "GET /organization/projects/proj_1":
			writeJSON(w, map[string]any{"id": "proj_1", "name": "App", "status": "active", "created_at": 10})
		case "POST /organization/projects/proj_1":
			writeJSON(w, map[string]any{"id": "proj_1", "name": "App 2", "status": "active", "created_at": 10})
		case "POST /organization/projects/proj_1/archive":
			writeJSON(w, map[string]any{"id": "proj_1", "name": "App 2", "status": "archived", "created_at": 10, "archived_at": 20})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cl := New("admin-key", server.URL, "test-agent", 0)
	created, err := cl.CreateProject(context.Background(), ProjectCreateRequest{Name: "App", ExternalKeyID: "ext", Geography: "eu"})
	if err != nil || created.ID != "proj_1" || created.Name != "App" {
		t.Fatalf("CreateProject() = %#v, %v", created, err)
	}
	if got, err := cl.GetProject(context.Background(), "proj_1"); err != nil || got.Status != "active" {
		t.Fatalf("GetProject() = %#v, %v", got, err)
	}
	if got, err := cl.UpdateProject(context.Background(), "proj_1", ProjectUpdateRequest{Name: "App 2"}); err != nil || got.Name != "App 2" {
		t.Fatalf("UpdateProject() = %#v, %v", got, err)
	}
	if got, err := cl.ArchiveProject(context.Background(), "proj_1"); err != nil || got.Status != "archived" || got.ArchivedAt != 20 {
		t.Fatalf("ArchiveProject() = %#v, %v", got, err)
	}
	if len(requests) != 4 {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestOpenAIAdminClientServiceAccountLifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST /organization/projects/proj_1/service_accounts":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "svc" || body["create_service_account_only"] != true {
				t.Fatalf("unexpected create body: %#v", body)
			}
			writeJSON(w, map[string]any{"id": "sa_1", "name": "svc", "role": "none", "created_at": 11})
		case "POST /organization/projects/proj_1/service_accounts/sa_1":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["role"] != "member" {
				t.Fatalf("expected member role update: %#v", body)
			}
			writeJSON(w, map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11})
		case "GET /organization/projects/proj_1/service_accounts/sa_1":
			writeJSON(w, map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11})
		case "DELETE /organization/projects/proj_1/service_accounts/sa_1":
			writeJSON(w, map[string]any{"id": "sa_1", "deleted": true})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cl := New("admin-key", server.URL, "test-agent", 0)
	created, err := cl.CreateServiceAccount(context.Background(), "proj_1", ServiceAccountCreateRequest{Name: "svc", Role: "member"})
	if err != nil || created.ID != "sa_1" || created.Role != "member" {
		t.Fatalf("CreateServiceAccount() = %#v, %v", created, err)
	}
	if got, err := cl.GetServiceAccount(context.Background(), "proj_1", "sa_1"); err != nil || got.Name != "svc" {
		t.Fatalf("GetServiceAccount() = %#v, %v", got, err)
	}
	if err := cl.DeleteServiceAccount(context.Background(), "proj_1", "sa_1"); err != nil {
		t.Fatalf("DeleteServiceAccount() = %v", err)
	}
}

func TestOpenAIAdminClientAPIKeyLifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST /organization/projects/proj_1/service_accounts/sa_1/api_keys":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["name"] != "svc-key" {
				t.Fatalf("unexpected key create body: %#v", body)
			}
			writeJSON(w, map[string]any{"id": "key_1", "name": "svc-key", "value": "sk-created", "created_at": 12})
		case "GET /organization/projects/proj_1/api_keys/key_1":
			writeJSON(w, map[string]any{
				"id": "key_1", "name": "svc-key", "redacted_value": "sk-...", "created_at": 12, "last_used_at": 0,
				"owner_project_access": "active",
				"owner":                map[string]any{"type": "service_account", "service_account": map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11}},
			})
		case "DELETE /organization/projects/proj_1/api_keys/key_1":
			writeJSON(w, map[string]any{"id": "key_1", "deleted": true})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cl := New("admin-key", server.URL, "test-agent", 0)
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

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		panic(err)
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
