package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func serviceAccountErrorResponse(status int) *http.Response {
	response := jsonResponse(map[string]any{"error": map[string]any{"message": "account not ready", "type": "invalid_request_error"}})
	response.StatusCode = status
	response.Status = fmt.Sprintf("%d %s", status, http.StatusText(status))
	return response
}

func TestServiceAccountCreateRetriesOnlyRoleAssignment(t *testing.T) {
	for _, createOnly := range []bool{true, false} {
		t.Run(fmt.Sprintf("createOnly=%t", createOnly), func(t *testing.T) {
			creates, updates := 0, 0
			cl := testClient(func(r *http.Request) (*http.Response, error) {
				switch r.Method + " " + r.URL.Path {
				case "POST /organization/projects/proj_1/service_accounts":
					creates++
					return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "none", "created_at": 11, "api_key": map[string]any{"id": "key_1", "name": "key", "value": "sk-bootstrap", "created_at": 11}}), nil
				case "POST /organization/projects/proj_1/service_accounts/sa_1":
					updates++
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if len(body) != 1 || body["role"] != "member" {
						t.Fatalf("not a role-only assignment: %#v", body)
					}
					deadline, ok := r.Context().Deadline()
					if !ok || time.Until(deadline) > 30*time.Second {
						t.Fatal("role assignment must have a bounded deadline")
					}
					if updates < 3 {
						return serviceAccountErrorResponse(http.StatusNotFound), nil
					}
					return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "member", "created_at": 11}), nil
				default:
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				return nil, nil
			})
			account, err := cl.CreateServiceAccount(context.Background(), "proj_1", ServiceAccountCreateRequest{Name: "svc", Role: "member", CreateServiceAccountOnly: createOnly})
			if err != nil || account == nil || account.ID != "sa_1" || account.Role != "member" {
				t.Fatalf("create = %#v, %v", account, err)
			}
			if creates != 1 || updates != 3 {
				t.Fatalf("creates=%d, updates=%d", creates, updates)
			}
			if createOnly && account.APIKey != nil {
				t.Fatal("unexpected bootstrap key")
			}
			if !createOnly && (account.APIKey == nil || account.APIKey.Value != "sk-bootstrap") {
				t.Fatal("bootstrap key was not preserved")
			}
		})
	}
}

func TestServiceAccountCreateRoleFailurePreservesCreatedAccount(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			creates, updates := 0, 0
			cl := testClient(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/organization/projects/proj_1/service_accounts" {
					creates++
					return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "none", "created_at": 11, "api_key": map[string]any{"id": "key_1", "name": "key", "value": "sk-bootstrap", "created_at": 11}}), nil
				}
				updates++
				return serviceAccountErrorResponse(status), nil
			})
			account, err := cl.CreateServiceAccount(context.Background(), "proj_1", ServiceAccountCreateRequest{Name: "svc", Role: "member"})
			if err == nil || account == nil || account.ID != "sa_1" || account.Role != "none" || account.APIKey == nil || account.APIKey.Value != "sk-bootstrap" {
				t.Fatalf("create = %#v, %v", account, err)
			}
			if creates != 1 || updates != 1 {
				t.Fatalf("creates=%d, updates=%d; non-404 must not retry", creates, updates)
			}
		})
	}
}

func TestServiceAccountCreateRoleRetryHonorsContext(t *testing.T) {
	for _, canceled := range []bool{true, false} {
		t.Run(fmt.Sprintf("canceled=%t", canceled), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			expected := context.Canceled
			if !canceled {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
				expected = context.DeadlineExceeded
			}
			defer cancel()
			creates, updates := 0, 0
			cl := testClient(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/organization/projects/proj_1/service_accounts" {
					creates++
					return jsonResponse(map[string]any{"id": "sa_1", "name": "svc", "role": "none", "created_at": 11}), nil
				}
				updates++
				if canceled {
					cancel()
				}
				return serviceAccountErrorResponse(http.StatusNotFound), nil
			})
			started := time.Now()
			account, err := cl.CreateServiceAccount(ctx, "proj_1", ServiceAccountCreateRequest{Name: "svc", Role: "member", CreateServiceAccountOnly: true})
			if !errors.Is(err, expected) || account == nil || account.ID != "sa_1" {
				t.Fatalf("create = %#v, %v; want %v with partial account", account, err, expected)
			}
			if creates != 1 || updates != 1 {
				t.Fatalf("creates=%d, updates=%d", creates, updates)
			}
			if elapsed := time.Since(started); elapsed > time.Second {
				t.Fatalf("context interruption took %v", elapsed)
			}
		})
	}
}

func TestServiceAccountCreateDoesNotRetryNonIdempotentRequest(t *testing.T) {
	for _, transportFailure := range []bool{false, true} {
		t.Run(fmt.Sprintf("transportFailure=%t", transportFailure), func(t *testing.T) {
			creates := 0
			cl := testClient(func(r *http.Request) (*http.Response, error) {
				creates++
				if transportFailure {
					return nil, errors.New("connection lost")
				}
				return serviceAccountErrorResponse(http.StatusInternalServerError), nil
			})
			account, err := cl.CreateServiceAccount(context.Background(), "proj_1", ServiceAccountCreateRequest{Name: "svc", Role: "member", CreateServiceAccountOnly: true})
			if err == nil || account != nil || creates != 1 {
				t.Fatalf("create = %#v, %v; requests=%d", account, err, creates)
			}
		})
	}
}

func TestServiceAccountStandaloneUpdateDoesNotRetryNotFound(t *testing.T) {
	calls := 0
	cl := testClient(func(r *http.Request) (*http.Response, error) {
		calls++
		return serviceAccountErrorResponse(http.StatusNotFound), nil
	})
	account, err := cl.UpdateServiceAccount(context.Background(), "proj_1", "sa_1", ServiceAccountUpdateRequest{Role: "member"})
	if account != nil || !IsNotFound(err) || calls != 1 {
		t.Fatalf("update = %#v, %v; requests=%d", account, err, calls)
	}
}
