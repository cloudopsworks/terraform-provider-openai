# openai_project

API group: [Projects](../api-groups/projects.md).

Looks up one OpenAI organization project through the Administration API.

Use `id` for a direct project retrieve lookup.

## Example Usage

```terraform
data "openai_project" "app" {
  id = "proj_123"
}
```

## Schema

### Required

- `id` (String) OpenAI project ID. Lookup uses the project retrieve endpoint.

### Read-Only

- `name` (String) Project display name.
- `external_key_id` (String) External key ID associated with the project, when present.
- `status` (String) Project status, such as `active` or `archived`.
- `created_at` (Number) Unix timestamp for project creation.
- `archived_at` (Number) Unix timestamp for project archival when archived.
