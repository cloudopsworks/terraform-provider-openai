# openai_service_account

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Service Accounts`.

Looks up one OpenAI project service account through the Administration API.

Use `id` for direct lookup. When `id` is omitted, the provider performs an exact
name lookup by scanning the documented service account list endpoint for the
given `project_id`.

## Example Usage

```terraform
data "openai_service_account" "ops" {
  project_id = "proj_abc123"
  name       = "ops"
}

data "openai_service_account" "by_id" {
  project_id = "proj_abc123"
  id         = "svc_acct_abc123"
}
```

## Schema

### Required

- `project_id` (String) OpenAI project ID that owns the service account.

### Optional

- `id` (String) OpenAI service account ID. When set, lookup uses the service account retrieve endpoint.
- `name` (String) Exact service account name. Required when `id` is omitted.

### Read-Only

- `role` (String) Project role for the service account.
- `created_at` (Number) Unix timestamp for service-account creation.
