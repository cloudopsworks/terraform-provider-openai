# openai_project_api_key

API group: [Projects](../api-groups/projects.md).

Manages an OpenAI project API key owned by a project service account. The
unredacted `value` is returned only during create and is stored as Sensitive
Terraform state. Refresh and read operations return only metadata and the
redacted value.

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

resource "openai_project_api_key" "app" {
  project_id         = openai_project.app.id
  service_account_id = openai_service_account.app.id
  name               = "example-app"
  scopes             = ["responses.read"]
}

output "project_api_key" {
  value     = openai_project_api_key.app.value
  sensitive = true
}
```

## Import

```sh
terraform import openai_project_api_key.app proj_123/svc_acct_123/key_123
```

## Schema

### Required

- `project_id` (String) OpenAI project ID. Changes replace the key.
- `service_account_id` (String) OpenAI service account ID that owns the key. Changes replace the key.
- `name` (String) API key name. OpenAI does not expose update for service-account API keys, so changes replace the key.

### Optional

- `scopes` (Set of String) Optional API key scopes. Scopes are create-only and changing them replaces the key.

### Read-Only

- `id` (String) OpenAI project API key ID.
- `value` (String, Sensitive) Unredacted key value returned only during create.
- `redacted_value` (String) Redacted API key value returned by read operations.
- `owner_type` (String) API key owner type returned by OpenAI.
- `owner_project_access` (String) Whether the key owner currently has effective access to the project.
- `created_at` (Number) Unix timestamp for API key creation.
- `last_used_at` (Number) Unix timestamp when the key was last used, if known.

## Notes

- Treat Terraform state and plan artifacts as secret material whenever they contain this resource.
- OpenAI project API keys are immutable in this provider. Change `project_id`, `service_account_id`, `name`, or `scopes` by replacing the resource.
- Destroy deletes the key through the OpenAI Admin API.
