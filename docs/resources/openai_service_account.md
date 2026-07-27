# openai_service_account

API group: [Projects](../api-groups/projects.md).

Manages an OpenAI project service account through the Administration API.
Creation uses `create_service_account_only=true` so that OpenAI does not issue an
implicit unmanaged API key. Use `openai_project_api_key` to create explicit,
Terraform-managed keys for the service account.

## Example Usage

```terraform
resource "openai_project" "app" {
  name = "example-app"
}

resource "openai_service_account" "app" {
  project_id = openai_project.app.id
  name       = "example-app"
  role       = "member"
}
```

## Import

```sh
terraform import openai_service_account.app proj_123/svc_acct_123
```

## Schema

### Required

- `project_id` (String) OpenAI project ID that owns the service account. Changes replace the service account.
- `name` (String) Service account name.

### Optional

- `role` (String) Project role for the service account. Defaults to `member`. Supported values are `member` and `owner`.

### Read-Only

- `id` (String) OpenAI service account ID.
- `created_at` (Number) Unix timestamp for service-account creation.

## Notes

- Changing `project_id` replaces the service account.
- Deleting the resource deletes the service account in OpenAI.
- Keep service accounts narrowly scoped and issue keys through `openai_project_api_key` so key creation is explicit and auditable.
