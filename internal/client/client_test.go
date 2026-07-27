package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
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
