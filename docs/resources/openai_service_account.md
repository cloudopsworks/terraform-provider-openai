# openai_service_account

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Service Accounts`.

Manages an OpenAI project service account through the Administration API. By
OpenAI API default, creating a service account also returns one unredacted API key
for that service account; the provider captures that bootstrap key in
`api_key_value`.

When `scopes` are configured, OpenAI does not accept them on the service-account
create endpoint. The provider therefore creates the service account without the
default key, updates the requested role, and then creates a scoped service-account
API key through the same API-key creation path used by `openai_project_api_key`.
Use `openai_project_api_key` for additional standalone keys or independent key
rotation.

## Example Usage

Default OpenAI bootstrap key:

```terraform
resource "openai_project" "app" {
  name = "example-app"
}

resource "openai_service_account" "app" {
  project_id = openai_project.app.id
  name       = "example-app"
  role       = "member"
}

output "service_account_api_key" {
  value     = openai_service_account.app.api_key_value
  sensitive = true
}
```

Scoped bootstrap key:

```terraform
resource "openai_service_account" "scoped" {
  project_id = openai_project.app.id
  name       = "example-scoped-app"
  role       = "member"
  scopes     = ["responses.read", "models.read"]
}
```

## Import

```sh
terraform import openai_service_account.app proj_123/svc_acct_123
```

Imported service accounts include service-account metadata only. Existing API key
values cannot be recovered from OpenAI and will remain unset unless Terraform
created the key.

## Schema

### Required

- `project_id` (String) OpenAI project ID that owns the service account. Changes replace the service account.
- `name` (String) Service account name.

### Optional

- `role` (String) Project role for the service account. Defaults to `member`. Supported values are `member` and `owner`.
- `scopes` (Set of String) Optional scopes for the bootstrap API key. Non-empty scopes use the two-step create path: service account without a default key, then scoped service-account API key creation. Scope changes replace the service account and bootstrap key.

### Read-Only

- `id` (String) OpenAI service account ID.
- `created_at` (Number) Unix timestamp for service-account creation.
- `api_key_id` (String) ID of the API key created during service-account creation.
- `api_key_name` (String) Name of the API key created during service-account creation.
- `api_key_value` (String, Sensitive) Unredacted key value returned only during create. Protect Terraform state accordingly.
- `api_key_redacted_value` (String) Redacted API key value returned by read operations when available.
- `api_key_created_at` (Number) Unix timestamp for API key creation.
- `api_key_last_used_at` (Number) Unix timestamp when the API key was last used, if known.
- `api_key_owner_project_access` (String) Whether the API key owner currently has effective access to the project.

## Notes

- Treat Terraform state and plan artifacts as secret material whenever they contain `api_key_value`.
- Without `scopes`, OpenAI creates the default service-account API key and the provider captures it.
- With non-empty `scopes`, the provider suppresses the default key and creates one scoped API key through OpenAI's service-account API-key endpoint.
- Use `openai_project_api_key` when you need additional standalone keys for the same service account.
- Deleting the resource deletes the captured bootstrap key first when known, then deletes the service account.
